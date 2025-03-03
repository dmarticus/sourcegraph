package batches

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/batches/env"
	"github.com/sourcegraph/sourcegraph/lib/batches/overridable"
	"github.com/sourcegraph/sourcegraph/lib/batches/schema"
	"github.com/sourcegraph/sourcegraph/lib/batches/template"
	"github.com/sourcegraph/sourcegraph/lib/batches/yaml"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Some general notes about the struct definitions below.
//
// 1. They map _very_ closely to the batch spec JSON schema. We don't
//    auto-generate the types because we need YAML support (more on that in a
//    moment) and because no generator can currently handle oneOf fields
//    gracefully in Go, but that's a potential future enhancement.
//
// 2. Fields are tagged with _both_ JSON and YAML tags. Internally, the JSON
//    schema library needs to be able to marshal the struct to JSON for
//    validation, so we need to ensure that we're generating the right JSON to
//    represent the YAML that we unmarshalled.
//
// 3. All JSON tags include omitempty so that the schema validation can pick up
//    omitted fields. The other option here was to have everything unmarshal to
//    pointers, which is ugly and inefficient.

type BatchSpec struct {
	Name              string                   `json:"name,omitempty" yaml:"name"`
	Description       string                   `json:"description,omitempty" yaml:"description"`
	On                []OnQueryOrRepository    `json:"on,omitempty" yaml:"on"`
	Workspaces        []WorkspaceConfiguration `json:"workspaces,omitempty"  yaml:"workspaces"`
	Steps             []Step                   `json:"steps,omitempty" yaml:"steps"`
	TransformChanges  *TransformChanges        `json:"transformChanges,omitempty" yaml:"transformChanges,omitempty"`
	ImportChangesets  []ImportChangeset        `json:"importChangesets,omitempty" yaml:"importChangesets"`
	ChangesetTemplate *ChangesetTemplate       `json:"changesetTemplate,omitempty" yaml:"changesetTemplate"`
}

type ChangesetTemplate struct {
	Title     string                       `json:"title,omitempty" yaml:"title"`
	Body      string                       `json:"body,omitempty" yaml:"body"`
	Branch    string                       `json:"branch,omitempty" yaml:"branch"`
	Commit    ExpandedGitCommitDescription `json:"commit,omitempty" yaml:"commit"`
	Published *overridable.BoolOrString    `json:"published" yaml:"published"`
}

type GitCommitAuthor struct {
	Name  string `json:"name" yaml:"name"`
	Email string `json:"email" yaml:"email"`
}

type ExpandedGitCommitDescription struct {
	Message string           `json:"message,omitempty" yaml:"message"`
	Author  *GitCommitAuthor `json:"author,omitempty" yaml:"author"`
}

type ImportChangeset struct {
	Repository  string        `json:"repository" yaml:"repository"`
	ExternalIDs []interface{} `json:"externalIDs" yaml:"externalIDs"`
}

type WorkspaceConfiguration struct {
	RootAtLocationOf   string `json:"rootAtLocationOf,omitempty" yaml:"rootAtLocationOf"`
	In                 string `json:"in,omitempty" yaml:"in"`
	OnlyFetchWorkspace bool   `json:"onlyFetchWorkspace,omitempty" yaml:"onlyFetchWorkspace"`
}

type OnQueryOrRepository struct {
	RepositoriesMatchingQuery string   `json:"repositoriesMatchingQuery,omitempty" yaml:"repositoriesMatchingQuery"`
	Repository                string   `json:"repository,omitempty" yaml:"repository"`
	Branch                    string   `json:"branch,omitempty" yaml:"branch"`
	Branches                  []string `json:"branches,omitempty" yaml:"branches"`
}

var ErrConflictingBranches = NewValidationError(errors.New("both branch and branches specified"))

func (oqor *OnQueryOrRepository) GetBranches() ([]string, error) {
	if oqor.Branch != "" {
		if len(oqor.Branches) > 0 {
			return nil, ErrConflictingBranches
		}
		return []string{oqor.Branch}, nil
	}
	return oqor.Branches, nil
}

type Step struct {
	Run       string            `json:"run,omitempty" yaml:"run"`
	Container string            `json:"container,omitempty" yaml:"container"`
	Env       env.Environment   `json:"env,omitempty" yaml:"env"`
	Files     map[string]string `json:"files,omitempty" yaml:"files,omitempty"`
	Outputs   Outputs           `json:"outputs,omitempty" yaml:"outputs,omitempty"`

	If interface{} `json:"if,omitempty" yaml:"if,omitempty"`
}

func (s *Step) IfCondition() string {
	switch v := s.If.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return v
	default:
		return ""
	}
}

type Outputs map[string]Output

type Output struct {
	Value  string `json:"value,omitempty" yaml:"value,omitempty"`
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
}

type TransformChanges struct {
	Group []Group `json:"group,omitempty" yaml:"group"`
}

type Group struct {
	Directory  string `json:"directory,omitempty" yaml:"directory"`
	Branch     string `json:"branch,omitempty" yaml:"branch"`
	Repository string `json:"repository,omitempty" yaml:"repository"`
}

type ParseBatchSpecOptions struct {
	AllowArrayEnvironments bool
	AllowTransformChanges  bool
	AllowConditionalExec   bool
}

func ParseBatchSpec(data []byte, opts ParseBatchSpecOptions) (*BatchSpec, error) {
	return parseBatchSpec(schema.BatchSpecJSON, data, opts)
}

func parseBatchSpec(schema string, data []byte, opts ParseBatchSpecOptions) (*BatchSpec, error) {
	var spec BatchSpec
	if err := yaml.UnmarshalValidate(schema, data, &spec); err != nil {
		var multiErr *errors.MultiError
		if errors.As(err, &multiErr) {
			var newMultiError *errors.MultiError

			for _, e := range multiErr.Errors {
				// In case of `name` we try to make the error message more user-friendly.
				if strings.Contains(e.Error(), "name: Does not match pattern") {
					newMultiError = errors.Append(newMultiError, NewValidationError(errors.Newf("The batch change name can only contain word characters, dots and dashes. No whitespace or newlines allowed.")))
				} else {
					newMultiError = errors.Append(newMultiError, NewValidationError(e))
				}
			}

			return nil, newMultiError.ErrorOrNil()
		}

		return nil, err
	}

	var errs *errors.MultiError

	if !opts.AllowArrayEnvironments {
		for i, step := range spec.Steps {
			if !step.Env.IsStatic() {
				errs = errors.Append(errs, NewValidationError(errors.Errorf("step %d includes one or more dynamic environment variables, which are unsupported in this Sourcegraph version", i+1)))
			}
		}
	}

	if len(spec.Steps) != 0 && spec.ChangesetTemplate == nil {
		errs = errors.Append(errs, NewValidationError(errors.New("batch spec includes steps but no changesetTemplate")))
	}

	if spec.TransformChanges != nil && !opts.AllowTransformChanges {
		errs = errors.Append(errs, NewValidationError(errors.New("batch spec includes transformChanges, which is not supported in this Sourcegraph version")))
	}

	if len(spec.Workspaces) != 0 && !opts.AllowTransformChanges {
		errs = errors.Append(errs, NewValidationError(errors.New("batch spec includes workspaces, which is not supported in this Sourcegraph version")))
	}

	if !opts.AllowConditionalExec {
		for i, step := range spec.Steps {
			if step.IfCondition() != "" {
				errs = errors.Append(errs, NewValidationError(errors.Newf(
					"step %d in batch spec uses the 'if' attribute for conditional execution, which is not supported in this Sourcegraph version",
					i+1,
				)))
			}
		}
	}

	return &spec, errs.ErrorOrNil()
}

func (on *OnQueryOrRepository) String() string {
	if on.RepositoriesMatchingQuery != "" {
		return on.RepositoriesMatchingQuery
	} else if on.Repository != "" {
		return "r:" + on.Repository
	}

	return fmt.Sprintf("%v", *on)
}

// BatchSpecValidationError is returned when parsing/using values from the batch spec failed.
type BatchSpecValidationError struct {
	err error
}

func NewValidationError(err error) BatchSpecValidationError {
	return BatchSpecValidationError{err}
}

func (e BatchSpecValidationError) Error() string {
	return e.err.Error()
}

func IsValidationError(err error) bool {
	return errors.HasType(err, &BatchSpecValidationError{})
}

// SkippedStepsForRepo calculates the steps required to run on the given repo.
func SkippedStepsForRepo(spec *BatchSpec, repoName string, fileMatches []string) (skipped map[int32]struct{}, err error) {
	skipped = map[int32]struct{}{}

	for idx, step := range spec.Steps {
		// If no if condition is given, just go ahead and add the step to the list.
		if step.IfCondition() == "" {
			continue
		}

		batchChange := template.BatchChangeAttributes{
			Name:        spec.Name,
			Description: spec.Description,
		}
		// TODO: This step ctx is incomplete, is this allowed?
		// We can at least optimize further here and do more static evaluation
		// when we have a cached result for the previous step.
		stepCtx := &template.StepContext{
			Repository: template.Repository{
				Name:        repoName,
				FileMatches: fileMatches,
			},
			BatchChange: batchChange,
		}
		static, boolVal, err := template.IsStaticBool(step.IfCondition(), stepCtx)
		if err != nil {
			return nil, err
		}

		if static && !boolVal {
			skipped[int32(idx)] = struct{}{}
		}
	}

	return skipped, nil
}
