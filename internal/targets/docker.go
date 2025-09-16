package targets

import (
    "fmt"
    "os/exec"
)

type DockerTarget struct{}

func (d *DockerTarget) Name() string { return "Docker Image" }

func (d *DockerTarget) Build(version string) error {
    fmt.Printf("Building Docker image: myapp:%s\n", version)
    cmd := exec.Command("docker", "build", "-t", "myapp:"+version, ".")
    cmd.Stdout = nil
    cmd.Stderr = nil
    return cmd.Run()
}

func (d *DockerTarget) Publish(version string) error {
    fmt.Printf("Pushing Docker image: myapp:%s\n", version)
    cmd := exec.Command("docker", "push", "myapp:"+version)
    cmd.Stdout = nil
    cmd.Stderr = nil
    return cmd.Run()
}
