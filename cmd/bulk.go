package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Papiermond/gitea-sync/internal/config"
	"github.com/Papiermond/gitea-sync/internal/gitea"
	"github.com/spf13/cobra"
)

var bulkCmd = &cobra.Command{
	Use:   "bulk",
	Short: "Bulk setup mirrors for multiple repositories",
	Long: `Bulk setup mirrors for multiple repositories.
You will be prompted to enter repository names, one per line.
Press Ctrl+D (EOF) when done.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Get repository list
		fmt.Println("Enter repository names (one per line, Ctrl+D when done):")
		var repos []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				repos = append(repos, line)
			}
		}

		if len(repos) == 0 {
			return fmt.Errorf("no repositories provided")
		}

		// Initialize client
		giteaClient := gitea.NewClient(cfg.Gitea.URL, cfg.Gitea.Token)

		fmt.Println("\n================================================")
		fmt.Printf("Processing %d repositories\n", len(repos))
		fmt.Println("================================================")

		successCount := 0
		for _, repoName := range repos {
			fmt.Printf("\n================================================\n")
			fmt.Printf("Processing: %s\n", repoName)

			// Check if repo exists in Gitea
			exists, err := giteaClient.RepoExists(cfg.Gitea.Username, repoName)
			if err != nil {
				fmt.Printf("  ✗ Error checking repo: %v\n", err)
				continue
			}

			if !exists {
				// Create repo
				fmt.Println("  → Creating Gitea repo...")
				err = giteaClient.CreateRepo(gitea.CreateRepoRequest{
					Name:     repoName,
					Private:  false,
					AutoInit: false,
				})
				if err != nil {
					fmt.Printf("  ✗ Failed to create repo: %v\n", err)
					continue
				}
				fmt.Println("  ✓ Gitea repo created")
			} else {
				fmt.Println("  ✓ Gitea repo already exists")
			}

			// Add push mirror
			fmt.Println("  → Setting up GitHub mirror...")
			err = giteaClient.AddPushMirror(cfg.Gitea.Username, repoName, gitea.PushMirrorRequest{
				RemoteAddress:  fmt.Sprintf("https://github.com/%s/%s.git", cfg.GitHub.Username, repoName),
				RemotePassword: cfg.GitHub.Token,
				RemoteUsername: cfg.GitHub.Username,
				SyncOnCommit:   true,
				Interval:       "8h",
			})
			if err != nil {
				fmt.Printf("  ✗ Mirror setup failed: %v\n", err)
				continue
			}
			fmt.Println("  ✓ Mirror configured")
			fmt.Printf("  ✓ %s complete!\n", repoName)
			successCount++
		}

		fmt.Println("\n================================================")
		fmt.Printf("✓ Completed %d/%d repositories\n", successCount, len(repos))
		fmt.Println("================================================")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(bulkCmd)
}
