package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"time"

	"gopkg.in/yaml.v3"
)

// ***
// Interface
// ***
type Executor interface {
	Execute() error
}

// ***
// Config
// ***
type setupConfig struct {
	Steps []stepConfig `yaml:"steps"`
}

type stepConfig struct {
	Key         string             `yaml:"key"`
	Command     string             `yaml:"command"`
	Fixes       []fixConfig        `yaml:"fixes,omitempty"`
	KnownIssues []knownIssueConfig `yaml:"known-issues,omitempty"`
}

type fixConfig struct {
	Key     string `yaml:"key"`
	Command string `yaml:"command"`
}

type knownIssueConfig struct {
	Key      string `yaml:"key"`
	Problem  string `yaml:"problem"`
	Solution string `yaml:"solution"`
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

// ***
// Implementation
// ***
type executor struct {
	config  *setupConfig
	context *executionContext
}

type executionContext struct {
	ProjectPath   string
	LogOutputPath string
}

func NewExecutor(projectPath string) (Executor, error) {
	configPath := filepath.Join(projectPath, ".gdev", "gdev.setup.yaml")
	config, err := parseConfig(configPath)
	if err != nil {
		return nil, err
	}

	logOutputPath, err := os.MkdirTemp("", "gdev-setup")
	if err != nil {
		return nil, err
	}

	return &executor{
		config: config,
		context: &executionContext{
			ProjectPath:   projectPath,
			LogOutputPath: logOutputPath,
		},
	}, nil
}

func (e *executor) Execute() error {
	for _, s := range e.config.Steps {
		err := executeStep(e.context, &s, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

// ***
// Steps
// ***
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

func executeStep(context *executionContext, step *stepConfig, state *stepState) error {
	if state == nil {
		state = NewStepState(step)
	}

	result, err := executeStepCommand(context, step)
	// error with command preparation
	if err != nil {
		return err
	}

	// step command exited successfully
	if result.RuntimeError == nil {
		return nil
	}

	// step command did not exit successfully
	err = result.RuntimeError

	fmt.Printf("Step '%s' failed to run, trying fixes 🛠️\n", step.Key)
	for _, f := range step.Fixes {
		if state.WasAttempted(&f) {
			continue
		}

		result, err := executeFixCommand(context, &f, result.LogFilePath)
		if err != nil {
			return err
		}

		switch result {
		case FixResultSuccess:
			state.AddAttempt(&f)
			fmt.Printf("- Trying step '%s' again 🤞\n", step.Key)
			return executeStep(context, step, state)
		case FixResultSkipped:
			// try the next fix
			continue
		case FixResultFailed:
			state.AddAttempt(&f)
			// try the next fix
			continue
		}
	}

	// no more fixes to attempt, fall back to known issues
	showKnownIssues(step)

	return err
}

type stepExecResult struct {
	LogFilePath  string
	RuntimeError error
}

func executeStepCommand(context *executionContext, step *stepConfig) (*stepExecResult, error) {
	cmd := exec.Command("bash", "-c", step.Command)
	cmd.Env = os.Environ()
	cmd.Dir = context.ProjectPath

	// Capture output
	logFileName := fmt.Sprintf("%d-%s.log", time.Now().UnixMilli(), step.Key)
	logFilePath := filepath.Join(context.LogOutputPath, logFileName)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("Step '%s' failed to create output log file: %s", step.Key, err)
	}
	defer logFile.Close()
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	err = cmd.Run()
	if err == nil {
		fmt.Printf("Step '%s' ran successfully\n", step.Key)
	}

	return &stepExecResult{
		LogFilePath:  logFilePath,
		RuntimeError: err,
	}, nil
}

// ***
// Fixes
// ***
type fixResult int

const (
	FixResultSuccess fixResult = iota
	FixResultSkipped fixResult = iota
	FixResultFailed  fixResult = iota
)

func executeFixCommand(context *executionContext, fix *fixConfig, stepLogPath string) (fixResult, error) {
	cmd := exec.Command("bash", "-c", fix.Command)
	cmd.Env = os.Environ()
	cmd.Dir = context.ProjectPath
	cmd.Env = append(cmd.Env, "STEP_LOG_PATH="+stepLogPath)

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

// ***
// Known Issues
// ***
func showKnownIssues(step *stepConfig) {
	if len(step.KnownIssues) == 0 {
		return
	}

	fmt.Printf("Step '%s' has the following known issues:\n", step.Key)

	for i, ki := range step.KnownIssues {
		fmt.Printf("Problem (%d): %s\nSolution (%d): %s\n", i+1, ki.Problem, i+1, ki.Solution)
	}
}
