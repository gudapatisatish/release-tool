package git

import (
    "bytes"
    "fmt"
    "os/exec"
    "regexp"
    "strings"
)

type Commit struct {
    Type    string
    Scope   string
    Message string
    Breaking bool
}

// GetCommitsSinceLastTag fetches commits since the last git tag
func GetCommitsSinceLastTag() ([]Commit, string, error) {
    // Get last tag
    cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
    lastTagBytes, _ := cmd.Output()
    lastTag := strings.TrimSpace(string(lastTagBytes))

    if lastTag == "" {
        lastTag = "0.0.0"
    }

    // Get commits since last tag
    logCmd := exec.Command("git", "log", fmt.Sprintf("%s..HEAD", lastTag), "--pretty=format:%s")
    var out bytes.Buffer
    logCmd.Stdout = &out
    err := logCmd.Run()
    if err != nil {
        return nil, lastTag, err
    }

    lines := strings.Split(out.String(), "\n")
    commits := []Commit{}
    re := regexp.MustCompile(`^(?P<type>\w+)(?:\((?P<scope>[^\)]+)\))?: (?P<msg>.+)`)

    for _, line := range lines {
        match := re.FindStringSubmatch(line)
        if match == nil {
            continue
        }
        commitType := match[1]
        scope := match[2]
        msg := match[3]
        breaking := strings.Contains(line, "BREAKING CHANGE")

        commits = append(commits, Commit{Type: commitType, Scope: scope, Message: msg, Breaking: breaking})
    }

    return commits, lastTag, nil
}

// CreateTag creates and pushes a git tag
func CreateTag(version string) error {
    tagCmd := exec.Command("git", "tag", version)
    if err := tagCmd.Run(); err != nil {
        return err
    }

    pushCmd := exec.Command("git", "push", "origin", version)
    return pushCmd.Run()
}
