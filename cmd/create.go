package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/Papiermond/gitea-sync/internal/config"
	"github.com/Papiermond/gitea-sync/internal/gitea"
	"github.com/Papiermond/gitea-sync/internal/github"
	"github.com/Papiermond/gitea-sync/internal/gitlab"
	"github.com/spf13/cobra"
)

var (
	privateFlag bool
	useGitLab   bool
	useGitHub   bool
)

var createCmd = &cobra.Command{
	Use:   "create <repo-name>",
	Short: "Create a new repository on Gitea with mirroring to GitHub or GitLab",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoName := args[0]

		// Validate flags
		if useGitLab && useGitHub {
			return fmt.Errorf("cannot use both -gitlab and -github flags")
		}

		// Default to GitHub if neither flag is set
		if !useGitLab && !useGitHub {
			useGitHub = true
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		// Validate configuration based on selected platform
		if useGitLab {
			if cfg.GitLab.Token == "" || cfg.GitLab.Username == "" {
				return fmt.Errorf("GitLab credentials not configured. Run 'gitea-sync init' to configure")
			}
		}

		// Initialize clients
		giteaClient := gitea.NewClient(cfg.Gitea.URL, cfg.Gitea.Token)
		var githubClient *github.Client
		var gitlabClient *gitlab.Client

		if useGitHub {
			githubClient = github.NewClient(cfg.GitHub.Token)
		} else {
			gitlabClient = gitlab.NewClient(cfg.GitLab.URL, cfg.GitLab.Token)
		}

		fmt.Println("================================================")
		fmt.Printf("Creating repository: %s\n", repoName)
		fmt.Printf("Privacy setting: %t\n", privateFlag)
		if useGitHub {
			fmt.Println("Mirror target: GitHub")
		} else {
			fmt.Println("Mirror target: GitLab")
		}
		fmt.Println("================================================")

		// 1. Create on GitHub or GitLab
		var exists bool
		if useGitHub {
			fmt.Println("\n1. Checking GitHub...")
			exists, err = githubClient.RepoExists(cfg.GitHub.Username, repoName)
			if err != nil {
				return fmt.Errorf("failed to check GitHub: %w", err)
			}

			if !exists {
				fmt.Println("  → Creating GitHub repo...")
				err = githubClient.CreateRepo(github.CreateRepoRequest{
					Name:     repoName,
					Private:  privateFlag,
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
				if privateFlag {
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
				Private:  privateFlag,
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
		if useGitHub {
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

		// 4. Initialize repo
		fmt.Println("\n4. Initializing repository...")
		tempDir, err := os.MkdirTemp("", "gitea-sync-*")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		if err := initRepo(tempDir, repoName, cfg); err != nil {
			return err
		}

		// 5. Pull the repo locally
		fmt.Println("\n5. Pulling repository to current directory...")
		if err := pullRepo(repoName, cfg); err != nil {
			return err
		}

		// 6. Done!
		fmt.Println("\n================================================")
		fmt.Println("✓ Repository fully initialized and ready!")
		fmt.Println("================================================")
		fmt.Printf("\nRepository URLs:\n")
		fmt.Printf("  Gitea:  %s/%s/%s\n", cfg.Gitea.URL, cfg.Gitea.Username, repoName)
		if useGitHub {
			fmt.Printf("  GitHub: https://github.com/%s/%s\n", cfg.GitHub.Username, repoName)
		} else {
			gitlabURL := cfg.GitLab.URL
			if gitlabURL == "" {
				gitlabURL = "https://gitlab.com"
			}
			fmt.Printf("  GitLab: %s/%s/%s\n", gitlabURL, cfg.GitLab.Username, repoName)
		}
		fmt.Printf("\nLocal directory: ./%s\n", repoName)
		fmt.Println("\nThe repo is initialized with:")
		fmt.Println("  • README.md")
		fmt.Println("  • .gitignore")
		fmt.Println("  • Initial commit")
		fmt.Println("================================================")

		return nil
	},
}

func initRepo(tempDir, repoName string, cfg *config.Config) error {
	// Initialize git
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to init git: %w", err)
	}
	fmt.Println("  ✓ Git initialized")

	// Create README
	readme := fmt.Sprintf("# %s\n\nRepository created on %s\n", repoName, time.Now().Format("2006-01-02"))
	if err := os.WriteFile(filepath.Join(tempDir, "README.md"), []byte(readme), 0644); err != nil {
		return err
	}
	fmt.Println("  ✓ README.md created")

	// Create .gitignore
	gitignore := `# Common ignores
.DS_Store
*.log
node_modules/
__pycache__/
*.pyc
.env
`
	if err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return err
	}
	fmt.Println("  ✓ .gitignore created")

	// Initial commit
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	fmt.Println("  ✓ Initial commit created")

	// Push to Gitea
	remoteURL := fmt.Sprintf("http://%s:%s@%s/%s/%s.git",
		cfg.Gitea.Username,
		cfg.Gitea.Token,
		cfg.Gitea.URL[7:], // Remove http://
		cfg.Gitea.Username,
		repoName)

	cmd = exec.Command("git", "remote", "add", "origin", remoteURL)
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add remote: %w", err)
	}

	cmd = exec.Command("git", "push", "-u", "origin", "main")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	fmt.Println("  ✓ Pushed to Gitea")
	fmt.Println("  ✓ Mirroring to GitHub...")
	time.Sleep(2 * time.Second) // Give mirror time to sync

	return nil
}

func pullRepo(repoName string, cfg *config.Config) error {
	// Clone the repo to current directory
	cloneURL := fmt.Sprintf("%s/%s/%s.git", cfg.Gitea.URL, cfg.Gitea.Username, repoName)

	cmd := exec.Command("git", "clone", cloneURL)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}
	fmt.Printf("  ✓ Repository cloned to ./%s\n", repoName)

	return nil
}

func init() {
	createCmd.Flags().BoolVarP(&privateFlag, "private", "p", false, "Make the repository private")
	createCmd.Flags().BoolVar(&useGitLab, "gitlab", false, "Mirror to GitLab instead of GitHub")
	createCmd.Flags().BoolVar(&useGitHub, "github", false, "Mirror to GitHub (default)")
	rootCmd.AddCommand(createCmd)
}
