// Pinewood derby scoring server + CLI.
package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cpuchip/pinewood/internal/api"
	"github.com/cpuchip/pinewood/internal/audit"
	"github.com/cpuchip/pinewood/internal/db"
	"github.com/cpuchip/pinewood/internal/schedule"
	"github.com/cpuchip/pinewood/internal/ws"
)

//go:embed all:dist
var distFS embed.FS

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "schedule":
			cliSchedule(os.Args[2:])
			return
		case "serve", "":
			os.Args = append(os.Args[:1], os.Args[2:]...)
		}
	}

	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "derby.db", "SQLite database path")
	logPath := flag.String("log", "derby.log", "JSONL audit log path")
	flag.Parse()

	d, err := db.Open(*dbPath)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer d.Close()

	al, err := audit.Open(*logPath)
	if err != nil {
		log.Fatalf("audit open: %v", err)
	}
	defer al.Close()

	hub := ws.New()

	// Embedded SPA via dist/.
	var spa fs.FS
	if sub, err := fs.Sub(distFS, "dist"); err == nil {
		spa = sub
	}

	srv := &api.Server{DB: d, Audit: al, Hub: hub, Static: spa}

	httpSrv := &http.Server{Addr: *addr, Handler: srv.Routes()}
	go func() {
		log.Printf("pinewood listening on %s (db=%s log=%s)", *addr, *dbPath, *logPath)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	httpSrv.Shutdown(ctx)
}

func cliSchedule(args []string) {
	fs := flag.NewFlagSet("schedule", flag.ExitOnError)
	cars := fs.Int("cars", 25, "number of cars")
	runs := fs.Int("runs", 6, "runs per car")
	lanes := fs.Int("lanes", 3, "lane count")
	verify := fs.Bool("verify", false, "print fairness stats")
	fs.Parse(args)

	carNums := make([]int, *cars)
	for i := range carNums {
		carNums[i] = i + 1
	}
	ch, err := schedule.Generate(carNums, schedule.Options{
		Lanes: *lanes, RunsPerCar: *runs, MinGap: 1, Seed: 42,
	})
	if err != nil {
		log.Fatalf("generate: %v", err)
	}
	fmt.Printf("Heat\tLane 1\tLane 2\tLane 3\n")
	for i, h := range ch.Heats {
		fmt.Printf("%d", i+1)
		for _, c := range h {
			fmt.Printf("\t%d", c)
		}
		fmt.Println()
	}
	if *verify {
		st := schedule.Analyze(ch)
		fmt.Println()
		fmt.Printf("Total heats:    %d\n", st.TotalHeats)
		fmt.Printf("Cars:           %d\n", len(st.RunsPerCar))
		runsOK := true
		for _, r := range st.RunsPerCar {
			if r != *runs {
				runsOK = false
			}
		}
		fmt.Printf("All cars run %d times: %v\n", *runs, runsOK)
		laneOK := true
		expectPerLane := *runs / *lanes
		for _, lc := range st.LaneCounts {
			for _, c := range lc {
				if c != expectPerLane {
					laneOK = false
				}
			}
		}
		fmt.Printf("All cars %d/lane: %v\n", expectPerLane, laneOK)
		fmt.Printf("Unique pairs:   %d\n", st.UniquePairs)
		fmt.Printf("Pair freq dist: %v\n", st.PairCountDist)
		fmt.Printf("Gap min/max/avg: %d / %d / %.2f\n", st.MinGap, st.MaxGap, st.AvgGap)
	}
}
