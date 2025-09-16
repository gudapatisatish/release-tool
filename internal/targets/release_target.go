package targets

type ReleaseTarget interface {
    Name() string
    Build(version string) error
    Publish(version string) error
}

// Utility function to detect files
import "os"

func FileExists(path string) bool {
    if _, err := os.Stat(path); err == nil {
        return true
    }
    return false
}
