package orchestrator

import (
    "fmt"

    "release-tool/internal/targets"
    "release-tool/internal/version"
)

func Run(targetArgs []string, preRelease bool, dryRun bool) error {
    fmt.Println("Detecting release targets...")
    var releaseTargets []targets.ReleaseTarget

    if len(targetArgs) == 0 {
        if targets.FileExists("pyproject.toml") {
            releaseTargets = append(releaseTargets, &targets.PythonTarget{})
        }
        if targets.FileExists("Dockerfile") {
            releaseTargets = append(releaseTargets, &targets.DockerTarget{})
        }
    }

    commits, lastTag, _ := git.GetCommitsSinceLastTag()
    nextVersion, err := version.CalculateNextVersion(preRelease, os.Getenv("CI_PIPELINE_ID"))
    if err != nil {
        return err
    }

    fmt.Printf("Releasing version %s (previous %s)\n", nextVersion, lastTag)

    err = changelog.Generate(nextVersion, commits)
    if err != nil {
        return err
    }

    for _, t := range releaseTargets {
        fmt.Printf("Building %s...\n", t.Name())
        if err := t.Build(nextVersion); err != nil {
            return err
        }
        if !dryRun {
            if err := t.Publish(nextVersion); err != nil {
                return err
            }
        }
    }

    if !dryRun {
        return git.CreateTag(nextVersion)
    }

    fmt.Println("Dry-run complete.")
    return nil
}
