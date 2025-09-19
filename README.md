# Phase 0: Release Tool - Version Calculator

A minimal implementation focusing **only** on version calculation and commit analysis.

## What Phase 0 Does

- **Version Calculation**: Analyzes commits and calculates next semantic version  
- **Commit Parsing**: Parses conventional commits (`feat:`, `fix:`, `BREAKING CHANGE`)  
- **Detailed Analysis**: Shows breakdown of changes (features, fixes, breaking changes)  
- **JSON Output**: Machine-readable output for CI integration  
- **Error Handling**: Graceful handling of malformed commits and git errors  
- **Dry Run**: No side effects - only analysis and calculation  

## What Phase 0 Doesn't Do

- No Git operations (no tagging, no pushing)  
- No package publishing (no Python, Docker, etc.)  
- No config files (pure CLI arguments)  
- No state management  
- No network operations  

## Usage

```bash
# Basic analysis
./release-tool

# Detailed commit breakdown
./release-tool --verbose

# JSON output for CI scripts
./release-tool --json | jq '.next_version'

# Analyze different repository
./release-tool --repo /path/to/other/repo

# Strict mode (treat unknown commits as patch bumps)
./release-tool --strict
```

## Example Output

```
📊 Version Analysis
==================

Current Version: 1.2.0
Next Version:    1.3.0
Bump Type:       minor
Commits Since:   4

📈 Change Summary
-----------------
Breaking Changes: 0
Features:         2
Fixes:           1
Other:           1

🚀 Next Steps
-------------
To release version 1.3.0:
1. Review the changes above
2. Run: git tag 1.3.0
3. Run: git push origin 1.3.0
```

## CI Integration

```yaml
# .gitlab-ci.yml
calculate_version:
  script:
    - go build -o release-tool main.go
    - NEXT_VERSION=$(./release-tool --json | jq -r '.next_version')
    - echo "Next version will be $NEXT_VERSION"
    - echo "NEXT_VERSION=$NEXT_VERSION" >> version.env
  artifacts:
    reports:
      dotenv: version.env
```

## Build & Test

```bash
# Build
go build -o release-tool main.go

# Test with sample repo
chmod +x test.sh
./test.sh
```

## Commit Message Format

Follows [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types that trigger version bumps:
- `feat:` → Minor version bump
- `fix:` → Patch version bump  
- `BREAKING CHANGE:` or `!` → Major version bump

### Other types (no version bump):
- `docs:`, `style:`, `refactor:`, `test:`, `chore:`

## Error Handling

- **No tags**: Starts from `0.0.0`
- **No commits**: Reports no changes needed
- **Malformed commits**: Reports count, optional strict mode
- **Git errors**: Clear error messages with context
- **Invalid versions**: Helpful parsing error messages

## Next Phase Preview

Phase 1 will add:
- Actual Git tagging operations
- Python package publishing (`pyproject.toml` updates)
- Configuration file support (`release.yml`)
- State management for rollback scenarios
- Pre-release versioning with timestamps
