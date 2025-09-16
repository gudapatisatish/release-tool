package targets

import (
    "fmt"
    "os/exec"
)

type PythonTarget struct{}

func (p *PythonTarget) Name() string { return "Python Package" }

func (p *PythonTarget) Build(version string) error {
    fmt.Println("Building Python package...")
    cmd := exec.Command("python", "-m", "build", "--sdist", "--wheel", ".")
    cmd.Stdout = nil
    cmd.Stderr = nil
    return cmd.Run()
}

func (p *PythonTarget) Publish(version string) error {
    fmt.Println("Uploading Python package...")
    cmd := exec.Command("twine", "upload", "dist/*")
    cmd.Stdout = nil
    cmd.Stderr = nil
    return cmd.Run()
}
