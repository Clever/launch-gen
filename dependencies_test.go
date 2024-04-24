package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDependencies(t *testing.T) {
	fakeConfig := &LaunchYML{
		Dependencies: []string{
			"service-a",
			"service-b",
			"service-c@v4",
			"service-d",
		},
	}

	skip := flagsSet{}
	skip.Set("service-b")
	result := parseDependencies(fakeConfig, skip, map[string]string{
		"service-d": "service-d/custom/package/path",
	})

	assert.Equal(t, []ServiceDependency{
		{
			name: "service-a",
		},
		{
			name:    "service-c",
			version: "v4",
		},
		{
			name:     "service-d",
			override: "service-d/custom/package/path",
		},
	}, result)
}
