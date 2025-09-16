# release-tool
Go-based release orchestrator that interacts with Python, Docker

# **Local Development Guide**

## **1. Prerequisites**

Install these tools on your machine:

| Tool                      | Purpose                         | Minimum Version |
| ------------------------- | ------------------------------- | --------------- |
| Go                        | CLI development                 | 1.22            |
| Git                       | Source control                  | Latest          |
| Python                    | For Python package testing      | 3.10            |
| Poetry or `build`/`twine` | Python packaging                | Latest          |
| Docker                    | Docker image builds             | Latest          |
| `direnv` (optional)       | Environment variable management | Latest          |

**Verify installs:**

```bash
go version
git --version
python3 --version
docker --version
```

---

## **2. Clone the Repository**

```bash
mkdir -p ~/dev/release-tool
cd ~/dev/release-tool
git clone <your-repo-url> .
```

---

## **3. Set Up Go Modules**

```bash
go mod tidy
go install ./...
```

* This ensures all dependencies are installed.
* Running `go install ./...` builds the CLI locally.

---

## **4. Configure Environment Variables**

Create a `.env` or use `direnv` to manage credentials safely:

```bash
export ARTIFACTORY_USER=<your-user>
export ARTIFACTORY_PW=<your-password>
export PYPI_TOKEN=<your-pypi-token>
export CI_PIPELINE_ID=local
export CI_COMMIT_REF_NAME=dev
export CI_DEFAULT_BRANCH=main
```

* This allows local dry-runs and builds without touching production.

---

## **5. Set Up Example Repositories**

Create local example repos to simulate different release targets:

```text
examples/
├── python-only/
│   ├── pyproject.toml
│   └── src/example/
├── docker-only/
│   ├── Dockerfile
│   └── app/
└── multi-target/
    ├── pyproject.toml
    ├── Dockerfile
    └── src/example/
```

* Commit some conventional commits for testing:

  ```bash
  git commit -m "feat: add authentication module"
  git commit -m "fix: resolve login crash"
  ```

---

## **6. CLI Commands for Local Testing**

### **Dry Run**

Simulate a release without making changes:

```bash
./release-tool dry-run --repo examples/python-only
./release-tool dry-run --repo examples/multi-target
```

* Validates commit parsing, version calculation, changelog generation, and target detection.

### **Release**

Trigger an actual release locally (test only in a sandbox repo):

```bash
./release-tool release --repo examples/multi-target
```

* Auto-detects targets (`Python`, `Docker`).
* Builds artifacts and updates changelogs.
* Pushes tags if `GITLAB_TOKEN` or equivalent is configured.

### **Manual Target Override**

```bash
./release-tool release --repo examples/multi-target --targets=python
./release-tool release --repo examples/multi-target --skip=docker
```

---

## **7. Version Management Testing**

* Run `./release-tool version` to see the calculated next version.
* Supports pre-release with:

```bash
./release-tool release --pre-release
```

* Sanitized versions will be automatically applied to `pyproject.toml` and Docker tags.

---

## **8. Changelog Generation Testing**

* Run:

```bash
./release-tool changelog --repo examples/multi-target
```

* Generates `CHANGELOG.md` automatically, grouped by commit type and section.

---

## **9. Testing Python Package Release Locally**

```bash
cd examples/python-only
./release-tool release --targets=python --dry-run
python -m build --sdist --wheel .
twine check dist/*
```

* Confirms the orchestrator correctly updates version and builds artifacts.

---

## **10. Testing Docker Release Locally**

```bash
cd examples/docker-only
./release-tool release --targets=docker --dry-run
docker build -t test-image:0.1.0 .
docker tag test-image:0.1.0 myregistry.local/test-image:0.1.0
```

* Confirms Docker plugin works and tags are applied correctly.

---

## **11. Documenting While Developing**

**Folder structure:**

```text
docs/
├── README.md             # Overview of release tool
├── development.md        # Local dev setup, CLI commands
├── release.md            # How to run releases
├── targets.md            # Plugin-specific docs (Python, Docker)
├── contributing.md       # Commit, PR, and testing guidelines
examples/                  # Sample repos for local dev
```

* Update `docs/development.md` as you add features.
* Keep CLI `--help` updated for all commands and flags.

---

## **12. Suggested Local Workflow**

1. Create a new branch for feature/bug:

   ```bash
   git checkout -b feature/git-parser
   ```
2. Implement feature in `internal/` or `targets/`
3. Run unit tests:

   ```bash
   go test ./...
   ```
4. Test in example repos:

   ```bash
   ./release-tool dry-run --repo examples/python-only
   ```
5. Update documentation in `docs/`
6. Push branch → CI runs → review
7. Merge to `main` when validated
