package batches

import (
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
)

func TestParseBatchSpec(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		const spec = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3
changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err != nil {
			t.Fatalf("parsing valid spec returned error: %s", err)
		}
	})

	t.Run("missing changesetTemplate", func(t *testing.T) {
		const spec = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err == nil {
			t.Fatal("no error returned")
		}

		wantErr := `1 error occurred:
	* batch spec includes steps but no changesetTemplate

`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})

	t.Run("invalid batch change name", func(t *testing.T) {
		const spec = `
name: this name is invalid cause it contains whitespace
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3
changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err == nil {
			t.Fatal("no error returned")
		}

		// We expect this error to be user-friendly, which is why we test for
		// it specifically here.
		wantErr := `1 error occurred:
	* The batch change name can only contain word characters, dots and dashes. No whitespace or newlines allowed.

`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})

	t.Run("uses unsupported conditional exec", func(t *testing.T) {
		const spec = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    if: "false"
    container: alpine:3

changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err == nil {
			t.Fatal("no error returned")
		}

		wantErr := `1 error occurred:
	* step 1 in batch spec uses the 'if' attribute for conditional execution, which is not supported in this Sourcegraph version

`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})

	t.Run("parsing if attribute", func(t *testing.T) {
		const specTemplate = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repositoriesMatchingQuery: file:README.md
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    if: %s
    container: alpine:3

changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		for _, tt := range []struct {
			raw  string
			want string
		}{
			{raw: `"true"`, want: "true"},
			{raw: `"false"`, want: "false"},
			{raw: `true`, want: "true"},
			{raw: `false`, want: "false"},
			{raw: `"${{ foobar }}"`, want: "${{ foobar }}"},
			{raw: `${{ foobar }}`, want: "${{ foobar }}"},
			{raw: `foobar`, want: "foobar"},
		} {
			spec := fmt.Sprintf(specTemplate, tt.raw)
			batchSpec, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{AllowConditionalExec: true})
			if err != nil {
				t.Fatal(err)
			}

			if batchSpec.Steps[0].IfCondition() != tt.want {
				t.Fatalf("wrong IfCondition. want=%q, got=%q", tt.want, batchSpec.Steps[0].IfCondition())
			}
		}
	})
	t.Run("uses conflicting branch attributes", func(t *testing.T) {
		const spec = `
name: hello-world
description: Add Hello World to READMEs
on:
  - repository: github.com/foo/bar
    branch: foo
    branches: [bar]
steps:
  - run: echo Hello World | tee -a $(find -name README.md)
    container: alpine:3

changesetTemplate:
  title: Hello World
  body: My first batch change!
  branch: hello-world
  commit:
    message: Append Hello World to all README.md files
  published: false
`

		_, err := ParseBatchSpec([]byte(spec), ParseBatchSpecOptions{})
		if err == nil {
			t.Fatal("no error returned")
		}

		wantErr := `3 errors occurred:
	* on.0: Must validate one and only one schema (oneOf)
	* on.0: Must validate at least one schema (anyOf)
	* on.0: Must validate one and only one schema (oneOf)

`
		haveErr := err.Error()
		if haveErr != wantErr {
			t.Fatalf("wrong error. want=%q, have=%q", wantErr, haveErr)
		}
	})
}

func TestOnQueryOrRepository_Branches(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			input *OnQueryOrRepository
			want  []string
		}{
			"no branches": {
				input: &OnQueryOrRepository{},
				want:  nil,
			},
			"single branch": {
				input: &OnQueryOrRepository{Branch: "foo"},
				want:  []string{"foo"},
			},
			"single branch, non-nil but empty branches": {
				input: &OnQueryOrRepository{
					Branch:   "foo",
					Branches: []string{},
				},
				want: []string{"foo"},
			},
			"multiple branches": {
				input: &OnQueryOrRepository{
					Branches: []string{"foo", "bar"},
				},
				want: []string{"foo", "bar"},
			},
		} {
			t.Run(name, func(t *testing.T) {
				have, err := tc.input.GetBranches()
				assert.Nil(t, err)
				assert.Equal(t, tc.want, have)
			})
		}
	})

	t.Run("error", func(t *testing.T) {
		_, err := (&OnQueryOrRepository{
			Branch:   "foo",
			Branches: []string{"bar"},
		}).GetBranches()
		assert.Equal(t, ErrConflictingBranches, err)
	})
}

func TestSkippedStepsForRepo(t *testing.T) {
	tests := map[string]struct {
		spec        *BatchSpec
		wantSkipped []int32
	}{
		"no if": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1"},
				},
			},
			wantSkipped: []int32{},
		},

		"if has static true value": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: "true"},
				},
			},
			wantSkipped: []int32{},
		},

		"one of many steps has if with static true value": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "true"},
					{Run: "echo 3"},
				},
			},
			wantSkipped: []int32{},
		},

		"if has static non-true value": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: "this is not true"},
				},
			},
			wantSkipped: []int32{0},
		},

		"one of many steps has if with static non-true value": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "every type system needs generics"},
					{Run: "echo 3"},
				},
			},
			wantSkipped: []int32{1},
		},

		"if expression that can be partially evaluated to true": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/src*" }}`},
				},
			},
			wantSkipped: []int32{},
		},

		"if expression that can be partially evaluated to false": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSkipped: []int32{0},
		},

		"one of many steps has if expression that can be evaluated to false": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: `${{ matches repository.name "horse" }}`},
					{Run: "echo 3"},
				},
			},
			wantSkipped: []int32{1},
		},

		"if expression that can NOT be partially evaluated": {
			spec: &BatchSpec{
				Steps: []Step{
					{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
				},
			},
			wantSkipped: []int32{},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			haveSkipped, err := SkippedStepsForRepo(tt.spec, "github.com/sourcegraph/src-cli", []string{})
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			want := tt.wantSkipped
			sort.Sort(sortableInt32(want))
			have := make([]int32, 0, len(haveSkipped))
			for s := range haveSkipped {
				have = append(have, s)
			}
			sort.Sort(sortableInt32(have))
			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

type sortableInt32 []int32

func (s sortableInt32) Len() int { return len(s) }

func (s sortableInt32) Less(i, j int) bool { return s[i] < s[j] }

func (s sortableInt32) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
