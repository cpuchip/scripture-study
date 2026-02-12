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
	dbPath := flag.String("db", envOrDefault("BECOMING_DB", "becoming.db"), "SQLite database path")
	scriptures := flag.String("scriptures", "../../gospel-library/eng/scriptures", "path to scriptures directory")
	dev := flag.Bool("dev", false, "development mode (CORS allow-all, skip auth)")
	flag.Parse()

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
	authHandlers := &auth.Handlers{
		DB:      database,
		DevMode: *dev,
		Secure:  !*dev, // Secure cookies only in production (HTTPS)
	}

	// Build router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	if *dev {
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:8080"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization"},
			AllowCredentials: true,
		}))
	}

	// Public auth routes (no authentication required)
	r.Post("/auth/register", authHandlers.Register)
	r.Post("/auth/login", authHandlers.Login)
	r.Post("/auth/logout", authHandlers.Logout)
	r.Get("/api/auth/providers", authHandlers.Providers)

	// Protected API routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Required(database, *dev))

		// User profile
		r.Get("/api/me", authHandlers.Me)
		r.Put("/api/me", authHandlers.UpdateMe)

		// API tokens
		r.Get("/api/tokens", authHandlers.ListTokens)
		r.Post("/api/tokens", authHandlers.CreateToken)
		r.Delete("/api/tokens/{id}", authHandlers.DeleteToken)

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
	fmt.Printf("\n  → http://localhost%s\n\n", *addr)

	if err := http.ListenAndServe(*addr, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// envOrDefault returns the environment variable value, or the default if empty.
func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
