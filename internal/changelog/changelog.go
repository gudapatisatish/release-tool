package changelog

import (
    "fmt"
    "os"
    "strings"

    "release-tool/internal/git"
)

func Generate(version string, commits []git.Commit) error {
    file, err := os.OpenFile("CHANGELOG.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer file.Close()

    sections := map[string][]string{}

    for _, c := range commits {
        switch c.Type {
        case "feat":
            sections["✨ Features ✨"] = append(sections["✨ Features ✨"], c.Message)
        case "fix":
            sections["🐛 Bug Fixes 🐛"] = append(sections["🐛 Bug Fixes 🐛"], c.Message)
        case "perf":
            sections["⚡ Performance ⚡"] = append(sections["⚡ Performance ⚡"], c.Message)
        case "refactor":
            sections["🔨 Refactors 🔨"] = append(sections["🔨 Refactors 🔨"], c.Message)
        }
    }

    _, _ = file.WriteString(fmt.Sprintf("\n## %s\n", version))
    for title, msgs := range sections {
        _, _ = file.WriteString(fmt.Sprintf("### %s\n", title))
        for _, msg := range msgs {
            _, _ = file.WriteString(fmt.Sprintf("- %s\n", msg))
        }
    }

    return nil
}
