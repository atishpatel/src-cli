package service

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

func TestFindWorkspaces(t *testing.T) {
	repos := []*graphql.Repository{
		{ID: "repo-id-0", Name: "github.com/sourcegraph/automation-testing"},
		{ID: "repo-id-1", Name: "github.com/sourcegraph/sourcegraph"},
		{ID: "repo-id-2", Name: "bitbucket.sgdev.org/SOUR/automation-testing"},
	}
	steps := []batcheslib.Step{{Run: "echo 1"}}

	type finderResults map[*graphql.Repository][]string

	tests := map[string]struct {
		spec          *batcheslib.BatchSpec
		finderResults map[*graphql.Repository][]string

		// workspaces in which repo/path they are executed
		wantWorkspaces []RepoWorkspace
	}{
		"no workspace configuration": {
			spec:          &batcheslib.BatchSpec{Steps: steps},
			finderResults: finderResults{},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[0].ID, Steps: steps, Path: ""},
				{RepoID: repos[1].ID, Steps: steps, Path: ""},
				{RepoID: repos[2].ID, Steps: steps, Path: ""},
			},
		},

		"workspace configuration matching no repos": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{In: "this-does-not-match", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[0].ID, Steps: steps, Path: ""},
				{RepoID: repos[1].ID, Steps: steps, Path: ""},
				{RepoID: repos[2].ID, Steps: steps, Path: ""},
			},
		},

		"workspace configuration matching 2 repos with no results": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{In: "*automation-testing", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{
				repos[0]: []string{},
				repos[2]: []string{},
			},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[1].ID, Steps: steps, Path: ""},
			},
		},

		"workspace configuration matching 2 repos with 3 results each": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{In: "*automation-testing", RootAtLocationOf: "package.json"},
				},
			},
			finderResults: finderResults{
				repos[0]: {"a/b", "a/b/c", "d/e/f"},
				repos[2]: {"a/b", "a/b/c", "d/e/f"},
			},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[0].ID, Steps: steps, Path: "a/b"},
				{RepoID: repos[0].ID, Steps: steps, Path: "a/b/c"},
				{RepoID: repos[0].ID, Steps: steps, Path: "d/e/f"},
				{RepoID: repos[1].ID, Steps: steps, Path: ""},
				{RepoID: repos[2].ID, Steps: steps, Path: "a/b"},
				{RepoID: repos[2].ID, Steps: steps, Path: "a/b/c"},
				{RepoID: repos[2].ID, Steps: steps, Path: "d/e/f"},
			},
		},

		"workspace configuration matches repo with OnlyFetchWorkspace": {
			spec: &batcheslib.BatchSpec{
				Steps: steps,
				Workspaces: []batcheslib.WorkspaceConfiguration{
					{
						OnlyFetchWorkspace: true,
						In:                 "*automation-testing",
						RootAtLocationOf:   "package.json",
					},
				},
			},
			finderResults: finderResults{
				repos[0]: {"a/b", "a/b/c", "d/e/f"},
				repos[2]: {"a/b", "a/b/c", "d/e/f"},
			},
			wantWorkspaces: []RepoWorkspace{
				{RepoID: repos[0].ID, Steps: steps, Path: "a/b", OnlyFetchWorkspace: true},
				{RepoID: repos[0].ID, Steps: steps, Path: "a/b/c", OnlyFetchWorkspace: true},
				{RepoID: repos[0].ID, Steps: steps, Path: "d/e/f", OnlyFetchWorkspace: true},
				{RepoID: repos[1].ID, Steps: steps, Path: ""},
				{RepoID: repos[2].ID, Steps: steps, Path: "a/b", OnlyFetchWorkspace: true},
				{RepoID: repos[2].ID, Steps: steps, Path: "a/b/c", OnlyFetchWorkspace: true},
				{RepoID: repos[2].ID, Steps: steps, Path: "d/e/f", OnlyFetchWorkspace: true},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			finder := &mockDirectoryFinder{results: tt.finderResults}
			workspaces, err := findWorkspaces(context.Background(), tt.spec, finder, repos)
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			if diff := cmp.Diff(tt.wantWorkspaces, workspaces); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

type mockDirectoryFinder struct {
	results map[*graphql.Repository][]string
}

func (m *mockDirectoryFinder) FindDirectoriesInRepos(ctx context.Context, fileName string, repos ...*graphql.Repository) (map[*graphql.Repository][]string, error) {
	return m.results, nil
}

func TestStepsForRepo(t *testing.T) {
	tests := map[string]struct {
		spec *batcheslib.BatchSpec

		wantSteps []batcheslib.Step
	}{
		"no if": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
			},
		},

		"if has static true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: "true"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: "true"},
			},
		},

		"one of many steps has if with static true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "true"},
					{Run: "echo 3"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
				{Run: "echo 2", If: "true"},
				{Run: "echo 3"},
			},
		},

		"if has static non-true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: "this is not true"},
				},
			},
			wantSteps: []batcheslib.Step{},
		},

		"one of many steps has if with static non-true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "every type system needs generics"},
					{Run: "echo 3"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
				{Run: "echo 3"},
			},
		},

		"if expression that can be partially evaluated to true": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/src*" }}`},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/src*" }}`},
			},
		},

		"if expression that can be partially evaluated to false": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSteps: []batcheslib.Step{},
		},

		"one of many steps has if expression that can be evaluated to true": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSteps: []batcheslib.Step{},
		},

		"if expression that can NOT be partially evaluated": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			haveSteps, err := stepsForRepo(tt.spec, testRepo1)
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			opts := cmpopts.IgnoreUnexported(batcheslib.Step{})
			if diff := cmp.Diff(tt.wantSteps, haveSteps, opts); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}