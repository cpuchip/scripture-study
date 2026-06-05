// persona-host is pg-ai-stewards' optional persona SIDECAR (the
// substrate-persona-concept proposal). It owns persona identity, EdDSA/JWT
// credential minting, and the ai-chattermax room handshake — kept OUT of the
// core extension so a general substrate install never runs it. State lives in a
// sidecar-managed `persona_host` schema in the substrate's Postgres. One host
// serves many personas.
//
// Security discipline: the Ed25519 private signing key (PS.2) is generated and
// stored in persona_host.signing_key and is NEVER logged, exported, or placed in
// any model context — the same handling class as the coder's GitHub token.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const version = "0.1.0"

func main() {
	var (
		smoke = flag.Bool("smoke", false, "Boot, apply the persona_host migration, verify the tables, and exit (PS.1 smoke).")
		dsn   = flag.String("dsn", "", "Postgres DSN (overrides STEWARDS_DSN).")
		addr  = flag.String("addr", ":8090", "HTTP listen address (used from PS.2 onward).")
	)
	flag.Parse()

	log.SetOutput(os.Stderr)
	log.SetPrefix("persona-host: ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	if *dsn == "" {
		*dsn = os.Getenv("STEWARDS_DSN")
	}
	if *dsn == "" {
		log.Fatalf("no DSN: set STEWARDS_DSN or pass -dsn")
	}

	if *smoke {
		if err := runSmoke(*dsn); err != nil {
			log.Fatalf("smoke FAILED: %v", err)
		}
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := OpenStore(ctx, *dsn)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Printf("persona_host schema ready")

	key, err := LoadOrCreateKey(ctx, store)
	if err != nil {
		log.Fatalf("signing key: %v", err)
	}
	// Log the public fingerprint only — never the private key.
	log.Printf("signing key ready (ed25519, fingerprint=%s)", key.Fingerprint())

	if err := SeedDefaultPersonas(ctx, store); err != nil {
		log.Fatalf("seed personas: %v", err)
	}
	log.Printf("default personas seeded")

	srv := NewServer(store, key, NewMinter(store, key))
	httpSrv := &http.Server{Addr: *addr, Handler: srv.Handler()}
	go func() {
		log.Printf("persona-host %s listening on %s", version, *addr)
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	log.Printf("persona-host stopped cleanly")
}

// runSmoke proves PS.1 end to end: connect → apply migration → confirm the four
// persona_host tables exist.
func runSmoke(dsn string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	store, err := OpenStore(ctx, dsn)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer store.Close()

	fmt.Println("persona-host smoke: applying persona_host schema migration…")
	if err := store.Migrate(ctx); err != nil {
		return err
	}

	tables, err := store.tableNames(ctx)
	if err != nil {
		return fmt.Errorf("list tables: %w", err)
	}
	fmt.Printf("persona-host smoke: persona_host tables present: %v\n", tables)

	want := []string{"persona_rooms", "personas", "signing_key", "token_issuance"}
	if !sameStringSet(tables, want) {
		return fmt.Errorf("table mismatch: want %v, got %v", want, tables)
	}

	// PS.2: signing key generates once and persists (idempotent across calls).
	k1, err := LoadOrCreateKey(ctx, store)
	if err != nil {
		return fmt.Errorf("load/create signing key: %w", err)
	}
	k2, err := LoadOrCreateKey(ctx, store)
	if err != nil {
		return fmt.Errorf("reload signing key: %w", err)
	}
	if !k1.Pub.Equal(k2.Pub) {
		return errors.New("signing key not stable across calls — it regenerated instead of persisting")
	}
	if _, perr := parsePublicPEM(k1.PublicPEM); perr != nil {
		return fmt.Errorf("published pubkey unparseable: %w", perr)
	}
	fmt.Printf("persona-host smoke: signing key stable + parseable (ed25519, fingerprint=%s)\n", k1.Fingerprint())

	// PS.3: mint a real DB-backed token and verify it round-trips; an
	// independent key must NOT verify it (inverse hypothesis).
	pid, err := store.UpsertPersona(ctx, Persona{Slug: "smoke-persona", DisplayName: "Smoke Persona", AgentFamily: "study"})
	if err != nil {
		return fmt.Errorf("upsert smoke persona: %w", err)
	}
	p, err := store.PersonaBySlug(ctx, "smoke-persona")
	if err != nil {
		return fmt.Errorf("load smoke persona: %w", err)
	}
	if p.ID != pid {
		return fmt.Errorf("persona id mismatch: upsert=%s select=%s", pid, p.ID)
	}
	tok, _, err := NewMinter(store, k1).MintToken(ctx, p, "smoke-room", 0)
	if err != nil {
		return fmt.Errorf("mint token: %w", err)
	}
	claims, err := VerifyToken(tok, k1.Pub)
	if err != nil {
		return fmt.Errorf("verify minted token: %w", err)
	}
	if claims.Subject != p.ID || claims.Room != "smoke-room" || claims.ID == "" {
		return fmt.Errorf("unexpected claims: sub=%s room=%s jti=%s", claims.Subject, claims.Room, claims.ID)
	}
	// Inverse hypothesis: an independent key must NOT verify this token.
	_, otherPubPEM, gerr := generateKeyPEM()
	if gerr != nil {
		return fmt.Errorf("generate inverse key: %w", gerr)
	}
	otherPub, perr := parsePublicPEM(otherPubPEM)
	if perr != nil {
		return fmt.Errorf("parse inverse key: %w", perr)
	}
	if _, verr := VerifyToken(tok, otherPub); verr == nil {
		return errors.New("a token verified against the WRONG key — signature check is broken")
	}
	// Never print tok itself — only the safe claim summary.
	fmt.Printf("persona-host smoke: minted+verified token (sub=%s room=%s jti=%s); wrong-key correctly rejected\n", claims.Subject, claims.Room, claims.ID)

	// PS.4: seed the default personas and confirm they're registered.
	if err := SeedDefaultPersonas(ctx, store); err != nil {
		return fmt.Errorf("seed personas: %w", err)
	}
	personas, err := store.ListPersonas(ctx)
	if err != nil {
		return fmt.Errorf("list personas: %w", err)
	}
	have := make(map[string]string, len(personas))
	for _, pr := range personas {
		have[pr.Slug] = pr.AgentFamily
	}
	for _, want := range []string{"dm-assistant", "npc-ally"} {
		if _, ok := have[want]; !ok {
			return fmt.Errorf("seeded persona %q missing from registry", want)
		}
	}
	fmt.Printf("persona-host smoke: personas registered: dm-assistant(%s), npc-ally(%s)\n", have["dm-assistant"], have["npc-ally"])

	// PS.5: room handshake — JoinRoom mints a verifiable token AND records
	// persona_rooms membership.
	joinSrv := NewServer(store, k1, NewMinter(store, k1))
	res, err := joinSrv.JoinRoom(ctx, "dm-assistant", "tavern-smoke")
	if err != nil {
		return fmt.Errorf("join room: %w", err)
	}
	jc, err := VerifyToken(res.Token, k1.Pub)
	if err != nil {
		return fmt.Errorf("join token verify: %w", err)
	}
	if jc.Room != "tavern-smoke" || jc.Slug != "dm-assistant" {
		return fmt.Errorf("join claims mismatch: room=%s slug=%s", jc.Room, jc.Slug)
	}
	joined, err := store.HasPersonaRoom(ctx, jc.Subject, "tavern-smoke")
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if !joined {
		return errors.New("persona_rooms row missing after join")
	}
	fmt.Printf("persona-host smoke: dm-assistant joined tavern-smoke (token verified, membership recorded)\n")

	fmt.Println("persona-host smoke: PASS")
	return nil
}

func sameStringSet(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[string]bool, len(got))
	for _, g := range got {
		seen[g] = true
	}
	for _, w := range want {
		if !seen[w] {
			return false
		}
	}
	return true
}
