package main

import (
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
