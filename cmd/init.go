package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Papiermond/gitea-sync/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize gitea-sync configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("================================================")
		fmt.Println("Gitea-Sync Configuration Setup")
		fmt.Println("================================================\n")

		// Get Gitea details
		fmt.Print("Gitea URL (e.g., http://pi-nas.local:3000): ")
		giteaURL, _ := reader.ReadString('\n')
		giteaURL = strings.TrimSpace(giteaURL)

		fmt.Print("Gitea Username: ")
		giteaUsername, _ := reader.ReadString('\n')
		giteaUsername = strings.TrimSpace(giteaUsername)

		fmt.Print("Gitea Token: ")
		giteaToken, _ := reader.ReadString('\n')
		giteaToken = strings.TrimSpace(giteaToken)

		// Get GitHub details
		fmt.Print("\nGitHub Username: ")
		githubUsername, _ := reader.ReadString('\n')
		githubUsername = strings.TrimSpace(githubUsername)

		fmt.Print("GitHub Token: ")
		githubToken, _ := reader.ReadString('\n')
		githubToken = strings.TrimSpace(githubToken)

		// Get GitLab details (optional)
		fmt.Print("\nGitLab URL (press Enter to skip, default: https://gitlab.com): ")
		gitlabURL, _ := reader.ReadString('\n')
		gitlabURL = strings.TrimSpace(gitlabURL)

		var gitlabUsername, gitlabToken string
		if gitlabURL != "" {
			fmt.Print("GitLab Username: ")
			gitlabUsername, _ = reader.ReadString('\n')
			gitlabUsername = strings.TrimSpace(gitlabUsername)

			fmt.Print("GitLab Token: ")
			gitlabToken, _ = reader.ReadString('\n')
			gitlabToken = strings.TrimSpace(gitlabToken)
		}

		// Create config
		cfg := &config.Config{
			Gitea: config.GiteaConfig{
				URL:      giteaURL,
				Token:    giteaToken,
				Username: giteaUsername,
			},
			GitHub: config.GitHubConfig{
				Token:    githubToken,
				Username: githubUsername,
			},
			GitLab: config.GitLabConfig{
				URL:      gitlabURL,
				Token:    gitlabToken,
				Username: gitlabUsername,
			},
		}

		// Save config
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		path, _ := config.ConfigPath()
		fmt.Println("\n================================================")
		fmt.Println("✓ Configuration saved!")
		fmt.Printf("Config file: %s\n", path)
		fmt.Println("================================================")
		fmt.Println("\nYou can now use:")
		fmt.Println("  gitea-sync create <repo-name>    # Create new repo")
		fmt.Println("  gitea-sync mirror <repo-name>    # Add mirror to existing repo")
		fmt.Println("  gitea-sync bulk                  # Bulk setup repos")
		fmt.Println("================================================")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
