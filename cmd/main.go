package main

import (
    "fmt"
    "os"
    "strings"

    "release-tool/internal/orchestrator"
)

func main() {
    targets := []string{}
    preRelease := false
    dryRun := false

    // Parse CLI args (simplified)
    for i, arg := range os.Args {
        switch arg {
        case "--targets":
            if i+1 < len(os.Args) {
                targets = strings.Split(os.Args[i+1], ",")
            }
        case "--pre-release":
            preRelease = true
        case "--dry-run":
            dryRun = true
        }
    }

    fmt.Println("Starting release tool...")
    err := orchestrator.Run(targets, preRelease, dryRun)
    if err != nil {
        fmt.Printf("Release failed: %v\n", err)
        os.Exit(1)
    }
}
