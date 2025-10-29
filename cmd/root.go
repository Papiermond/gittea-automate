package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gitea-sync",
	Short: "A CLI tool to manage Gitea and GitHub repository synchronization",
	Long: `gitea-sync helps you manage repositories across Gitea and GitHub.

It allows you to:
  - Create new repositories on both platforms
  - Set up automatic push mirroring from Gitea to GitHub
  - Bulk configure existing repositories
  - Initialize repositories with common files`,
}

func Execute() error {
	return rootCmd.Execute()
}
