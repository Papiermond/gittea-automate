package cmd

import (
	"fmt"

	"github.com/Papiermond/gitea-sync/internal/config"
	"github.com/Papiermond/gitea-sync/internal/gitea"
	"github.com/spf13/cobra"
)

var mirrorCmd = &cobra.Command{
	Use:   "mirror <repo-name>",
	Short: "Add GitHub push mirror to an existing Gitea repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Initialize client
		giteaClient := gitea.NewClient(cfg.Gitea.URL, cfg.Gitea.Token)

		fmt.Println("================================================")
		fmt.Printf("Setting up mirror for: %s\n", repoName)
		fmt.Println("================================================")

		// Check if repo exists
		fmt.Println("\nChecking Gitea repository...")
		exists, err := giteaClient.RepoExists(cfg.Gitea.Username, repoName)
		if err != nil {
			return fmt.Errorf("failed to check Gitea: %w", err)
		}

		if !exists {
			return fmt.Errorf("repository does not exist on Gitea")
		}
		fmt.Println("  ✓ Repository found")

		// Set up push mirror
		fmt.Println("\nSetting up GitHub mirror...")
		err = giteaClient.AddPushMirror(cfg.Gitea.Username, repoName, gitea.PushMirrorRequest{
			RemoteAddress:  fmt.Sprintf("https://github.com/%s/%s.git", cfg.GitHub.Username, repoName),
			RemotePassword: cfg.GitHub.Token,
			RemoteUsername: cfg.GitHub.Username,
			SyncOnCommit:   true,
			Interval:       "8h",
		})
		if err != nil {
			return fmt.Errorf("failed to set up mirror: %w", err)
		}
		fmt.Println("  ✓ Mirror configured")

		fmt.Println("\n================================================")
		fmt.Println("✓ Mirror setup complete!")
		fmt.Println("================================================")
		fmt.Printf("\nRepository URLs:\n")
		fmt.Printf("  Gitea:  %s/%s/%s\n", cfg.Gitea.URL, cfg.Gitea.Username, repoName)
		fmt.Printf("  GitHub: https://github.com/%s/%s\n", cfg.GitHub.Username, repoName)
		fmt.Println("\nThe repository will sync on every commit and every 8 hours.")
		fmt.Println("================================================")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(mirrorCmd)
}
