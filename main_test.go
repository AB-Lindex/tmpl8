package main

import (
	"bytes"
	"testing"
)

func TestPrepareUsesConfiguredDelimiters(t *testing.T) {
	originalArgs := args
	t.Cleanup(func() {
		args = originalArgs
	})

	args.LeftDelimiter = "[["
	args.RightDelimiter = "]]"

	templates, err := prepare([]entry{{name: "custom", data: "Hello, [[ .name ]]!"}})
	if err != nil {
		t.Fatalf("prepare returned an error: %v", err)
	}

	var output bytes.Buffer
	if err := process(templates, map[string]string{"name": "world"}, &output); err != nil {
		t.Fatalf("process returned an error: %v", err)
	}

	if got, want := output.String(), "Hello, world!\n"; got != want {
		t.Errorf("output = %q, want %q", got, want)
	}
}
