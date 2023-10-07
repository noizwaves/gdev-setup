/*
Copyright © 2023 Adam Neumann <adam@noizwaves.com>
*/
package cmd

import (
	"os"

	"github.com/noizwaves/gdev-setup/core"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:          "gdev-setup",
	Short:        "Set up local development environment",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		workDir, err := cmd.Flags().GetString("workDir")
		if err != nil {
			return err
		}

		if workDir == "." {
			workDir, err = os.Getwd()
			if err != nil {
				return err
			}
		}

		return rootAction(workDir)
	},
}

func rootAction(workDir string) error {
	executor, err := core.NewExecutor(workDir)
	if err != nil {
		return err
	}

	return executor.Execute()
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("workDir", ".", "The application directory")
}
