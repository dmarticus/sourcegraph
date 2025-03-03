package runtype

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

// RunType indicates the type of this run. Each CI pipeline can only be a single run type.
type RunType int

const (
	// RunTypes should be defined by order of precedence.

	PullRequest RunType = iota // pull request build

	// Nightly builds - must be first because they take precedence

	ReleaseNightly // release branch nightly healthcheck builds
	BextNightly    // browser extension nightly build

	// Release branches

	TaggedRelease     // semver-tagged release
	ReleaseBranch     // release branch build
	BextReleaseBranch // browser extension release build

	// Main branches

	MainBranch // main branch build
	MainDryRun // run everything main does, except for deploy-related steps

	// Patches (NOT patch releases)

	ImagePatch          // build a patched image after testing
	ImagePatchNoTest    // build a patched image without testing
	CandidatesNoTest    // build all candidates without testing
	ExecutorPatchNoTest // build executor image without testing

	// Special test branches

	BackendIntegrationTests // run backend tests that are used on main

	// None is a no-op, add all run types above this type.
	None
)

// Compute determines what RunType matches the given parameters.
func Compute(tag, branch string, env map[string]string) RunType {
	for runType := PullRequest + 1; runType < None; runType += 1 {
		if runType.Matcher().Matches(tag, branch, env) {
			return runType
		}
	}
	// RunType is PullRequest by default
	return PullRequest
}

// Is returns true if this run type is one of the given RunTypes
func (t RunType) Is(oneOfTypes ...RunType) bool {
	for _, rt := range oneOfTypes {
		if t == rt {
			return true
		}
	}
	return false
}

// Matcher returns the requirements for a build to be considered of this RunType.
func (t RunType) Matcher() *RunTypeMatcher {
	switch t {
	case ReleaseNightly:
		return &RunTypeMatcher{
			EnvIncludes: map[string]string{
				"RELEASE_NIGHTLY": "true",
			},
		}
	case BextNightly:
		return &RunTypeMatcher{
			EnvIncludes: map[string]string{
				"BEXT_NIGHTLY": "true",
			},
		}

	case TaggedRelease:
		return &RunTypeMatcher{
			TagPrefix: "v",
		}
	case ReleaseBranch:
		return &RunTypeMatcher{
			Branch:       `^[0-9]+\.[0-9]+$`,
			BranchRegexp: true,
		}
	case BextReleaseBranch:
		return &RunTypeMatcher{
			Branch:      "bext/release",
			BranchExact: true,
		}

	case MainBranch:
		return &RunTypeMatcher{
			Branch:      "main",
			BranchExact: true,
		}
	case MainDryRun:
		return &RunTypeMatcher{
			Branch: "main-dry-run/",
		}

	case ImagePatch:
		return &RunTypeMatcher{
			Branch: "docker-images-patch/",
		}
	case ImagePatchNoTest:
		return &RunTypeMatcher{
			Branch: "docker-images-patch-notest/",
		}
	case CandidatesNoTest:
		return &RunTypeMatcher{
			Branch: "docker-images-candidates-notest/",
		}
	case ExecutorPatchNoTest:
		return &RunTypeMatcher{
			Branch: "executor-patch-notest/",
		}

	case BackendIntegrationTests:
		return &RunTypeMatcher{
			Branch: "backend-integration/",
		}
	}

	return nil
}

func (t RunType) String() string {
	switch t {
	case PullRequest:
		return "Pull request"

	case ReleaseNightly:
		return "Release branch nightly healthcheck build"
	case BextNightly:
		return "Browser extension nightly release build"

	case TaggedRelease:
		return "Tagged release"
	case ReleaseBranch:
		return "Release branch"
	case BextReleaseBranch:
		return "Browser extension release build"

	case MainBranch:
		return "Main branch"
	case MainDryRun:
		return "Main dry run"

	case ImagePatch:
		return "Patch image"
	case ImagePatchNoTest:
		return "Patch image without testing"
	case CandidatesNoTest:
		return "Build all candidates without testing"
	case ExecutorPatchNoTest:
		return "Build executor without testing"

	case BackendIntegrationTests:
		return "Backend integration tests"
	}
	return ""
}

// RunTypeMatcher defines the requirements for any given build to be considered a build of
// this RunType.
type RunTypeMatcher struct {
	// Branch loosely matches branches that begin with this value, unless a different type
	// of match is indicated (e.g. BranchExact, BranchRegexp)
	Branch       string
	BranchExact  bool
	BranchRegexp bool

	// TagPrefix matches tags that begin with this value.
	TagPrefix string

	// EnvIncludes validates if these key-value pairs are configured in environment.
	EnvIncludes map[string]string
}

// Matches returns true if the given properties and environment match this RunType.
func (m *RunTypeMatcher) Matches(tag, branch string, env map[string]string) bool {
	if m.Branch != "" {
		switch {
		case m.BranchExact:
			return m.Branch == branch
		case m.BranchRegexp:
			return lazyregexp.New(m.Branch).MatchString(branch)
		default:
			return strings.HasPrefix(branch, m.Branch)
		}
	}

	if m.TagPrefix != "" {
		return strings.HasPrefix(tag, m.TagPrefix)
	}

	if len(m.EnvIncludes) > 0 && len(env) > 0 {
		for wantK, wantV := range m.EnvIncludes {
			gotV, exists := env[wantK]
			if !exists || (wantV != gotV) {
				return false
			}
		}
		return true
	}

	return false
}
