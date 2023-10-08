package main

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testProgram = "./gdev-setup-testable"

func TestExamples(t *testing.T) {
	buildGdevSetup(t)

	t.Run("HappyPath", func(t *testing.T) {
		exampleDir, err := filepath.Abs("testdata/happy-path")
		if err != nil {
			t.Fatalf("example directory path error: %s", err)
		}

		output := runGdevSetup(t, "--workDir", exampleDir)

		expected := `Step 'foo' ran successfully
Step 'bar' ran successfully
Step 'baz' failed to run, trying fixes 🛠️
- Fix 'always-skipped' was skipped
- Fix 'always-fails' failed with exit code 2
- Fix 'touch-baz1' ran successfully
- Trying step 'baz' again 🤞
Step 'baz' failed to run, trying fixes 🛠️
- Fix 'always-skipped' was skipped
- Fix 'touch-baz2' ran successfully
- Trying step 'baz' again 🤞
Step 'baz' ran successfully
`

		assert.Equal(t, expected, output)
	})
}

func runGdevSetup(t *testing.T, command ...string) string {
	t.Helper()

	cmd := exec.Command(testProgram, command...)
	stdout := bytes.Buffer{}
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		t.Fatalf("gdev-setup encountered an error: %s", err)
	}

	return stdout.String()
}

func buildGdevSetup(t *testing.T) {
	t.Helper()

	cmd := exec.Command("go", "build", "-o", testProgram, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build gdev-setup: %s", err)
	}
}
