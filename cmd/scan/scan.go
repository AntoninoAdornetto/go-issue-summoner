/*
Copyright © 2024 Antonino Adornetto

The scan command processes each source file individually and searches for specific tags (actionable comments) that the user specifies.
It respects the `.gitignore` settings and ensures that any files designated as ignored are not scanned.
Finally, a detailed report is presented to the user about the tags that were found during the scan.
*/
package scan

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/AntoninoAdornetto/issue-summoner/pkg/tag"
	"github.com/spf13/cobra"
)

type ScanManager struct{}

func (ScanManager) Open(fileName string) (*os.File, error) {
	return os.Open(fileName)
}

var ScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scans source code file(s) and searches for actionable comments",
	Long:  `Scans a local git respository for Tags (actionable comments) and reports findings to the console.`,
	Run: func(cmd *cobra.Command, _ []string) {
		path, err := cmd.Flags().GetString("path")
		if err != nil {
			log.Fatalf("Failed to read 'path' flag: %s", err)
		}

		gitIgnorePath, err := cmd.Flags().GetString("gitignorePath")
		if err != nil {
			log.Fatalf("Failed to read 'gitignorePath' flag\n%s", err)
		}

		if path == "" {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get working directory\n%s", err)
			}
			path = wd
		}

		if gitIgnorePath == "" {
			gitIgnorePath = filepath.Join(path, tag.GitIgnoreFile)
		}

		scanManager := ScanManager{}
		ignorePatterns, err := tag.ProcessIgnorePatterns(gitIgnorePath, scanManager)
		if err != nil {
			log.Fatal(err)
		}

		for _, p := range ignorePatterns {
			fmt.Printf("Ignore Pattern: %s\n", p.String())
		}
	},
}

func init() {
	ScanCmd.Flags().StringP("path", "p", "", "Path to your local git project.")
	ScanCmd.Flags().StringP("tag", "t", "@TODO", "Actionable comment tag to search for.")
	ScanCmd.Flags().StringP("mode", "m", "P", "Mode: 'I' (Issued) or 'P' (Pending).")
	ScanCmd.Flags().StringP("gitignorePath", "g", "", "Path to .gitignore file.")
}
