package targets

import (
    "fmt"
)

func GenerateChangelog(version string) error {
    fmt.Printf("Generating CHANGELOG for version %s...\n", version)
    // TODO: parse git commits and write CHANGELOG.md
    return nil
}
