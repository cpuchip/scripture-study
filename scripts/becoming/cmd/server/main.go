package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/cpuchip/scripture-study/scripts/becoming/internal/api"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/auth"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/db"
	"github.com/cpuchip/scripture-study/scripts/becoming/internal/envload"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

//go:embed dist/*
var frontendFS embed.FS

func main() {
	// Load .env first (before parsing flags, so env vars are available)
	envload.Load(".env")

	addr := flag.String("addr", envOrDefault("BECOMING_PORT", ":8080"), "listen address")
	dbPath := flag.String("db", envOrDefault("BECOMING_DB", "becoming.db"), "database path or PostgreSQL connection string")
	scriptures := flag.String("scriptures", envOrDefault("BECOMING_SCRIPTURES", "../../gospel-library/eng/scriptures"), "path to scriptures directory")
	dev := flag.Bool("dev", false, "development mode (CORS allow-all, skip auth)")
	tlsCert := flag.String("tls-cert", envOrDefault("TLS_CERT", ""), "TLS certificate file (enables HTTPS)")
	tlsKey := flag.String("tls-key", envOrDefault("TLS_KEY", ""), "TLS private key file")
	flag.Parse()

	useTLS := *tlsCert != "" && *tlsKey != ""

	// Normalize addr — if it's just a port number, add the colon
	if len(*addr) > 0 && (*addr)[0] != ':' {
		*addr = ":" + *addr
	}

	// Open database
	database, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()
	log.Printf("Database: %s", database.Path())

	// In dev mode, ensure the default user exists
	if *dev {
		if _, err := database.EnsureDefaultUser(); err != nil {
			log.Printf("Warning: could not ensure default user: %v", err)
		}
	}

	// Auth handlers
	oauthConfig := auth.OAuthConfigFromEnv()
	authHandlers := &auth.Handlers{
		DB:      database,
		DevMode: *dev,
		Secure:  !*dev || useTLS, // Secure cookies in production or with TLS
		OAuth:   oauthConfig,
	}
	if oauthConfig != nil {
		log.Printf("Google OAuth enabled (redirect: %s)", oauthConfig.RedirectURL)
	}

	// Build router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	if *dev {
		origins := []string{"http://localhost:5173", "http://localhost:8080"}
		if useTLS {
			origins = append(origins, "https://localhost:5173", "https://localhost"+*addr)
		}
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   origins,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
		}))
	}

	// Public auth routes (no authentication required)
	r.Post("/auth/register", authHandlers.Register)
	r.Post("/auth/login", authHandlers.Login)
	r.Post("/auth/logout", authHandlers.Logout)
	r.Get("/auth/google/login", authHandlers.GoogleLogin)
	r.Get("/auth/google/callback", authHandlers.GoogleCallback)
	r.Get("/api/auth/providers", authHandlers.Providers)

	// Protected API routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Required(database, *dev))

		// User profile
		r.Get("/api/me", authHandlers.Me)
		r.Put("/api/me", authHandlers.UpdateMe)
		r.Put("/api/me/password", authHandlers.ChangePassword)
		r.Delete("/api/me", authHandlers.DeleteAccount)
		r.Delete("/api/me/google", authHandlers.UnlinkGoogle)

		// Sessions
		r.Get("/api/sessions", authHandlers.ListSessions)
		r.Delete("/api/sessions", authHandlers.RevokeOtherSessions)
		r.Delete("/api/sessions/{id}", authHandlers.RevokeSession)

		// API tokens
		r.Get("/api/tokens", authHandlers.ListTokens)
		r.Post("/api/tokens", authHandlers.CreateToken)
		r.Delete("/api/tokens/{id}", authHandlers.DeleteToken)

		// Data export
		r.Get("/api/export", authHandlers.ExportData)

		// All existing API routes
		r.Mount("/api", api.Router(database, *scriptures))
	})

	// Serve frontend (embedded in production, Vite dev server in dev)
	if !*dev {
		distFS, err := fs.Sub(frontendFS, "dist")
		if err != nil {
			log.Fatalf("Failed to get frontend FS: %v", err)
		}
		fileServer := http.FileServer(http.FS(distFS))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file; if not found, serve index.html (SPA routing)
			path := r.URL.Path
			f, err := distFS.Open(path[1:]) // strip leading /
			if err != nil {
				// Serve index.html for SPA routes
				r.URL.Path = "/"
			} else {
				f.Close()
			}
			fileServer.ServeHTTP(w, r)
		})
	}

	log.Printf("Becoming server listening on %s", *addr)
	if *dev {
		log.Printf("Dev mode: API only (auto-login as user 1), frontend at http://localhost:5173")
	}

	if useTLS {
		fmt.Printf("\n  → https://localhost%s\n\n", *addr)
		if err := http.ListenAndServeTLS(*addr, *tlsCert, *tlsKey, r); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		fmt.Printf("\n  → http://localhost%s\n\n", *addr)
		if err := http.ListenAndServe(*addr, r); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

// envOrDefault returns the environment variable value, or the default if empty.
func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
