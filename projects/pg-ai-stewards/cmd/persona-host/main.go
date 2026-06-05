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
	"flag"
	"fmt"
	"log"
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

	// The HTTP surface (pubkey, join) lands in PS.2+. For PS.1 the process
	// proves the boot+migrate path and holds until signaled.
	log.Printf("persona-host %s up (addr=%s reserved for PS.2+); awaiting signal", version, *addr)
	<-ctx.Done()
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
