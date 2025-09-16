package main

import (
    "fmt"
    "os"

    "release-tool/internal/orchestrator"
)

func main() {
    preRelease := false
    dryRun := false

    // Parse CLI args (simplified)
    for i, arg := range os.Args {
        switch arg {
        case "--pre-release":
            preRelease = true
        case "--dry-run":
            dryRun = true
        }
    }

    fmt.Println("Starting release tool...")
    err := orchestrator.Run(preRelease, dryRun)
    if err != nil {
        fmt.Printf("Release failed: %v\n", err)
        os.Exit(1)
    }
}
