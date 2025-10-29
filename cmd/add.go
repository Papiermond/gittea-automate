package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Papiermond/gitea-sync/internal/config"
	"github.com/Papiermond/gitea-sync/internal/gitea"
	"github.com/Papiermond/gitea-sync/internal/github"
	"github.com/Papiermond/gitea-sync/internal/gitlab"
	"github.com/spf13/cobra"
)

var (
	addPrivateFlag bool
	addRepoName    string
	addUseGitLab   bool
	addUseGitHub   bool
)

var addCmd = &cobra.Command{
	Use:   "add [path]",
	Short: "Add an existing repository with code to Gitea with mirroring to GitHub or GitLab",
	Long: `Add an existing local repository to Gitea with mirroring to GitHub or GitLab.

If no path is provided, uses the current directory.
The repository name is detected from the directory name or can be specified with --name.

Examples:
  gitea-sync add                       # Add current directory (mirrors to GitHub)
  gitea-sync add ./my-project          # Add specific directory (mirrors to GitHub)
  gitea-sync add --name custom-name    # Add current dir with custom name
  gitea-sync add --gitlab              # Add and mirror to GitLab instead`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate flags
		if addUseGitLab && addUseGitHub {
			return fmt.Errorf("cannot use both -gitlab and -github flags")
		}

		// Default to GitHub if neither flag is set
		if !addUseGitLab && !addUseGitHub {
			addUseGitHub = true
		}

		// Determine the path
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		// Get absolute path
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}

		// Check if it's a git repository
		if !isGitRepo(absPath) {
			return fmt.Errorf("not a git repository: %s", absPath)
		}

		// Determine repo name
		repoName := addRepoName
		if repoName == "" {
			repoName = filepath.Base(absPath)
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Validate configuration based on selected platform
		if addUseGitLab {
			if cfg.GitLab.Token == "" || cfg.GitLab.Username == "" {
				return fmt.Errorf("GitLab credentials not configured. Run 'gitea-sync init' to configure")
			}
		}

		// Initialize clients
		giteaClient := gitea.NewClient(cfg.Gitea.URL, cfg.Gitea.Token)
		var githubClient *github.Client
		var gitlabClient *gitlab.Client

		if addUseGitHub {
			githubClient = github.NewClient(cfg.GitHub.Token)
		} else {
			gitlabClient = gitlab.NewClient(cfg.GitLab.URL, cfg.GitLab.Token)
		}

		fmt.Println("================================================")
		fmt.Printf("Adding repository: %s\n", repoName)
		fmt.Printf("Path: %s\n", absPath)
		fmt.Printf("Privacy setting: %t\n", addPrivateFlag)
		if addUseGitHub {
			fmt.Println("Mirror target: GitHub")
		} else {
			fmt.Println("Mirror target: GitLab")
		}
		fmt.Println("================================================")

		// 1. Create on GitHub or GitLab
		var exists bool
		if addUseGitHub {
			fmt.Println("\n1. Checking GitHub...")
			exists, err = githubClient.RepoExists(cfg.GitHub.Username, repoName)
			if err != nil {
				return fmt.Errorf("failed to check GitHub: %w", err)
			}

			if !exists {
				fmt.Println("  → Creating GitHub repo...")
				err = githubClient.CreateRepo(github.CreateRepoRequest{
					Name:     repoName,
					Private:  addPrivateFlag,
					AutoInit: false,
				})
				if err != nil {
					return fmt.Errorf("failed to create GitHub repo: %w", err)
				}
				fmt.Println("  ✓ GitHub repo created")
			} else {
				fmt.Println("  ✓ GitHub repo already exists")
			}
		} else {
			fmt.Println("\n1. Checking GitLab...")
			exists, err = gitlabClient.RepoExists(cfg.GitLab.Username, repoName)
			if err != nil {
				return fmt.Errorf("failed to check GitLab: %w", err)
			}

			if !exists {
				fmt.Println("  → Creating GitLab repo...")
				visibility := "public"
				if addPrivateFlag {
					visibility = "private"
				}
				err = gitlabClient.CreateRepo(gitlab.CreateRepoRequest{
					Name:       repoName,
					Visibility: visibility,
				})
				if err != nil {
					return fmt.Errorf("failed to create GitLab repo: %w", err)
				}
				fmt.Println("  ✓ GitLab repo created")
			} else {
				fmt.Println("  ✓ GitLab repo already exists")
			}
		}

		// 2. Create on Gitea
		fmt.Println("\n2. Checking Gitea...")
		exists, err = giteaClient.RepoExists(cfg.Gitea.Username, repoName)
		if err != nil {
			return fmt.Errorf("failed to check Gitea: %w", err)
		}

		if !exists {
			fmt.Println("  → Creating Gitea repo...")
			err = giteaClient.CreateRepo(gitea.CreateRepoRequest{
				Name:     repoName,
				Private:  addPrivateFlag,
				AutoInit: false,
			})
			if err != nil {
				return fmt.Errorf("failed to create Gitea repo: %w", err)
			}
			fmt.Println("  ✓ Gitea repo created")
		} else {
			fmt.Println("  ✓ Gitea repo already exists")
		}

		// 3. Set up push mirror
		if addUseGitHub {
			fmt.Println("\n3. Setting up GitHub mirror...")
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
		} else {
			fmt.Println("\n3. Setting up GitLab mirror...")
			gitlabURL := cfg.GitLab.URL
			if gitlabURL == "" {
				gitlabURL = "https://gitlab.com"
			}
			err = giteaClient.AddPushMirror(cfg.Gitea.Username, repoName, gitea.PushMirrorRequest{
				RemoteAddress:  fmt.Sprintf("%s/%s/%s.git", gitlabURL, cfg.GitLab.Username, repoName),
				RemotePassword: cfg.GitLab.Token,
				RemoteUsername: cfg.GitLab.Username,
				SyncOnCommit:   true,
				Interval:       "8h",
			})
			if err != nil {
				return fmt.Errorf("failed to set up mirror: %w", err)
			}
			fmt.Println("  ✓ Mirror configured")
		}

		// 4. Set up git remote and push
		fmt.Println("\n4. Configuring git remote...")
		if err := setupGitRemote(absPath, repoName, cfg); err != nil {
			return err
		}

		// 5. Done!
		fmt.Println("\n================================================")
		fmt.Println("✓ Repository successfully added!")
		fmt.Println("================================================")
		fmt.Printf("\nRepository URLs:\n")
		fmt.Printf("  Gitea:  %s/%s/%s\n", cfg.Gitea.URL, cfg.Gitea.Username, repoName)
		if addUseGitHub {
			fmt.Printf("  GitHub: https://github.com/%s/%s\n", cfg.GitHub.Username, repoName)
		} else {
			gitlabURL := cfg.GitLab.URL
			if gitlabURL == "" {
				gitlabURL = "https://gitlab.com"
			}
			fmt.Printf("  GitLab: %s/%s/%s\n", gitlabURL, cfg.GitLab.Username, repoName)
		}
		fmt.Println("\nYour local repository is now:")
		fmt.Println("  • Connected to Gitea as 'origin'")
		if addUseGitHub {
			fmt.Println("  • Mirroring to GitHub automatically")
		} else {
			fmt.Println("  • Mirroring to GitLab automatically")
		}
		fmt.Println("  • Ready for commits")
		fmt.Println("================================================")

		return nil
	},
}

func isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func setupGitRemote(repoPath, repoName string, cfg *config.Config) error {
	// Check if 'origin' remote exists
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoPath
	output, err := cmd.Output()

	giteaRemoteURL := fmt.Sprintf("http://%s:%s@%s/%s/%s.git",
		cfg.Gitea.Username,
		cfg.Gitea.Token,
		strings.TrimPrefix(cfg.Gitea.URL, "http://"),
		cfg.Gitea.Username,
		repoName)

	if err != nil {
		// No origin remote exists, add it
		fmt.Println("  → Adding Gitea as origin remote...")
		cmd = exec.Command("git", "remote", "add", "origin", giteaRemoteURL)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to add remote: %w", err)
		}
		fmt.Println("  ✓ Remote added")
	} else {
		// Origin exists, check if it's Gitea
		existingRemote := strings.TrimSpace(string(output))
		if !strings.Contains(existingRemote, cfg.Gitea.URL) {
			// Origin points elsewhere, add Gitea as 'gitea' remote
			fmt.Printf("  ℹ Origin exists (%s)\n", existingRemote)
			fmt.Println("  → Adding Gitea as 'gitea' remote...")

			// Remove gitea remote if it exists
			exec.Command("git", "remote", "remove", "gitea").Run()

			cmd = exec.Command("git", "remote", "add", "gitea", giteaRemoteURL)
			cmd.Dir = repoPath
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to add gitea remote: %w", err)
			}
			fmt.Println("  ✓ Remote 'gitea' added")

			fmt.Println("\n  → Pushing to Gitea...")
			cmd = exec.Command("git", "push", "-u", "gitea", "main")
			cmd.Dir = repoPath
			if err := cmd.Run(); err != nil {
				// Try 'master' if 'main' fails
				cmd = exec.Command("git", "push", "-u", "gitea", "master")
				cmd.Dir = repoPath
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to push (tried both 'main' and 'master'): %w", err)
				}
			}
			fmt.Println("  ✓ Pushed to Gitea (use 'git push gitea' in the future)")
			return nil
		}

		// Origin is already Gitea, update it
		fmt.Println("  → Updating origin URL...")
		cmd = exec.Command("git", "remote", "set-url", "origin", giteaRemoteURL)
		cmd.Dir = repoPath
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update remote: %w", err)
		}
		fmt.Println("  ✓ Remote updated")
	}

	// Get current branch
	cmd = exec.Command("git", "branch", "--show-current")
	cmd.Dir = repoPath
	branchOutput, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}
	branch := strings.TrimSpace(string(branchOutput))
	if branch == "" {
		branch = "main" // Default if detached HEAD or other issue
	}

	// Push to origin
	fmt.Printf("  → Pushing to Gitea (branch: %s)...\n", branch)
	cmd = exec.Command("git", "push", "-u", "origin", branch)
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	fmt.Println("  ✓ Pushed to Gitea")
	fmt.Println("  ✓ Mirroring to GitHub...")

	return nil
}

func init() {
	addCmd.Flags().BoolVarP(&addPrivateFlag, "private", "p", false, "Make the repository private")
	addCmd.Flags().StringVarP(&addRepoName, "name", "n", "", "Custom repository name (defaults to directory name)")
	addCmd.Flags().BoolVar(&addUseGitLab, "gitlab", false, "Mirror to GitLab instead of GitHub")
	addCmd.Flags().BoolVar(&addUseGitHub, "github", false, "Mirror to GitHub (default)")
	rootCmd.AddCommand(addCmd)
}
