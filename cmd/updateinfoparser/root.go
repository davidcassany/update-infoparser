package updateinfoparser

import (
	"fmt"
	"os"
	"time"

	"github.com/davidcassany/updateinfo-parser/pkg/parser"
	"github.com/spf13/cobra"
)

const securityType = "security"

var rootCmd = &cobra.Command{
	Use:   "updateinfo-parser [flags] updateinfo",
	Short: "updateinfo-parser - A simple CLI to parser updateinfo XML files",
	Long:  `A simple CLI to parser updateinfo XML files`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		updateInfo := args[0]
		if _, err := os.Stat(updateInfo); err != nil {
			return fmt.Errorf("could not fild updateinfo file '%s'", updateInfo)
		}
		flags := cmd.Flags()

		beforeStr, _ := flags.GetString("beforeDate")
		afterStr, _ := flags.GetString("afterDate")
		packagesF, _ := flags.GetString("packages")
		tmplF, _ := flags.GetString("template")
		output, _ := flags.GetString("output")
		sec, _ := flags.GetBool("security")
		var updateType string

		if sec {
			updateType = securityType
		}

		cfg, err := parser.NewConfig(updateInfo, beforeStr, afterStr, packagesF, tmplF, output, updateType)
		if err != nil {
			return err
		}

		return parser.Parse(cfg)
	},
}

func init() {
	rootCmd.Flags().StringP("beforeDate", "b", time.Now().Format(parser.DateLayout), "Filter updates released before the given date (format: 'YYYY-MM-DD'). Defaults to current date")
	rootCmd.Flags().StringP("afterDate", "a", parser.DateLayout, "Filter updates released after the given date (format: 'YYYY-MM-DD'). Defaults to '2006-01-02'")
	rootCmd.Flags().StringP("output", "o", "", "Output file. Defaults to 'stdout'")
	rootCmd.Flags().StringP("template", "t", "", "Provides a custom update template file")
	rootCmd.Flags().StringP("packages", "p", "", "Package file list to filter updates modiying any of listed packages")
	rootCmd.Flags().BoolP("security", "s", false, "Match only security updates")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Whoops. There was an error: %v\n", err)
		os.Exit(1)
	}
}
