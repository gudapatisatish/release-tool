package orchestrator

import (
    "fmt"
    "os"
    "text/template"

    "release-tool/internal/changelog"
    "release-tool/internal/config"
    "release-tool/internal/git"
    "release-tool/internal/targets"
    "release-tool/internal/version"
)

func Run(preRelease, dryRun bool) error {
    cfg, _ := config.LoadConfig()

    commits, lastTag, _ := git.GetCommitsSinceLastTag()
    nextVersion, err := version.CalculateNextVersion(preRelease, os.Getenv("CI_PIPELINE_ID"))
    if err != nil {
        return err
    }

    fmt.Printf("Calculated next version: %s (previous %s)\n", nextVersion, lastTag)

    changelogFile := "CHANGELOG.md"
    if cfg != nil && cfg.Versioning.Changelog != "" {
        changelogFile = cfg.Versioning.Changelog
    }
    if err := changelog.GenerateToFile(changelogFile, nextVersion, commits); err != nil {
        return err
    }

    var releaseTargets []targets.ReleaseTarget
    if cfg != nil {
        for _, t := range cfg.Targets {
            switch t.Type {
            case "python":
                releaseTargets = append(releaseTargets, &targets.PythonTarget{
                    Pyproject: t.PythonTarget.Pyproject,
                    Repository: t.PythonTarget.Repository,
                    Username: t.PythonTarget.Username,
                    Password: t.PythonTarget.Password,
                })
            case "docker":
                releaseTargets = append(releaseTargets, &targets.DockerTarget{
                    Dockerfile: t.DockerTarget.Dockerfile,
                    Image:      t.DockerTarget.Image,
                    Tags:       t.DockerTarget.Tags,
                })
            }
        }
    } else {
        fmt.Println("No config file detected → auto-detect mode")
        if targets.FileExists("pyproject.toml") {
            releaseTargets = append(releaseTargets, &targets.PythonTarget{Pyproject: "pyproject.toml"})
        }
        if targets.FileExists("Dockerfile") {
            releaseTargets = append(releaseTargets, &targets.DockerTarget{Dockerfile: "Dockerfile"})
        }
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
        tagFormat := "{{ .Version }}"
        if cfg != nil && cfg.Versioning.TagFormat != "" {
            tagFormat = cfg.Versioning.TagFormat
        }

        tmpl, _ := template.New("tag").Parse(tagFormat)
        var buf string
        _ = tmpl.Execute(&buf, map[string]string{"Version": nextVersion})
        return git.CreateTag(buf)
    }

    fmt.Println("Dry-run complete.")
    return nil
}
