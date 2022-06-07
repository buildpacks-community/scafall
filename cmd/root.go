package cmd

import (
	"github.com/spf13/cobra"

	scafall "github.com/AidanDelaney/scafall/pkg"
)

const (
	outputFolderFlag = "output-folder"
)

var (
	rootCmd = &cobra.Command{
		Use:   "scafall url",
		Short: "A project generation tool",
		Long:  `Scafall creates new project from project templates.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			s := scafall.NewScafall()
			outputDir, err := cmd.Flags().GetString(outputFolderFlag)
			if err != nil {
				scafall.WithOutputFolder(outputDir)(&s)
			}

			return s.Scaffold(url)
		},
	}
)

func init() {
	rootCmd.Flags().String(outputFolderFlag, ".", "scaffold project in the provided output directory")
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
