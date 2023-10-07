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

func (e *executor) Execute() error {
STEP:
	for _, s := range e.config.Steps {
		attemptedFixes := make([]string, len(s.Fixes))

	WHILE:
		for {
			err := executeStep(e.projectPath, &s)

			if err == nil {
				continue STEP
			}

			fmt.Printf("Step '%s' failed to run, trying fixes 🛠️\n", s.Key)
			for _, f := range s.Fixes {
				if slices.Contains(attemptedFixes, f.Key) {
					continue
				}

				result, err := executeFix(e.projectPath, &f)

				if err != nil {
					return err
				}

				switch result {
				case FixResultSuccess:
					attemptedFixes = append(attemptedFixes, f.Key)
					// try the step again
					fmt.Printf("- Trying step '%s' again 🤞\n", s.Key)
					continue WHILE
				case FixResultSkipped:
					// try the next fix
					continue
				case FixResultFailed:
					attemptedFixes = append(attemptedFixes, f.Key)
					// try the next fix
					continue
				}
			}

			// no more fixes to attempt
			return err
		}
	}
	return nil
}

type fixResult int

const (
	FixResultSuccess fixResult = iota
	FixResultSkipped fixResult = iota
	FixResultFailed  fixResult = iota
)

func executeFix(projectPath string, fix *fixConfig) (fixResult, error) {
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

func executeStep(projectPath string, step *stepConfig) error {
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
