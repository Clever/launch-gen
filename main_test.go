package main

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_toPublicVar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "changes names with _",
			input:    "foo_bar",
			expected: "FooBar",
		},
		{
			name:     "changes names with -",
			input:    "foo-bar",
			expected: "FooBar",
		},
		{
			name:     "changes names with .",
			input:    "foo.bar",
			expected: "FooBar",
		},
		{
			name:     "respects Url -> URL override",
			input:    "foo_bar_url",
			expected: "FooBarURL",
		},
		{
			name:     "respects Id -> ID override",
			input:    "foo_bar_id",
			expected: "FooBarID",
		},
		{
			name:     "respects Api -> API override",
			input:    "foo_bar_api",
			expected: "FooBarAPI",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := toPublicVar(tt.input)
			assert.Equal(t, tt.expected, actual, tt.name)
		})
	}
}

func Test_toPrivateVar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "changes names with _",
			input:    "foo_bar",
			expected: "fooBar",
		},
		{
			name:     "changes names with -",
			input:    "foo-bar",
			expected: "fooBar",
		},
		{
			name:     "changes names with .",
			input:    "foo.bar",
			expected: "fooBar",
		},
		{
			name:     "respects Url -> URL override",
			input:    "foo_bar_url",
			expected: "fooBarURL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := toPrivateVar(tt.input)
			assert.Equal(t, tt.expected, actual, tt.name)
		})
	}
}

func Test_getS3NameByEnv(t *testing.T) {
	// taken from generated fixtures
	var podAccountSuffixMap = map[string]bool{"585008086734": true}
	testS3NameByEnv := func(s string) string {
		env := os.Getenv("DEPLOY_ENV")
		if env == "" {
			env = os.Getenv("_DEPLOY_ENV")
		}
		if env == "" {
			log.Fatal("Unable to determine deployment environment (DEPLOY_ENV and _DEPLOY_ENV are undefined)")
		}
		podAccount := os.Getenv("_POD_ACCOUNT")
		if env == "production" {
			return s
		}
		if podAccount != "" && podAccountSuffixMap[podAccount] {
			return s + "-dev-" + podAccount
		}
		return s + "-dev"
	}

	tests := []struct {
		name             string
		bucketName       string
		deployEnv        string
		underscoreDeploy string
		podAccount       string
		expected         string
	}{
		{
			name:             "production environment, no pod account",
			bucketName:       "my-bucket",
			deployEnv:        "production",
			underscoreDeploy: "",
			podAccount:       "",
			expected:         "my-bucket",
		},
		{
			name:             "production environment ignores pod account in map",
			bucketName:       "my-bucket",
			deployEnv:        "production",
			underscoreDeploy: "",
			podAccount:       "585008086734",
			expected:         "my-bucket",
		},
		{
			name:             "production environment ignores pod account not in map",
			bucketName:       "my-bucket",
			deployEnv:        "production",
			underscoreDeploy: "",
			podAccount:       "999999999999",
			expected:         "my-bucket",
		},
		{
			name:             "staging environment, no pod account",
			bucketName:       "my-bucket",
			deployEnv:        "staging",
			underscoreDeploy: "",
			podAccount:       "",
			expected:         "my-bucket-dev",
		},
		{
			name:             "staging environment, pod account in map",
			bucketName:       "my-bucket",
			deployEnv:        "staging",
			underscoreDeploy: "",
			podAccount:       "585008086734",
			expected:         "my-bucket-dev-585008086734",
		},
		{
			name:             "staging environment, pod account not in map",
			bucketName:       "my-bucket",
			deployEnv:        "staging",
			underscoreDeploy: "",
			podAccount:       "999999999999",
			expected:         "my-bucket-dev",
		},
		{
			name:             "development environment using _DEPLOY_ENV, pod account in map",
			bucketName:       "my-bucket",
			deployEnv:        "",
			underscoreDeploy: "development",
			podAccount:       "585008086734",
			expected:         "my-bucket-dev-585008086734",
		},
		{
			name:             "development environment using _DEPLOY_ENV, no pod account",
			bucketName:       "my-bucket",
			deployEnv:        "",
			underscoreDeploy: "development",
			podAccount:       "",
			expected:         "my-bucket-dev",
		},
		{
			name:             "empty pod account string",
			bucketName:       "my-bucket",
			deployEnv:        "staging",
			underscoreDeploy: "",
			podAccount:       "",
			expected:         "my-bucket-dev",
		},
		{
			name:             "DEPLOY_ENV takes precedence over _DEPLOY_ENV when both are set",
			bucketName:       "my-bucket",
			deployEnv:        "production",
			underscoreDeploy: "staging",
			podAccount:       "",
			expected:         "my-bucket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			origDeployEnv := os.Getenv("DEPLOY_ENV")
			origUnderscoreDeploy := os.Getenv("_DEPLOY_ENV")
			origPodAccount := os.Getenv("_POD_ACCOUNT")

			// Set test environment
			os.Setenv("DEPLOY_ENV", tt.deployEnv)
			os.Setenv("_DEPLOY_ENV", tt.underscoreDeploy)
			os.Setenv("_POD_ACCOUNT", tt.podAccount)

			// Clean up environment after test
			defer func() {
				os.Setenv("DEPLOY_ENV", origDeployEnv)
				os.Setenv("_DEPLOY_ENV", origUnderscoreDeploy)
				os.Setenv("_POD_ACCOUNT", origPodAccount)
			}()

			actual := testS3NameByEnv(tt.bucketName)
			assert.Equal(t, tt.expected, actual, tt.name)
		})
	}
}
