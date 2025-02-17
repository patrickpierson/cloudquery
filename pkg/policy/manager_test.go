package policy

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudquery/cloudquery/internal/file"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManagerImpl_DownloadPolicy(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestManagerImpl_DownloadPolicy")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := NewManager(tmpDir, nil, hclog.New(&hclog.LoggerOptions{}))

	cases := []struct {
		Name           string
		PolicyPath     string
		RepositoryPath string
	}{
		{
			Name:           "policy_hub_policy",
			PolicyPath:     "cloudquery/cq-policy-core",
			RepositoryPath: "test",
		},
		{
			Name:       "private_policy_main_branch",
			PolicyPath: "michelvocks/my-cq-policy",
		},
		{
			Name:       "private_policy_master_branch",
			PolicyPath: "michelvocks/cq-test-policy",
		},
	}

	osFs := file.NewOsFs()
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			policyPath := []string{tc.PolicyPath, tc.RepositoryPath}
			p, err := m.ParsePolicyHubPath(policyPath, "")
			assert.NoError(t, err)

			// Download policy
			err = m.DownloadPolicy(context.Background(), p)
			assert.NoError(t, err)

			// Make sure downloaded policy folder exists
			policyFolder := filepath.Join(tmpDir, p.Organization, p.Repository)
			_, err = osFs.Stat(policyFolder)
			assert.NoError(t, err)

			// Download policy again (which should always work)
			err = m.DownloadPolicy(context.Background(), p)
			assert.NoError(t, err)
		})
	}
}

func TestManagerImpl_RunPolicy(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "TestManagerImpl_RunPolicy")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup database
	pool, tearDownFunc := setupDatabase(t, "test_policy_table")
	defer tearDownFunc(t)

	m := NewManager(tmpDir, pool, hclog.New(&hclog.LoggerOptions{}))

	cases := []struct {
		Name             string
		PolicyPath       string
		RepositoryPath   string
		ProviderVersions map[string]*version.Version
		ErrorString      string
	}{
		{
			Name:           "policy_hub_policy",
			PolicyPath:     "cloudquery/cq-policy-core",
			RepositoryPath: "test",
			ProviderVersions: map[string]*version.Version{
				"aws": version.Must(version.NewVersion("v1.0")),
			},
		},
		{
			Name:       "private_policy_main_branch",
			PolicyPath: "michelvocks/my-cq-policy@v0.0.2",
			ProviderVersions: map[string]*version.Version{
				"aws": version.Must(version.NewVersion("1.0.0")),
			},
		},
		{
			Name:       "private_policy_query_in_file",
			PolicyPath: "fdistorted/my-cq-policy@v0.0.4",
			ProviderVersions: map[string]*version.Version{
				"aws": version.Must(version.NewVersion("1.0.0")),
			},
		},
		{
			Name:       "too old provider",
			PolicyPath: "michelvocks/my-cq-policy@v0.0.2",
			ProviderVersions: map[string]*version.Version{
				"aws": version.Must(version.NewVersion("0.5")),
			},
			ErrorString: "test-policy: provider aws does not satisfy version requirement >= 1.0",
		},
		{
			Name:        "provider version unknown",
			PolicyPath:  "michelvocks/my-cq-policy@v0.0.2",
			ErrorString: "test-policy: provider aws version is unknown",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			policyHubPath := []string{tc.PolicyPath, tc.RepositoryPath}
			p, err := m.ParsePolicyHubPath(policyHubPath, "")
			assert.NoError(t, err)

			if err := m.DownloadPolicy(context.Background(), p); err != nil {
				t.Fatal(err)
			}

			results, err := m.RunPolicy(context.Background(), &ExecuteRequest{
				Policy:           p,
				UpdateCallback:   nil,
				StopOnFailure:    true,
				ProviderVersions: tc.ProviderVersions,
			})
			if tc.ErrorString == "" {
				require.NoError(t, err)
				assert.True(t, results.Passed)

				// Make sure all expected keys are contained
				expectedKeys := []string{
					"test-policy/top-level-query",
					"test-policy/sub-policy-1/sub-level-query",
					"test-policy/sub-policy-2/sub-level-query",
				}
				for k := range results.Results {
					assert.Contains(t, expectedKeys, k)
				}
			} else {
				require.Error(t, err)
				assert.Equal(t, tc.ErrorString, err.Error())
			}
		})
	}
}
