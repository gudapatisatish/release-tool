// Phase 0: Version Calculation Only
// Focus: commit parsing and semantic versioning logic
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"flag"
)

// Core types
type CommitType string

const (
	CommitTypeFeat     CommitType = "feat"
	CommitTypeFix      CommitType = "fix"
	CommitTypeRefactor CommitType = "refactor"
	CommitTypePerf     CommitType = "perf"
	CommitTypeDocs     CommitType = "docs"
	CommitTypeStyle    CommitType = "style"
	CommitTypeTest     CommitType = "test"
	CommitTypeChore    CommitType = "chore"
	CommitTypeUnknown  CommitType = "unknown"
)

type BumpType string

const (
	BumpMajor BumpType = "major"
	BumpMinor BumpType = "minor"
	BumpPatch BumpType = "patch"
	BumpNone  BumpType = "none"
)

type Commit struct {
	Hash        string     `json:"hash"`
	Message     string     `json:"message"`
	Type        CommitType `json:"type"`
	Scope       string     `json:"scope,omitempty"`
	Description string     `json:"description"`
	Body        string     `json:"body,omitempty"`
	Breaking    bool       `json:"breaking"`
	Author      string     `json:"author"`
	Timestamp   time.Time  `json:"timestamp"`
}

type Version struct {
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Patch      int    `json:"patch"`
	PreRelease string `json:"pre_release,omitempty"`
	Build      string `json:"build,omitempty"`
}

func (v Version) String() string {
	base := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		base += "-" + v.PreRelease
	}
	if v.Build != "" {
		base += "+" + v.Build
	}
	return base
}

type VersionCalculationResult struct {
	CurrentVersion Version   `json:"current_version"`
	NextVersion    Version   `json:"next_version"`
	BumpType       BumpType  `json:"bump_type"`
	Commits        []Commit  `json:"commits"`
	CommitsSince   int       `json:"commits_since"`
	Analysis       Analysis  `json:"analysis"`
}

type Analysis struct {
	BreakingChanges int `json:"breaking_changes"`
	Features        int `json:"features"`
	Fixes          int `json:"fixes"`
	Other          int `json:"other"`
	Malformed      int `json:"malformed"`
}

// Services
type GitService interface {
	GetLatestTag(ctx context.Context) (string, error)
	GetCommitsSinceTag(ctx context.Context, tag string) ([]Commit, error)
	IsCleanWorkingDirectory(ctx context.Context) (bool, error)
}

type CommitParser interface {
	ParseCommit(message string, hash string, author string, timestamp time.Time) Commit
}

type VersionCalculator interface {
	CalculateNextVersion(current Version, commits []Commit) (Version, BumpType, Analysis)
}

// Git Service Implementation
type LocalGitService struct {
	repoPath string
}

func NewLocalGitService(repoPath string) *LocalGitService {
	return &LocalGitService{repoPath: repoPath}
}

func (g *LocalGitService) GetLatestTag(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", g.repoPath, "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		// No tags found - start from 0.0.0
		return "0.0.0", nil
	}
	return strings.TrimSpace(string(output)), nil
}

func (g *LocalGitService) GetCommitsSinceTag(ctx context.Context, tag string) ([]Commit, error) {
	var gitRange string
	if tag == "0.0.0" {
		// No previous tag, get all commits
		gitRange = "HEAD"
	} else {
		gitRange = fmt.Sprintf("%s..HEAD", tag)
	}

	// Get commit log with format: hash|author|timestamp|subject|body
	cmd := exec.CommandContext(ctx, "git", "-C", g.repoPath, "log", gitRange,
		"--pretty=format:%H|%an|%at|%s|%b", "--no-merges")
	
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	if len(output) == 0 {
		return []Commit{}, nil
	}

	return g.parseGitLog(string(output))
}

func (g *LocalGitService) parseGitLog(output string) ([]Commit, error) {
	var commits []Commit
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		
		timestamp, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}
		
		commit := Commit{
			Hash:      parts[0],
			Author:    parts[1],
			Timestamp: time.Unix(timestamp, 0),
			Message:   parts[3],
		}
		
		if len(parts) > 4 {
			commit.Body = strings.Join(parts[4:], "|")
		}
		
		commits = append(commits, commit)
	}
	
	return commits, nil
}

func (g *LocalGitService) IsCleanWorkingDirectory(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", g.repoPath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to check git status: %w", err)
	}
	return len(strings.TrimSpace(string(output))) == 0, nil
}

// Commit Parser Implementation
type ConventionalCommitParser struct {
	// Regex patterns for parsing conventional commits
	typeRegex     *regexp.Regexp
	breakingRegex *regexp.Regexp
}

func NewConventionalCommitParser() *ConventionalCommitParser {
	return &ConventionalCommitParser{
		// Matches: type(scope): description
		typeRegex:     regexp.MustCompile(`^(\w+)(\([^)]+\))?\s*:\s*(.+)$`),
		// Matches: BREAKING CHANGE: or BREAKING-CHANGE: or !
		breakingRegex: regexp.MustCompile(`(?i)BREAKING[- ]CHANGE:|!`),
	}
}

func (p *ConventionalCommitParser) ParseCommit(message, hash, author string, timestamp time.Time) Commit {
	commit := Commit{
		Hash:      hash,
		Message:   message,
		Author:    author,
		Timestamp: timestamp,
		Type:      CommitTypeUnknown,
	}

	// Parse first line for type and description
	lines := strings.Split(message, "\n")
	firstLine := strings.TrimSpace(lines[0])
	
	matches := p.typeRegex.FindStringSubmatch(firstLine)
	if len(matches) >= 4 {
		commit.Type = CommitType(strings.ToLower(matches[1]))
		if matches[2] != "" {
			// Remove parentheses from scope
			commit.Scope = strings.Trim(matches[2], "()")
		}
		commit.Description = matches[3]
	} else {
		// Fallback: treat entire first line as description
		commit.Description = firstLine
	}

	// Check for breaking changes
	fullMessage := message
	if len(lines) > 1 {
		commit.Body = strings.Join(lines[1:], "\n")
		fullMessage = message
	}

	commit.Breaking = p.breakingRegex.MatchString(fullMessage) || strings.Contains(firstLine, "!")

	return commit
}

// Version Calculator Implementation
type SemanticVersionCalculator struct {
	strictMode bool // If true, unknown commit types cause errors
}

func NewSemanticVersionCalculator(strictMode bool) *SemanticVersionCalculator {
	return &SemanticVersionCalculator{strictMode: strictMode}
}

func (c *SemanticVersionCalculator) CalculateNextVersion(current Version, commits []Commit) (Version, BumpType, Analysis) {
	analysis := Analysis{}
	bumpType := BumpNone
	
	for _, commit := range commits {
		switch {
		case commit.Breaking:
			analysis.BreakingChanges++
			if bumpType != BumpMajor {
				bumpType = BumpMajor
			}
		case commit.Type == CommitTypeFeat:
			analysis.Features++
			if bumpType != BumpMajor && bumpType != BumpMinor {
				bumpType = BumpMinor
			}
		case commit.Type == CommitTypeFix || commit.Type == CommitTypePerf:
			analysis.Fixes++
			if bumpType == BumpNone {
				bumpType = BumpPatch
			}
		case commit.Type == CommitTypeUnknown:
			analysis.Malformed++
			if c.strictMode {
				// In strict mode, treat unknown as patch
				if bumpType == BumpNone {
					bumpType = BumpPatch
				}
			}
		default:
			analysis.Other++
		}
	}
	
	next := current
	switch bumpType {
	case BumpMajor:
		next.Major++
		next.Minor = 0
		next.Patch = 0
	case BumpMinor:
		next.Minor++
		next.Patch = 0
	case BumpPatch:
		next.Patch++
	}
	
	// Clear pre-release and build for normal releases
	next.PreRelease = ""
	next.Build = ""
	
	return next, bumpType, analysis
}

// Version parsing utility
func ParseVersion(versionStr string) (Version, error) {
	// Remove 'v' prefix if present
	versionStr = strings.TrimPrefix(versionStr, "v")
	
	// Simple semver parsing (doesn't handle all edge cases)
	parts := strings.Split(versionStr, ".")
	if len(parts) < 3 {
		return Version{}, fmt.Errorf("invalid version format: %s", versionStr)
	}
	
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version: %s", parts[0])
	}
	
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version: %s", parts[1])
	}
	
	// Handle patch with potential pre-release/build
	patchPart := parts[2]
	preReleaseIdx := strings.Index(patchPart, "-")
	buildIdx := strings.Index(patchPart, "+")
	
	patchStr := patchPart
	var preRelease, build string
	
	if preReleaseIdx != -1 {
		patchStr = patchPart[:preReleaseIdx]
		remaining := patchPart[preReleaseIdx+1:]
		if buildIdx != -1 {
			buildIdx = strings.Index(remaining, "+")
			if buildIdx != -1 {
				preRelease = remaining[:buildIdx]
				build = remaining[buildIdx+1:]
			} else {
				preRelease = remaining
			}
		} else {
			preRelease = remaining
		}
	} else if buildIdx != -1 {
		patchStr = patchPart[:buildIdx]
		build = patchPart[buildIdx+1:]
	}
	
	patch, err := strconv.Atoi(patchStr)
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version: %s", patchStr)
	}
	
	return Version{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		PreRelease: preRelease,
		Build:      build,
	}, nil
}

// CLI Application
type App struct {
	git        GitService
	parser     CommitParser
	calculator VersionCalculator
}

func NewApp(repoPath string, strictMode bool) *App {
	return &App{
		git:        NewLocalGitService(repoPath),
		parser:     NewConventionalCommitParser(),
		calculator: NewSemanticVersionCalculator(strictMode),
	}
}

func (app *App) CalculateVersion(ctx context.Context) (*VersionCalculationResult, error) {
	// Get current version from latest tag
	latestTag, err := app.git.GetLatestTag(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tag: %w", err)
	}
	
	currentVersion, err := ParseVersion(latestTag)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current version %s: %w", latestTag, err)
	}
	
	// Get commits since last tag
	rawCommits, err := app.git.GetCommitsSinceTag(ctx, latestTag)
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}
	
	// Parse commits
	var commits []Commit
	for _, raw := range rawCommits {
		parsed := app.parser.ParseCommit(raw.Message, raw.Hash, raw.Author, raw.Timestamp)
		commits = append(commits, parsed)
	}
	
	// Calculate next version
	nextVersion, bumpType, analysis := app.calculator.CalculateNextVersion(currentVersion, commits)
	
	return &VersionCalculationResult{
		CurrentVersion: currentVersion,
		NextVersion:    nextVersion,
		BumpType:       bumpType,
		Commits:        commits,
		CommitsSince:   len(commits),
		Analysis:       analysis,
	}, nil
}

func (app *App) PrintAnalysis(result *VersionCalculationResult, verbose bool) {
	fmt.Printf("Version Analysis\n")
	fmt.Printf("==================\n\n")
	
	fmt.Printf("Current Version: %s\n", result.CurrentVersion)
	fmt.Printf("Next Version:    %s\n", result.NextVersion)
	fmt.Printf("Bump Type:       %s\n", result.BumpType)
	fmt.Printf("Commits Since:   %d\n\n", result.CommitsSince)
	
	if result.CommitsSince == 0 {
		fmt.Printf("No new commits since last release\n")
		return
	}
	
	fmt.Printf("Change Summary\n")
	fmt.Printf("-----------------\n")
	fmt.Printf("Breaking Changes: %d\n", result.Analysis.BreakingChanges)
	fmt.Printf("Features:         %d\n", result.Analysis.Features)
	fmt.Printf("Fixes:           %d\n", result.Analysis.Fixes)
	fmt.Printf("Other:           %d\n", result.Analysis.Other)
	
	if result.Analysis.Malformed > 0 {
		fmt.Printf("Malformed:      %d\n", result.Analysis.Malformed)
	}
	
	if verbose {
		fmt.Printf("\nCommit Details\n")
		fmt.Printf("------------------\n")
		
		// Group commits by type for better readability
		commitsByType := make(map[CommitType][]Commit)
		for _, commit := range result.Commits {
			commitsByType[commit.Type] = append(commitsByType[commit.Type], commit)
		}
		
		// Sort types for consistent output
		types := []CommitType{CommitTypeFeat, CommitTypeFix, CommitTypePerf, 
			CommitTypeRefactor, CommitTypeDocs, CommitTypeStyle, 
			CommitTypeTest, CommitTypeChore, CommitTypeUnknown}
		
		for _, commitType := range types {
			commits := commitsByType[commitType]
			if len(commits) == 0 {
				continue
			}
			
			fmt.Printf("\n%s (%d):\n", strings.Title(string(commitType)), len(commits))
			for _, commit := range commits {
				breaking := ""
				if commit.Breaking {
					breaking = " 💥"
				}
				
				fmt.Printf("  • %s: %s%s\n", 
					commit.Hash[:8], 
					commit.Description,
					breaking)
			}
		}
	}
}

func main() {
	var (
		repoPath   = flag.String("repo", ".", "Path to git repository")
		verbose    = flag.Bool("verbose", false, "Show detailed commit analysis")
		jsonOutput = flag.Bool("json", false, "Output result as JSON")
		strictMode = flag.Bool("strict", false, "Treat unknown commit types as patch bumps")
	)
	flag.Parse()
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	app := NewApp(*repoPath, *strictMode)
	
	result, err := app.CalculateVersion(ctx)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	
	if *jsonOutput {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(result); err != nil {
			log.Fatalf("JSON encoding error: %v", err)
		}
	} else {
		app.PrintAnalysis(result, *verbose)
		
		// Provide next steps hint
		if result.CommitsSince > 0 {
			fmt.Printf("\nNext Steps\n")
			fmt.Printf("-------------\n")
			fmt.Printf("To release version %s:\n", result.NextVersion)
			fmt.Printf("1. Review the changes above\n")
			fmt.Printf("2. Run: git tag %s\n", result.NextVersion)
			fmt.Printf("3. Run: git push origin %s\n", result.NextVersion)
		}
	}
	
	// Exit with appropriate code
	if result.Analysis.Malformed > 0 && *strictMode {
		fmt.Printf("\nWarning: %d malformed commits found in strict mode\n", result.Analysis.Malformed)
		os.Exit(1)
	}
}