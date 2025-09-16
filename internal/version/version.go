package version

import (
    "fmt"
)

func CalculateNextVersion(preRelease bool) (string, error) {
    // TODO: implement semantic version calculation from Git commits
    baseVersion := "1.0.0"
    if preRelease {
        return fmt.Sprintf("%s-beta.1", baseVersion), nil
    }
    return baseVersion, nil
}
