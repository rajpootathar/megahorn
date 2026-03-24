<p align="center">
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?style=flat-square&logo=go" alt="Go 1.26" />
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square" alt="MIT License" />
  <img src="https://img.shields.io/badge/Platforms-3-blue?style=flat-square" alt="3 Platforms" />
  <img src="https://img.shields.io/badge/MCP-Ready-purple?style=flat-square" alt="MCP Ready" />
</p>

# Megahorn

**One binary to rule them all.** Cross-post to social media from your terminal or let AI agents do it for you.

```
You: "promote our new feature"
Claude: crafts platform-specific posts, shows you for approval
You: "looks good"
Claude: megahorn_post → Twitter, LinkedIn, Reddit → done
```

Megahorn is a standalone Go CLI + MCP server. It handles authentication and publishing — your API keys never leave your machine, never enter an AI conversation.

---

## Why Megahorn?

- **Single binary** — `go install` and you're done. No Node.js, no Python, no Docker.
- **Three platforms, one command** — Twitter, LinkedIn, Reddit. More coming.
- **AI-native** — Ships as an MCP server. Claude Code, Cursor, or any MCP client can post on your behalf.
- **Secure by design** — Credentials live in your OS keychain. The MCP server is the security boundary.
- **Product-agnostic** — Works for any project. The AI skill reads your repo's context automatically.
- **Free** — No paid Twitter API. Browser automation for Twitter, free OAuth for LinkedIn and Reddit.

---

## Installation

### 1. Install the binary

```bash
# Option A: Go install
go install github.com/rajpootathar/megahorn@latest

# Option B: From source
git clone https://github.com/rajpootathar/megahorn.git
cd megahorn && make build
sudo mv megahorn /usr/local/bin/
```

### 2. Install the AI skill (optional)

```bash
npx skills add rajpootathar/megahorn
```

This installs the `social-post` skill for Claude Code, Cursor, or any compatible AI agent. The skill teaches the agent how to write platform-adapted content and call Megahorn's MCP tools.

### 3. Register MCP server (optional)

```bash
megahorn install
# Restart your AI agent to activate
```

This adds Megahorn to your Claude Code MCP settings so the agent can call `megahorn_post`, `megahorn_auth_status`, and `megahorn_auth` as native tools.

---

## Quick Start

### Authenticate

```bash
megahorn auth twitter        # Opens Chrome — log in manually, handle 2FA
megahorn auth linkedin       # OAuth2 flow — opens browser for consent
megahorn auth reddit         # OAuth2 flow — opens browser for consent
megahorn auth status         # Check what's connected
```

```
$ megahorn auth status
Platform       Status
------------------------------
twitter        + authenticated
linkedin       + authenticated
reddit         x not_configured
```

### Post

```bash
# To specific platforms
megahorn post -t "Just shipped journey tracking"
megahorn post -t -l "We cut query time from 12s to 40ms"
megahorn post -r --subreddit SaaS,startups "Launched our analytics dashboard"

# To all authenticated platforms
megahorn post --all "Big announcement today"

# Preview without posting
megahorn post --dry-run -t -l -r --subreddit webdev "test post"

# From a file
megahorn post --file announcement.md -t -l

# JSON output (for scripting / piping)
megahorn post --json --dry-run -t -l "test"
```

```
$ megahorn post -t -l -r --subreddit SaaS "Shipped user journey tracking"
TWITTER: https://x.com/yourhandle/status/1234567890
LINKEDIN: https://www.linkedin.com/feed/update/urn:li:share:12345
REDDIT: https://reddit.com/r/SaaS/comments/abc123

3/3 published.
```

### Configure

```bash
megahorn config                                          # Show current config
megahorn config set browser.headed true                  # Always use visible browser
megahorn config set platforms.defaults twitter,linkedin   # Default platforms when no flags
megahorn config set browser.chrome_path /usr/bin/chromium # Custom Chrome path
```

---

## Platform Setup

### Twitter

Twitter uses **browser automation** via [chromedp](https://github.com/chromedp/chromedp) — no paid API needed.

1. Run `megahorn auth twitter`
2. Chrome opens to twitter.com/login
3. Log in normally (2FA supported)
4. Press Enter in the terminal once you see your feed
5. Session cookies are saved to your OS keychain

**Headed vs headless:** Auth always opens a visible browser. Posting defaults to headless (background). Use `--headed` to watch it happen.

**If selectors break:** Twitter changes its DOM frequently. Megahorn uses `data-testid` attributes which are more stable, but if posting fails, you can update selectors without rebuilding:

```bash
# Override selectors at ~/.megahorn/selectors/twitter.yaml
compose_button: '[data-testid="SideNav_NewTweet_Button"]'
tweet_textarea: '[data-testid="tweetTextarea_0"]'
post_button: '[data-testid="tweetButtonInline"]'
```

### LinkedIn

LinkedIn uses the **Community Management API** (free, OAuth2).

1. Go to [linkedin.com/developers/apps/new](https://www.linkedin.com/developers/apps/new)
2. Create an app (any name works, e.g., "Megahorn")
3. Under **Products**, request "Share on LinkedIn" and "Sign In with LinkedIn using OpenID Connect"
4. Under **Auth** settings, add redirect URL: `http://localhost:8338/callback`
5. Run `megahorn auth linkedin` — enter your Client ID and Client Secret
6. Browser opens for OAuth consent — authorize and you're done

**Token expiry:** LinkedIn tokens last ~60 days. Megahorn warns you when a token is approaching expiry. Re-run `megahorn auth linkedin` to refresh.

### Reddit

Reddit uses **OAuth2 web app** type (free, no password storage).

1. Go to [reddit.com/prefs/apps](https://www.reddit.com/prefs/apps)
2. Click "create another app"
3. Choose **web app**, set redirect URI to `http://localhost:8338/callback`
4. Run `megahorn auth reddit` — enter your Client ID and Client Secret
5. Browser opens for consent — authorize and you're done

**Token refresh:** Reddit tokens expire hourly. Megahorn auto-refreshes using the stored refresh token — you won't notice.

**Multi-subreddit:** Post to multiple subreddits in one command:
```bash
megahorn post -r --subreddit SaaS,startups,webdev "Your post title\nYour post body"
```
First line = title, rest = body.

---

## MCP Server

Megahorn doubles as an [MCP](https://modelcontextprotocol.io/) server, letting AI agents post to social media through native tool calls.

```bash
megahorn server   # Starts stdio JSON-RPC server
```

### Available Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `megahorn_post` | Post content to a platform | `platform`, `content`, `subreddit?`, `dry_run?` |
| `megahorn_auth_status` | Check authenticated platforms | — |
| `megahorn_auth` | Start auth flow (opens browser) | `platform`, `headed?` |

### Configuration

Add to your AI agent's MCP settings:

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

Or run `megahorn install` to auto-configure Claude Code.

### How AI Agents Use It

With the `social-post` skill installed, an AI agent will:

1. Read your project's CLAUDE.md, docs, and recent git history
2. Understand what you're promoting
3. Write platform-specific content (280ch tweet, LinkedIn thought-piece, Reddit discussion)
4. Show you the posts for approval
5. Call `megahorn_post` for each approved platform
6. Report back with published URLs

The agent crafts the content. Megahorn handles the publishing. Your keys stay local.

---

## Architecture

```
┌──────────────────┐     ┌─────────────────────┐
│  You (terminal)  │────>│                     │───> Twitter  (Chrome CDP)
│  $ megahorn post │     │                     │
└──────────────────┘     │    megahorn          │───> LinkedIn (REST API)
                         │    (Go binary)       │
┌──────────────────┐     │                     │───> Reddit   (REST API)
│  AI Agent (MCP)  │────>│  Credentials:       │
│  Claude / Cursor │     │  OS keychain        │
└──────────────────┘     │  never in context   │
                         └─────────────────────┘
```

### Project Structure

```
megahorn/
├── cmd/                  # CLI commands (cobra)
│   ├── auth.go           # megahorn auth
│   ├── post.go           # megahorn post
│   ├── server.go         # megahorn server (MCP)
│   ├── config.go         # megahorn config
│   └── install.go        # megahorn install
├── internal/
│   ├── platform/         # Platform interface + implementations
│   │   ├── twitter.go    # chromedp browser automation
│   │   ├── linkedin.go   # OAuth2 REST API
│   │   └── reddit.go     # OAuth2 REST API
│   ├── mcp/              # MCP server + tool handlers
│   ├── auth/             # Keyring, OAuth2, browser open
│   ├── browser/          # Chrome launcher, selectors
│   └── config/           # YAML config management
├── skills/
│   └── social-post/
│       └── SKILL.md      # AI agent skill (npx skills add)
├── selectors/
│   └── twitter.yaml      # Default Twitter DOM selectors
└── package.json          # For npx skills add discovery
```

---

## CLI Reference

```
megahorn [command]

Commands:
  auth [platform]     Authenticate with a platform (twitter, linkedin, reddit)
  auth status         Show auth status for all platforms
  post [content]      Post content to platforms
  server              Start MCP server (stdio)
  config              Show configuration
  config set [k] [v]  Set a config value
  install             Configure MCP in Claude Code
  version             Print version

Post Flags:
  -t, --twitter       Post to Twitter
  -l, --linkedin      Post to LinkedIn
  -r, --reddit        Post to Reddit
  -a, --all           All authenticated platforms
      --subreddit     Comma-separated subreddits
      --headed        Visible browser for Twitter
      --dry-run       Preview without posting
  -f, --file          Read content from file
      --json          JSON output

Auth Flags:
      --headed        Visible browser (default: true)
      --headless      Headless browser
```

---

## Security

| Layer | How |
|-------|-----|
| **At rest** | OS keychain (macOS Keychain / Linux secret-service) |
| **In transit** | HTTPS for APIs, WSS for Chrome CDP |
| **In AI context** | Never. MCP tools are the boundary — agents send commands, never see keys |
| **Config files** | Preferences only (browser mode, defaults). Zero secrets. |

---

## Roadmap

**v2 candidates** (no timeline):

Bluesky, Dev.to, Hacker News, Mastodon, Discord, Hashnode, Product Hunt, Threads, Indie Hackers | Read feeds & comments | Media attachments | Twitter threads | Scheduling | Analytics | Windows support

---

## Contributing

PRs welcome. The platform interface makes it easy to add new platforms:

```go
type Platform interface {
    Name() string
    Auth(opts AuthOpts) error
    Post(content string, opts PostOpts) (*PostResult, error)
    Status() AuthStatus
}
```

Implement these four methods, register in the registry, done.

---

## License

MIT
