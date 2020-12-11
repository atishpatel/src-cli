	"github.com/google/go-cmp/cmp"

	changesetTemplateBranch := "my-branch"

	type filesByBranch map[string][]string
	type filesByRepository map[string]filesByBranch

		repos     []*graphql.Repository
		archives  []mockRepoArchive
		steps     []Step
		transform *TransformChanges
		wantFilesChanged filesByRepository
			wantFilesChanged: filesByRepository{
				srcCLIRepo.ID: filesByBranch{
					changesetTemplateBranch: []string{"README.md", "main.go"},
				},
				sourcegraphRepo.ID: {
					changesetTemplateBranch: []string{"README.md"},
				},

			wantFilesChanged: filesByRepository{
				srcCLIRepo.ID: filesByBranch{
					changesetTemplateBranch: []string{"main.go", "modified-main.go.md", "added-modified-main.go.md"},
				},
			wantFilesChanged: filesByRepository{},
		},
		{
			name:  "transform group",
			repos: []*graphql.Repository{srcCLIRepo, sourcegraphRepo},
			archives: []mockRepoArchive{
				{repo: srcCLIRepo, files: map[string]string{
					"README.md":  "# Welcome to the README\n",
					"a/a.go":     "package a",
					"a/b/b.go":   "package b",
					"a/b/c/c.go": "package c",
				}},
				{repo: sourcegraphRepo, files: map[string]string{
					"README.md":  "# Welcome to the README\n",
					"a/a.go":     "package a",
					"a/b/b.go":   "package b",
					"a/b/c/c.go": "package c",
				}},
			},
			steps: []Step{
				{Run: `echo 'var a = 1' >> a/a.go`, Container: "doesntmatter:13"},
				{Run: `echo 'var b = 2' >> a/b/b.go`, Container: "doesntmatter:13"},
				{Run: `echo 'var c = 3' >> a/b/c/c.go`, Container: "doesntmatter:13"},
			},
			transform: &TransformChanges{
				Group: []Group{
					{Directory: "a/b/c", Branch: "in-directory-c"},
					{Directory: "a/b", Branch: "in-directory-b", Repository: sourcegraphRepo.Name},
				},
			},
			wantFilesChanged: filesByRepository{
				srcCLIRepo.ID: filesByBranch{
					changesetTemplateBranch: []string{
						"a/a.go",
						"a/b/b.go",
					},
					"in-directory-c": []string{
						"a/b/c/c.go",
					},
				},
				sourcegraphRepo.ID: filesByBranch{
					changesetTemplateBranch: []string{
						"a/a.go",
					},
					"in-directory-b": []string{
						"a/b/b.go",
						"a/b/c/c.go",
					},
				},
			},
				template := &ChangesetTemplate{Branch: changesetTemplateBranch}

					executor.AddTask(r, tc.steps, tc.transform, template)
				wantSpecs := 0
				for _, byBranch := range tc.wantFilesChanged {
					wantSpecs += len(byBranch)
				}
				if have, want := len(specs), wantSpecs; have != want {
					branch := strings.ReplaceAll(spec.HeadRef, "refs/heads/", "")
					wantFilesInBranch, ok := wantFiles[branch]
					if !ok {
						t.Fatalf("spec for repo %q and branch %q but no files expected in that branch", spec.BaseRepository, branch)
					}

					fileDiffs, err := diff.ParseMultiFileDiff([]byte(spec.Commits[0].Diff))
					if err != nil {
						t.Fatalf("failed to parse diff: %s", err)
					}

					if have, want := len(fileDiffs), len(wantFilesInBranch); have != want {
					for _, file := range wantFilesInBranch {
func TestValidateGroups(t *testing.T) {
	repoName := "github.com/sourcegraph/src-cli"
	defaultBranch := "my-campaign"

	tests := []struct {
		defaultBranch string
		groups        []Group
		wantErr       string
	}{
		{
			groups: []Group{
				{Directory: "a", Branch: "my-campaign-a"},
				{Directory: "b", Branch: "my-campaign-b"},
			},
			wantErr: "",
		},
		{
			groups: []Group{
				{Directory: "a", Branch: "my-campaign-SAME"},
				{Directory: "b", Branch: "my-campaign-SAME"},
			},
			wantErr: "transformChanges would lead to multiple changesets in repository github.com/sourcegraph/src-cli to have the same branch \"my-campaign-SAME\"",
		},
		{
			groups: []Group{
				{Directory: "a", Branch: "my-campaign-SAME"},
				{Directory: "b", Branch: defaultBranch},
			},
			wantErr: "transformChanges group branch for repository github.com/sourcegraph/src-cli is the same as branch \"my-campaign\" in changesetTemplate",
		},
	}

	for _, tc := range tests {
		err := validateGroups(repoName, defaultBranch, tc.groups)
		var haveErr string
		if err != nil {
			haveErr = err.Error()
		}

		if haveErr != tc.wantErr {
			t.Fatalf("wrong error:\nwant=%q\nhave=%q", tc.wantErr, haveErr)
		}
	}
}

func TestGroupFileDiffs(t *testing.T) {
	diff1 := `diff --git 1/1.txt 1/1.txt
new file mode 100644
index 0000000..19d6416
--- /dev/null
+++ 1/1.txt
@@ -0,0 +1,1 @@
+this is 1
`
	diff2 := `diff --git 1/2/2.txt 1/2/2.txt
new file mode 100644
index 0000000..c825d65
--- /dev/null
+++ 1/2/2.txt
@@ -0,0 +1,1 @@
+this is 2
`
	diff3 := `diff --git 1/2/3/3.txt 1/2/3/3.txt
new file mode 100644
index 0000000..1bd79fb
--- /dev/null
+++ 1/2/3/3.txt
@@ -0,0 +1,1 @@
+this is 3
`

	defaultBranch := "my-default-branch"
	allDiffs := diff1 + diff2 + diff3

	tests := []struct {
		diff          string
		defaultBranch string
		groups        []Group
		want          map[string]string
	}{
		{
			diff: allDiffs,
			groups: []Group{
				{Directory: "1/2/3", Branch: "everything-in-3"},
			},
			want: map[string]string{
				"my-default-branch": diff1 + diff2,
				"everything-in-3":   diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []Group{
				{Directory: "1/2", Branch: "everything-in-2-and-3"},
			},
			want: map[string]string{
				"my-default-branch":     diff1,
				"everything-in-2-and-3": diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []Group{
				{Directory: "1", Branch: "everything-in-1-and-2-and-3"},
			},
			want: map[string]string{
				"my-default-branch":           "",
				"everything-in-1-and-2-and-3": diff1 + diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []Group{
				// Each diff is matched against each directory, last match wins
				{Directory: "1", Branch: "only-in-1"},
				{Directory: "1/2", Branch: "only-in-2"},
				{Directory: "1/2/3", Branch: "only-in-3"},
			},
			want: map[string]string{
				"my-default-branch": "",
				"only-in-3":         diff3,
				"only-in-2":         diff2,
				"only-in-1":         diff1,
			},
		},
		{
			diff: allDiffs,
			groups: []Group{
				// Last one wins here, because it matches every diff
				{Directory: "1/2/3", Branch: "only-in-3"},
				{Directory: "1/2", Branch: "only-in-2"},
				{Directory: "1", Branch: "only-in-1"},
			},
			want: map[string]string{
				"my-default-branch": "",
				"only-in-1":         diff1 + diff2 + diff3,
			},
		},
		{
			diff: allDiffs,
			groups: []Group{
				{Directory: "", Branch: "everything"},
			},
			want: map[string]string{
				"my-default-branch": diff1 + diff2 + diff3,
			},
		},
	}

	for _, tc := range tests {
		have, err := groupFileDiffs(tc.diff, defaultBranch, tc.groups)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !cmp.Equal(tc.want, have) {
			t.Errorf("mismatch (-want +got):\n%s", cmp.Diff(tc.want, have))
		}
	}
}

	cache map[string]string
		cache: make(map[string]string),
func (c *inMemoryExecutionCache) Get(ctx context.Context, key ExecutionCacheKey) (string, bool, error) {
		return "", false, err
	if diff, ok := c.cache[k]; ok {
		return diff, true, nil
	return "", false, nil
func (c *inMemoryExecutionCache) Set(ctx context.Context, key ExecutionCacheKey, diff string) error {
	c.cache[k] = diff