package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
	for _, s := range e.config.Steps {
		err := executeStep(e.projectPath, &s)

		// Step succeeded -> move onto next step
		if err == nil {
			continue
		}

		// No fixes -> exit with failure
		if len(s.Fixes) == 0 {
			return err
		}

		// Try some fixes
		fmt.Printf("Step '%s' failed to run, trying fixes...\n", s.Key)
		for _, f := range s.Fixes {
			executeFix(e.projectPath, &f)
		}

		// Try the step again
		err = executeStep(e.projectPath, &s)

		// Fixes didn't work -> exit with failure
		if err != nil {
			return err
		}
	}
	return nil
}

func executeFix(projectPath string, fix *fixConfig) error {
	cmd := exec.Command("bash", "-c", fix.Command)
	cmd.Env = os.Environ()
	cmd.Dir = projectPath

	err := cmd.Run()

	// Some errors aren't failures
	exitCode := cmd.ProcessState.ExitCode()
	if exitCode == 1 {
		fmt.Printf("Fix '%s' was skipped\n", fix.Key)
		return nil
	}
	if exitCode != 0 {
		fmt.Printf("[Warning] Fix '%s' exited with code %d\n", fix.Key, exitCode)
		return nil
	}

	// Other falure with executing command
	if err != nil {
		return fmt.Errorf("Fix '%s' failed to run: %w", fix.Key, err)
	}

	fmt.Printf("Fix '%s' ran successfully\n", fix.Key)

	return nil
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
