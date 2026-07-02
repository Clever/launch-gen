package main

import (
	"bytes"
	"flag"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update golden fixture files instead of comparing against them")

func Test_generate(t *testing.T) {
	tests := []struct {
		name         string
		generate     func(string, map[string]bool, string, []byte, io.Writer) error
		input        string
		expected     string
		overrideDeps string
	}{
		{
			name:     "fargate launch1",
			generate: generateFargate,
			input:    "fixtures/launch1.yml",
			expected: "fixtures/launch1.expected",
		},
		{
			name:         "fargate launch2 with dependency override",
			generate:     generateFargate,
			input:        "fixtures/launch2.yml",
			expected:     "fixtures/launch2.expected",
			overrideDeps: "dapple:dapple/gen-go/client/v5",
		},
		{
			name:     "kubernetes values1",
			generate: generateKubernetes,
			input:    "fixtures/values1.yaml",
			expected: "fixtures/values1.expected",
		},
		{
			name:         "kubernetes values2 with dependency override",
			generate:     generateKubernetes,
			input:        "fixtures/values2.yaml",
			expected:     "fixtures/values2.expected",
			overrideDeps: "dapple:dapple/gen-go/client/v5",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			var output bytes.Buffer
			skipDependencies := map[string]bool{"dependency-to-skip": true}
			if err := tt.generate("packagename", skipDependencies, tt.overrideDeps, data, &output); err != nil {
				t.Fatal(err)
			}

			if *update {
				if err := os.WriteFile(tt.expected, output.Bytes(), 0644); err != nil {
					t.Fatal(err)
				}
				return
			}

			expected, err := os.ReadFile(tt.expected)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, string(expected), output.String())
		})
	}
}
