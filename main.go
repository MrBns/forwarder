package main

import (
	"log"

	"github.com/MrBns/forwarder/internal/app"
)

// main is intentionally minimal: dependency wiring is handled by the
// Wire-generated initializeApp() in wire_gen.go.
func main() {
	a, err := initializeApp()
	if err != nil {
		log.Fatalf("failed to initialise app: %v", err)
	}
	if err := a.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// Ensure the app package is used (import kept for godoc / IDE navigation).
var _ *app.App
