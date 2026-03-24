# Megahorn

One binary to rule them all — cross-post to social media from your terminal or via MCP for AI agents.

## What It Does

Megahorn posts to **Twitter**, **LinkedIn**, and **Reddit** from a single CLI. It also runs as an MCP server, so AI coding agents (Claude Code, Cursor, etc.) can write and publish posts on your behalf.

Your API keys and credentials stay on your machine — they never enter an AI conversation.

## Install

### Binary

```bash
go install github.com/rajpootathar/megahorn@latest
```

### AI Agent Skill

```bash
npx skills add rajpootathar/megahorn
```

### MCP Server (Claude Code)

```bash
megahorn install
# Restart Claude Code to activate
```

## Quick Start

### 1. Authenticate

```bash
megahorn auth twitter      # Opens Chrome, you log in
megahorn auth linkedin     # OAuth2 flow
megahorn auth reddit       # OAuth2 flow
megahorn auth status       # Check connections
```

### 2. Post

```bash
# Post to specific platforms
megahorn post -t "Just shipped journey tracking for QCK"
megahorn post -t -l -r --subreddit SaaS,startups "Big update today"

# Post to all authenticated platforms
megahorn post --all "Megahorn is live"

# Preview without posting
megahorn post --dry-run -t -l "test post"

# From a file
megahorn post --file announcement.md -t -l

# JSON output (for scripting)
megahorn post --json -t -l "test"
```

### 3. Configure

```bash
megahorn config                                    # Show config
megahorn config set browser.headed true            # Visible browser
megahorn config set platforms.defaults twitter,linkedin  # Default platforms
```

## Platform Details

| Platform | Method | Auth | Notes |
|----------|--------|------|-------|
| **Twitter** | Chrome (chromedp) | Session cookies via browser login | Supports headed/headless mode |
| **LinkedIn** | REST API | OAuth2 (you create a dev app) | Tokens expire ~2 months |
| **Reddit** | REST API | OAuth2 (you create a web app) | Auto-refreshes tokens |

### Twitter Setup

Twitter uses browser automation (no paid API needed). On first auth, Chrome opens and you log in normally (including 2FA). Megahorn saves your session cookies.

### LinkedIn Setup

1. Create an app at https://www.linkedin.com/developers/apps/new
2. Request "Share on LinkedIn" product
3. Add redirect URL: `http://localhost:8338/callback`
4. Run `megahorn auth linkedin` and enter your Client ID + Secret

### Reddit Setup

1. Create an app at https://www.reddit.com/prefs/apps
2. Choose "web app", redirect URI: `http://localhost:8338/callback`
3. Run `megahorn auth reddit` and enter your Client ID + Secret

## MCP Server

Megahorn runs as an MCP server for AI agents:

```bash
megahorn server   # Starts stdio JSON-RPC server
```

### Tools

| Tool | Description |
|------|-------------|
| `megahorn_post` | Post content to a platform |
| `megahorn_auth_status` | Check which platforms are authenticated |
| `megahorn_auth` | Initiate authentication for a platform |

### Claude Code Config

```json
{
  "mcpServers": {
    "megahorn": {
      "command": "megahorn",
      "args": ["server"]
    }
  }
}
```

Or just run `megahorn install` to configure automatically.

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--twitter` | `-t` | Post to Twitter |
| `--linkedin` | `-l` | Post to LinkedIn |
| `--reddit` | `-r` | Post to Reddit |
| `--all` | `-a` | All authenticated platforms |
| `--subreddit` | | Comma-separated subreddits |
| `--headed` | | Visible browser window |
| `--headless` | | Headless browser |
| `--dry-run` | | Preview without posting |
| `--file` | `-f` | Read content from file |
| `--json` | | JSON output |

## Security

- Credentials stored in OS keychain (macOS Keychain / Linux secret-service)
- Twitter cookies encrypted at rest
- No secrets in config files
- MCP server holds all secrets — AI agents never see API keys

## Config

Stored at `~/.megahorn/config.yaml`:

```yaml
browser:
  headed: false
  chrome_path: ""
platforms:
  defaults:
    - twitter
    - linkedin
reddit:
  default_subreddits: []
```

## License

MIT
