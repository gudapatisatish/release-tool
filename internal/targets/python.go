package targets

import (
    "fmt"
    "os/exec"
)

type PythonTarget struct{}

func (p *PythonTarget) Name() string { return "Python Package" }

func (p *PythonTarget) Build(version string) error {
    fmt.Println("Updating pyproject.toml version...")
    if err := p.UpdatePyprojectVersion(version); err != nil {
        return err
    }
    fmt.Println("Building Python package...")
    cmd := exec.Command("python", "-m", "build", "--sdist", "--wheel", ".")
    return cmd.Run()
}

func (p *PythonTarget) Publish(version string) error {
    fmt.Println("Uploading Python package...")
    cmd := exec.Command("twine", "upload", "dist/*")
    cmd.Stdout = nil
    cmd.Stderr = nil
    return cmd.Run()
}

// internal/targets/python.go (update)
func (p *PythonTarget) UpdatePyprojectVersion(version string) error {
    content, err := os.ReadFile("pyproject.toml")
    if err != nil {
        return err
    }
    updated := regexp.MustCompile(`version = ".*"`).
        ReplaceAllString(string(content), fmt.Sprintf(`version = "%s"`, version))

    return os.WriteFile("pyproject.toml", []byte(updated), 0644)
}
