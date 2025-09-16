package version

import (
    "fmt"
    "strconv"
    "strings"

    "release-tool/internal/git"
)

func bumpVersion(base string, commits []git.Commit, preRelease bool, pipelineID string) (string, error) {
    parts := strings.Split(base, ".")
    if len(parts) != 3 {
        return "", fmt.Errorf("invalid base version: %s", base)
    }

    major, _ := strconv.Atoi(parts[0])
    minor, _ := strconv.Atoi(parts[1])
    patch, _ := strconv.Atoi(parts[2])

    bumpType := "patch"

    for _, c := range commits {
        if c.Breaking {
            bumpType = "major"
            break
        }
        if c.Type == "feat" && bumpType != "major" {
            bumpType = "minor"
        }
    }

    switch bumpType {
    case "major":
        major++
        minor, patch = 0, 0
    case "minor":
        minor++
        patch = 0
    case "patch":
        patch++
    }

    newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)

    if preRelease {
        if pipelineID == "" {
            newVersion += "-beta"
        } else {
            newVersion += fmt.Sprintf("-beta.%s", pipelineID)
        }
    }

    return newVersion, nil
}

func CalculateNextVersion(preRelease bool, pipelineID string) (string, error) {
    commits, lastTag, err := git.GetCommitsSinceLastTag()
    if err != nil {
        return "", err
    }

    if len(commits) == 0 {
        return "", fmt.Errorf("no new commits since last tag")
    }

    return bumpVersion(lastTag, commits, preRelease, pipelineID)
}
