package cmd

import (
	"github.com/spf13/cobra"

	scafall "github.com/AidanDelaney/scafall/pkg"
)

const (
	outputFolderFlag = "path"
	overrideFlag     = "override"
	subPath          = "sub-path"
)

var (
	rootCmd = &cobra.Command{
		Use:   "scafall gitRepository",
		Short: "A project generation tool",
		Long:  `Scafall creates new project from project templates.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url := args[0]

			s := scafall.NewScafall()
			outputDir, err := cmd.Flags().GetString(outputFolderFlag)
			if err == nil {
				scafall.WithOutputFolder(outputDir)(&s)
			}
			overrides, err := cmd.Flags().GetStringToString(overrideFlag)
			if err == nil {
				scafall.WithOverrides(overrides)(&s)
			}
			subPath, err := cmd.Flags().GetString(subPath)
			if err == nil {
				scafall.WithSubPath(subPath)(&s)
			}

			return s.Scaffold(url)
		},
	}
)

func init() {
	rootCmd.Flags().StringP(outputFolderFlag, "p", ".", "scaffold project in the provided output directory")
	rootCmd.Flags().StringToStringP(overrideFlag, "o", map[string]string{}, "provide overrides as key-value pairs")
	rootCmd.Flags().StringP(subPath, "s", "", "use sub directory in template project ot scaffold project")
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
