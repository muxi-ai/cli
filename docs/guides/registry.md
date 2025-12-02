# Registry Guide

The MUXI Registry at `registry.muxi.org` is the central hub for sharing and discovering formations.

## Authentication

### Login

```bash
muxi login
```

This opens your browser for GitHub OAuth authentication. If the browser doesn't open, you can paste the authentication URL manually.

After successful login, your credentials are stored locally for future commands.

### Logout

```bash
muxi logout
```

Removes stored credentials for the default registry.

```bash
muxi logout --registry registry.muxi.org
```

Removes credentials for a specific registry.

## Publishing Formations

### Push

Publish your formation to the registry:

```bash
cd my-formation
muxi push
```

The push command:
1. Validates your formation
2. Creates a bundle (excludes secrets.enc, .key, .git, .env)
3. Publishes to the registry
4. Creates a GitHub repository for the formation

**Options:**

```bash
muxi push --org myorg        # Publish under an organization
muxi push --dry-run          # Preview without publishing
muxi push --registry <url>   # Use specific registry
```

**What gets published:**
- `formation.yaml`
- `README.md`, `LICENSE`
- `secrets` (template file, NOT secrets.enc)
- `agents/*.yaml`
- `mcps/*.yaml`
- `a2a/*.yaml`
- `sops/*.md`
- `triggers/*.yaml`
- `knowledge/*.md`

**What's excluded:**
- `secrets.enc` (encrypted secrets)
- `.key` (encryption key)
- `.git/`, `.env`, `node_modules/`
- Other dot files

## Downloading Formations

### Pull

Download a formation from the registry:

```bash
muxi pull @username/formation-name
```

**Options:**

```bash
muxi pull @user/formation -o ./my-dir   # Custom output directory
muxi pull @user/formation --force       # Overwrite existing
muxi pull @user/formation:1.0.0         # Specific version
```

After pulling, configure your secrets:

```bash
cd formation-name
muxi secrets setup
```

## Discovering Formations

### Search

Find formations by keyword:

```bash
muxi search "weather api"
```

**Sort options:**

```bash
muxi search "ai" --sort trending    # Default - popular this week
muxi search "ai" --sort downloads   # Most downloaded
muxi search "ai" --sort stars       # Most starred
muxi search "ai" --sort recent      # Recently published
```

**Pagination:**

Results are paginated (10 per page). Use `n` for next page, `p` for previous, `q` to quit.

### Show

View formation details:

```bash
muxi show @username/formation-name
```

**Show all versions:**

```bash
muxi show @user/formation --versions
```

## Your Formations

### List Your Publications

```bash
muxi registry mine
```

Shows all formations you've published with download and star counts.

## Managing Registries

### List Configured Registries

```bash
muxi registry list
```

### Add a Registry

```bash
muxi registry add
```

Prompts for URL and automatically authenticates.

### Remove a Registry

```bash
muxi registry remove
```

Select from configured registries to remove.

### Set Default Registry

```bash
muxi registry default
```

Select which registry to use by default.

## Examples

### Publish Your First Formation

```bash
# Create a new formation
muxi new formation

# Configure it
muxi config llm
muxi secrets set OPENAI_API_KEY

# Publish
muxi push
```

### Use a Shared Formation

```bash
# Find something useful
muxi search "slack integration"

# Download it
muxi pull @creator/slack-bot

# Set up your secrets
cd slack-bot
muxi secrets setup

# You're ready!
```

### Update a Published Formation

```bash
# Bump version in formation.yaml
# version: "1.0.0" -> "1.1.0"

# Push the update
muxi push
```

## Troubleshooting

### "Not authenticated"

Run `muxi login` to authenticate.

### "Version already exists"

Bump the version in `formation.yaml` before pushing.

### "Formation not found"

Check the spelling: `@username/formation-name`

### "You don't have permission"

You can only update formations you own.
