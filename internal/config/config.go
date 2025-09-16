// Example Config File

// versioning:
//   scheme: semver
//   pre_release: true
//   tag_format: "v{{ .Version }}" # Go template
//   changelog: "CHANGELOG.md"

// targets:
//   - type: python
//     pyproject: pyproject.toml
//     publish: true
//     repository: https://artifactory.example.com/api/pypi/pypi-local
//     username: ${ARTIFACTORY_USER}
//     password: ${ARTIFACTORY_PW}

//   - type: docker
//     dockerfile: Dockerfile
//     image: registry.example.com/my-team/my-app
//     tags:
//       - "{{ .Version }}"
//       - "latest"

// git:
//   push_tags: true
//   commit_message: "chore(release): {{ .Version }}"


package config

import (
    "os"

    "gopkg.in/yaml.v3"
)

type Versioning struct {
    Scheme     string `yaml:"scheme"`
    PreRelease bool   `yaml:"pre_release"`
    TagFormat  string `yaml:"tag_format"`
    Changelog  string `yaml:"changelog"`
}

type PythonTarget struct {
    Type       string `yaml:"type"`
    Pyproject  string `yaml:"pyproject"`
    Publish    bool   `yaml:"publish"`
    Repository string `yaml:"repository"`
    Username   string `yaml:"username"`
    Password   string `yaml:"password"`
}

type DockerTarget struct {
    Type      string   `yaml:"type"`
    Dockerfile string   `yaml:"dockerfile"`
    Image     string   `yaml:"image"`
    Tags      []string `yaml:"tags"`
}

type Target struct {
    Type   string `yaml:"type"`
    PythonTarget
    DockerTarget
}

type GitConfig struct {
    PushTags      bool   `yaml:"push_tags"`
    CommitMessage string `yaml:"commit_message"`
}

type ReleaseConfig struct {
    Versioning Versioning `yaml:"versioning"`
    Targets    []Target   `yaml:"targets"`
    Git        GitConfig  `yaml:"git"`
}

func LoadConfig() (*ReleaseConfig, error) {
    files := []string{"release.yml", ".release.yml"}
    for _, f := range files {
        if _, err := os.Stat(f); err == nil {
            data, err := os.ReadFile(f)
            if err != nil {
                return nil, err
            }
            var cfg ReleaseConfig
            if err := yaml.Unmarshal(data, &cfg); err != nil {
                return nil, err
            }
            return &cfg, nil
        }
    }
    return nil, nil // no config, auto-detect mode
}
