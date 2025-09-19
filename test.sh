#!/bin/bash
# Phase 0 Test Script and Usage Examples

# Build the tool
echo "🔨 Building release-tool..."
go build -o release-tool main.go

# Test 1: Basic version calculation
echo "📊 Test 1: Basic version calculation"
./release-tool --repo . --verbose

echo -e "\n" && read -p "Press Enter to continue..."

# Test 2: JSON output for CI integration
echo "📊 Test 2: JSON output"
./release-tool --repo . --json | jq .

echo -e "\n" && read -p "Press Enter to continue..."

# Test 3: Strict mode (treat unknown commits as patches)
echo "📊 Test 3: Strict mode"
./release-tool --repo . --strict --verbose

echo -e "\n" && read -p "Press Enter to continue..."

# Test 4: Different repository
echo "📊 Test 4: Different repository"
echo "Testing with a different repo path (this will likely fail, showing error handling)"
./release-tool --repo /nonexistent/repo 2>&1 || echo "✅ Error handling works"

# Create a sample test repository for demo
echo -e "\n🏗️  Creating sample test repository..."
mkdir -p test-repo
cd test-repo
git init
git config user.name "Test User"
git config user.email "test@example.com"

echo "# Test Project" > README.md
git add README.md
git commit -m "chore: initial commit"

echo "## Features" >> README.md
git add README.md
git commit -m "feat: add features section to README"

echo "## Bug Fixes" >> README.md
git add README.md
git commit -m "fix: add bug fixes section"

echo "## Breaking Changes" >> README.md
git add README.md
git commit -m "feat!: add breaking changes section

BREAKING CHANGE: This changes the README format"

# Tag the initial version
git tag v1.0.0

echo "## More Features" >> README.md
git add README.md
git commit -m "feat(readme): add more features section"

echo "## Patches" >> README.md
git add README.md
git commit -m "fix(readme): fix formatting issues"

echo "## Documentation" >> README.md
git add README.md
git commit -m "docs: add documentation section"

echo "## Malformed" >> README.md
git add README.md
git commit -m "this is not a conventional commit"

cd ..

echo -e "\n📊 Test 5: Sample repository analysis"
./release-tool --repo test-repo --verbose

echo -e "\n📊 Test 6: Sample repository JSON output"
./release-tool --repo test-repo --json

echo -e "\n🧹 Cleanup"
rm -rf test-repo

echo -e "\n✅ All tests completed!"
echo "🚀 Usage examples:"
echo "  ./release-tool                           # Analyze current repo"
echo "  ./release-tool --verbose                 # Show detailed commit info"
echo "  ./release-tool --json                    # JSON output for CI"
echo "  ./release-tool --repo /path/to/repo      # Analyze different repo"
echo "  ./release-tool --strict                  # Treat unknown commits as patches"
echo ""
echo "🔧 CI Integration example:"
echo "  NEXT_VERSION=\$(./release-tool --json | jq -r '.next_version')"
echo "  echo \"Next version will be: \$NEXT_VERSION\""
