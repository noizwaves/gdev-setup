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
		exampleDir := buildExampleDir(t, "happy-path")

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

	t.Run("SomeKnownIssues", func(t *testing.T) {
		exampleDir := buildExampleDir(t, "some-known-issues")

		output := runGdevSetup(t, "--workDir", exampleDir)

		expected := `Step 'foo' failed to run, trying fixes 🛠️
Step 'foo' has the following known issues:
Problem (1): the foo.txt file is missing
Solution (1): open foo.txt in your IDE and populate it
Problem (2): cosmic ray hits ssd, flips an important bit
Solution (2): hope another cosmic ray flips it back
`

		assert.Equal(t, expected, output)
	})
}

func buildExampleDir(t *testing.T, name string) string {
	examplePath := filepath.Join("testdata", name)
	absPath, err := filepath.Abs(examplePath)
	if err != nil {
		t.Fatalf("example directory absolute error: %s", err)
	}

	return absPath
}

func runGdevSetup(t *testing.T, command ...string) string {
	t.Helper()

	cmd := exec.Command(testProgram, command...)
	stdout := bytes.Buffer{}
	cmd.Stdout = &stdout

	cmd.Run()

	return stdout.String()
}

func buildGdevSetup(t *testing.T) {
	t.Helper()

	cmd := exec.Command("go", "build", "-o", testProgram, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build gdev-setup: %s", err)
	}
}
