package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"gopkg.in/yaml.v3"
)

/*
 * Interface
 */
type Executor interface {
	Execute() error
}

/*
 * Config
 */
type setupConfig struct {
	Steps []stepConfig `yaml:"steps"`
}

type stepConfig struct {
	Key     string      `yaml:"key"`
	Command string      `yaml:"command"`
	Fixes   []fixConfig `yaml:"fixes,omitempty"`
}

type fixConfig struct {
	Key     string `yaml:"key"`
	Command string `yaml:"command"`
}

func parseConfig(configPath string) (*setupConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var result setupConfig
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	// TODO: schema validation goes here

	return &result, nil
}

/*
 * Implementation
 */
type executor struct {
	config      *setupConfig
	projectPath string
}

func NewExecutor(projectPath string) (Executor, error) {
	configPath := filepath.Join(projectPath, ".gdev", "gdev.setup.yaml")
	config, err := parseConfig(configPath)
	if err != nil {
		return nil, err
	}

	return &executor{
		config:      config,
		projectPath: projectPath,
	}, nil
}

func (e *executor) Execute() error {
	for _, s := range e.config.Steps {
		err := executeStep(e.projectPath, &s, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
 * Steps
 */

type stepState struct {
	attemptedFixes []string
}

func (s *stepState) AddAttempt(f *fixConfig) {
	s.attemptedFixes = append(s.attemptedFixes, f.Key)
}

func (s *stepState) WasAttempted(f *fixConfig) bool {
	return slices.Contains(s.attemptedFixes, f.Key)
}

func NewStepState(step *stepConfig) *stepState {
	return &stepState{
		attemptedFixes: make([]string, len(step.Fixes)),
	}
}

func executeStep(projectPath string, step *stepConfig, state *stepState) error {
	if state == nil {
		state = NewStepState(step)
	}

	err := executeStepCommand(projectPath, step)
	if err == nil {
		return nil
	}

	fmt.Printf("Step '%s' failed to run, trying fixes 🛠️\n", step.Key)
	for _, f := range step.Fixes {
		if state.WasAttempted(&f) {
			continue
		}

		result, err := executeFixCommand(projectPath, &f)
		if err != nil {
			return err
		}

		switch result {
		case FixResultSuccess:
			state.AddAttempt(&f)
			fmt.Printf("- Trying step '%s' again 🤞\n", step.Key)
			return executeStep(projectPath, step, state)
		case FixResultSkipped:
			// try the next fix
			continue
		case FixResultFailed:
			state.AddAttempt(&f)
			// try the next fix
			continue
		}
	}

	// no more fixes to attempt
	return err
}

func executeStepCommand(projectPath string, step *stepConfig) error {
	cmd := exec.Command("bash", "-c", step.Command)
	cmd.Env = os.Environ()
	cmd.Dir = projectPath

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Step '%s' failed to run: %w", step.Key, err)
	}

	if exitCode := cmd.ProcessState.ExitCode(); exitCode != 0 {
		return fmt.Errorf("Step '%s' exited with code %d", step.Key, exitCode)
	}

	fmt.Printf("Step '%s' ran successfully\n", step.Key)

	return nil
}

/*
 * Fixes
 */
type fixResult int

const (
	FixResultSuccess fixResult = iota
	FixResultSkipped fixResult = iota
	FixResultFailed  fixResult = iota
)

func executeFixCommand(projectPath string, fix *fixConfig) (fixResult, error) {
	cmd := exec.Command("bash", "-c", fix.Command)
	cmd.Env = os.Environ()
	cmd.Dir = projectPath

	err := cmd.Run()

	// Some errors aren't fix failures
	exitCode := cmd.ProcessState.ExitCode()
	if exitCode == 1 {
		fmt.Printf("- Fix '%s' was skipped\n", fix.Key)
		return FixResultSkipped, nil
	}
	if exitCode != 0 {
		fmt.Printf("- Fix '%s' failed with exit code %d\n", fix.Key, exitCode)
		return FixResultFailed, nil
	}

	// Other falure with executing command
	if err != nil {
		return 0, fmt.Errorf("- Fix '%s' failed to run: %w", fix.Key, err)
	}

	fmt.Printf("- Fix '%s' ran successfully\n", fix.Key)

	return FixResultSuccess, nil
}
