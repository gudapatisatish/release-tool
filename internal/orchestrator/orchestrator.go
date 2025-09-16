package orchestrator

import (
    "fmt"

    "release-tool/internal/targets"
    "release-tool/internal/version"
)

func Run(targetArgs []string, preRelease bool, dryRun bool) error {
    fmt.Println("Detecting release targets...")
    var releaseTargets []targets.ReleaseTarget

    if len(targetArgs) > 0 {
        // Manual override
        for _, t := range targetArgs {
            switch t {
            case "python":
                releaseTargets = append(releaseTargets, &targets.PythonTarget{})
            case "docker":
                releaseTargets = append(releaseTargets, &targets.DockerTarget{})
            default:
                fmt.Printf("Unknown target: %s\n", t)
            }
        }
    } else {
        // Auto-detect
        if targets.FileExists("pyproject.toml") {
            releaseTargets = append(releaseTargets, &targets.PythonTarget{})
        }
        if targets.FileExists("Dockerfile") {
            releaseTargets = append(releaseTargets, &targets.DockerTarget{})
        }
    }

    if len(releaseTargets) == 0 {
        return fmt.Errorf("no release targets detected")
    }

    // Version management
    nextVersion, err := version.CalculateNextVersion(preRelease)
    if err != nil {
        return err
    }
    fmt.Printf("Next version: %s\n", nextVersion)

    // Generate changelog
    err = targets.GenerateChangelog(nextVersion)
    if err != nil {
        return err
    }

    // Build & Publish
    for _, t := range releaseTargets {
        fmt.Printf("Building target: %s\n", t.Name())
        if err := t.Build(nextVersion); err != nil {
            return err
        }
        if !dryRun {
            fmt.Printf("Publishing target: %s\n", t.Name())
            if err := t.Publish(nextVersion); err != nil {
                return err
            }
        } else {
            fmt.Printf("Dry-run: skipping publish for %s\n", t.Name())
        }
    }

    fmt.Println("Release completed successfully.")
    return nil
}
