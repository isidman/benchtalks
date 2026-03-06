package main

import (
	"log"
	"net/http"
	"time"

	"github.com/isidman/benchtalks/pkg/config"
	natspkg "github.com/isidman/benchtalks/pkg/nats"
	"github.com/isidman/benchtalks/pkg/public"
	"github.com/isidman/benchtalks/pkg/server"
)

func main() {
	cfg := config.Load()
	hub := server.NewHub()

	// If NATS_PEERS is configured, then connect to the park and wire up the
	// relay into the hub. If not, "hub.relay" stays nil and everything runs as
	// standalone
	if len(cfg.NATSPeers) > 0 {
		relay, err := natspkg.Connect(cfg.NATSPeers, cfg.BenchID, hub.BroadcastFromPark, hub.HandlePairClaim, hub.HandlePairApproved)
		if err != nil {
			// fatal fail since not connecting to the park.
			// if an operator has configured peers, they expect federation to
			// work. A silent failure would be WAY worse than a crash in that
			// context.
			// That's why this exists
			log.Fatalf("[nats] failed to connect to park: %v", err)
		}
		hub.SetRelay(relay)

		// close the relay, when the server shuts down.
		// Practice-wise, this runs on a clean exit, and it lets NATS know this
		// bench was taken by the tiny giant from the park.
		defer relay.Close()
	}

	router := server.NewRouter(hub, public.StaticFiles, time.Now())

	log.Printf("Benchtalks is listening on: %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatal("server error:", err)
	}
}
