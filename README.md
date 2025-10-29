# gitea-sync

A CLI tool to manage Gitea repository synchronization with GitHub or GitLab.

## Features

- **Create new repositories** on Gitea with automatic mirroring to GitHub or GitLab
- **Add existing repositories** with code to Gitea with mirroring to GitHub or GitLab
- **Add push mirrors** to existing Gitea repositories
- **Bulk setup** multiple repositories at once
- **Secure credential management** via config file
- **Auto-pull** newly created repos to your local machine
- **Support for both GitHub and GitLab** as mirror targets

## Installation

### Build from source

```bash
go build -o gitea-sync
```

### Install globally (optional)

```bash
sudo cp gitea-sync /usr/local/bin/
```

## Configuration

Before using gitea-sync, you need to initialize it with your credentials:

```bash
./gitea-sync init
```

This will prompt you for:
- Gitea URL (e.g., `http://pi-nas.local:3000`)
- Gitea username and token
- GitHub username and token
- GitLab URL (optional, defaults to https://gitlab.com)
- GitLab username and token (optional)

Configuration is stored in `~/.gitea-sync.yaml` with secure permissions (0600).

### Getting API Tokens

**Gitea:**
1. Go to Settings → Applications
2. Generate a new token with repository permissions

**GitHub:**
1. Go to Settings → Developer settings → Personal access tokens
2. Generate a new token with `repo` scope

**GitLab:**
1. Go to Settings → Access Tokens (or User Settings → Access Tokens)
2. Generate a new token with `api` scope

## Usage

### Create a new repository

Creates a new repository on Gitea with mirroring to GitHub or GitLab, initializes with README and .gitignore, and clones it locally:

```bash
# Create with GitHub mirroring (default)
./gitea-sync create my-new-project

# Create with GitLab mirroring
./gitea-sync create my-new-project --gitlab

# Create a private repository
./gitea-sync create my-private-project --private

# Explicitly specify GitHub
./gitea-sync create my-new-project --github
```

**What this does:**
1. Creates repo on GitHub or GitLab
2. Creates repo on Gitea
3. Sets up push mirror (Gitea → GitHub/GitLab)
4. Initializes with README.md and .gitignore
5. Pushes initial commit to Gitea
6. Clones the repo to current directory

### Add existing repository with code

Add a local repository that already has code to Gitea with mirroring to GitHub or GitLab:

```bash
# Add current directory (mirrors to GitHub by default)
cd my-existing-project
./gitea-sync add

# Add with GitLab mirroring
./gitea-sync add --gitlab

# Add a specific directory
./gitea-sync add ./my-existing-project

# Add with a custom name
./gitea-sync add --name custom-repo-name

# Add as private repository
./gitea-sync add --private
```

**What this does:**
1. Detects the repository name from directory (or use `--name`)
2. Creates repo on GitHub or GitLab
3. Creates repo on Gitea
4. Sets up push mirror (Gitea → GitHub/GitLab)
5. Adds Gitea as git remote (origin or gitea)
6. Pushes your existing code to Gitea
7. Automatically mirrors to GitHub or GitLab

**Smart remote handling:**
- If no `origin` exists: adds Gitea as `origin`
- If `origin` exists and is Gitea: updates it
- If `origin` exists and is something else: adds Gitea as `gitea` remote

### Add mirror to existing repository

Add a GitHub push mirror to an existing Gitea repository:

```bash
./gitea-sync mirror existing-repo
```

### Bulk setup

Set up mirrors for multiple repositories at once:

```bash
./gitea-sync bulk
```

Then enter repository names, one per line, and press Ctrl+D when done.

You can also pipe a list:

```bash
cat repos.txt | ./gitea-sync bulk
```

## How It Works

**Repository Creation Flow:**
1. Repository is created on GitHub or GitLab
2. Repository is created on Gitea
3. Push mirror is configured (Gitea → GitHub/GitLab)
4. Repository is initialized with:
   - README.md
   - .gitignore
5. Initial commit is pushed to Gitea
6. Mirror automatically syncs to GitHub or GitLab
7. Repository is cloned to current directory

**Mirror Sync:**
- Automatic sync on every commit to Gitea
- Periodic sync every 8 hours
- Changes pushed to Gitea are automatically mirrored to GitHub or GitLab

## Project Structure

```
.
├── main.go                      # Entry point
├── cmd/
│   ├── root.go                  # Root command
│   ├── init.go                  # Config initialization
│   ├── create.go                # Create new repo
│   ├── add.go                   # Add existing repo with code
│   ├── mirror.go                # Add mirror to existing repo
│   └── bulk.go                  # Bulk operations
└── internal/
    ├── config/
    │   └── config.go            # Config management
    ├── gitea/
    │   └── client.go            # Gitea API client
    ├── github/
    │   └── client.go            # GitHub API client
    └── gitlab/
        └── client.go            # GitLab API client
```

## Migration from Shell Scripts

If you were using the original shell scripts:

**bulk-mirror-setup.sh → gitea-sync bulk**
```bash
# Old way
./bulk-mirror-setup.sh

# New way
echo "repo1
repo2
repo3" | ./gitea-sync bulk
```

**setup-new-repo.sh → gitea-sync create**
```bash
# Old way
./setup-new-repo.sh my-repo

# New way
./gitea-sync create my-repo
```

## Security Notes

- API tokens are stored in `~/.gitea-sync.yaml` with permissions 0600
- Never commit the config file to version control
- Consider using environment variables for CI/CD pipelines
- Tokens are transmitted over HTTPS to GitHub (Gitea uses your configured URL)

## Requirements

- Go 1.21 or higher (for building)
- Git installed on your system
- Active Gitea instance
- GitHub account with API access (for GitHub mirroring)
- GitLab account with API access (for GitLab mirroring, optional)

## Troubleshooting

**"config file not found"**
- Run `./gitea-sync init` first

**"failed to create repo"**
- Check your API tokens have correct permissions
- Verify the repository doesn't already exist

**"failed to push"**
- Check your Gitea URL is accessible
- Verify your Gitea token has push permissions

**"not a git repository"**
- Make sure you're in a directory with a `.git` folder
- Run `git init` first if starting a new repo

## License

MIT
