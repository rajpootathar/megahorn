# Twitter Auth: Persistent Chrome Profile

> Replace cookie capture/replay with a persistent Chrome user data directory.

## Problem

Megahorn's Twitter auth captures browser cookies once during `megahorn auth twitter`, stores them in the OS keychain as JSON, and replays them into a fresh Chrome instance on every post. When Twitter invalidates those cookies (logout, expiry, token rotation), posting fails silently. Unlike LinkedIn (OAuth2 + refresh tokens), Twitter has no refresh mechanism.

## Solution

Use Chrome's built-in `--user-data-dir` flag to maintain a persistent browser profile at `~/.megahorn/chrome-profile/`. Chrome manages session cookies, storage, and token refreshes natively. No cookie capture or replay needed.

## Design

### 1. `browser.ChromeOpts` — add `UserDataDir`

```go
type ChromeOpts struct {
    Headed      bool
    ChromePath  string
    UserDataDir string // persistent Chrome profile path
}
```

In `NewContext`, when `UserDataDir != ""`:
- Create the directory if it doesn't exist (`os.MkdirAll` with `0700` permissions)
- Append `chromedp.UserDataDir(opts.UserDataDir)` to the allocator options

Existing anti-detection flags (`enable-automation: false`, `disable-blink-features: AutomationControlled`) remain applied — they are compatible with `--user-data-dir`.

### 2. `config.BrowserConfig` — add `ProfileDir`

```go
type BrowserConfig struct {
    Headed     bool   `mapstructure:"headed"`
    ChromePath string `mapstructure:"chrome_path"`
    ProfileDir string `mapstructure:"profile_dir"`
}
```

Default resolved via `filepath.Join(config.Dir(), "chrome-profile")` (i.e. `~/.megahorn/chrome-profile/`). Not stored as a literal `~` string — uses `os.UserHomeDir()` like existing `config.Dir()`.

Users can override via `config.yaml`:

```yaml
browser:
  profile_dir: /custom/path
```

The profile is shared across all browser-automated platforms (Twitter, future Reddit). Cookie domains are isolated by the browser, so no cross-platform conflicts.

### 3. `twitter.Auth` — simplify

Before:
1. Open fresh Chrome
2. Navigate to x.com/login
3. User logs in
4. Capture all cookies via `network.GetCookies()`
5. Serialize to JSON, store in keyring

After:
1. Open Chrome **with persistent profile**
2. Navigate to x.com/login
3. User logs in, presses Enter
4. Delete old `twitter:cookies` keyring entry if it exists (migration cleanup)
5. Store `"true"` marker in keyring as `twitter:auth` (so `Status()` works)

No `network.GetCookies()`, no JSON serialization.

**Login verification**: not added. The current code has the same limitation — it captures whatever cookies exist without validating login success. The `"true"` marker simply records that the auth flow was completed. If the user presses Enter without actually logging in, `Post()` will fail at compose with a clear error. Acceptable for v1.

### 4. `twitter.Post` — remove cookie restore

Before: call `restoreCookies()` which deserializes JSON cookies from keyring and sets them via `network.SetCookie()`.

After: open Chrome with persistent profile. Session is already there. Navigate directly to compose.

Delete the `restoreCookies()` method entirely.

**Session expiry at post time**: If Twitter has invalidated the session server-side (password change, security event), Chrome will redirect to a login page instead of showing compose. The existing `WaitVisible` for the tweet textarea will time out and return an error with a screenshot. This is the implicit detection — no change needed. The error message + screenshot make it clear the user needs to re-auth.

**Profile lock error**: If another megahorn instance is using the profile, chromedp returns a cryptic error. Catch this and surface a user-friendly message: "Another megahorn instance is running. Close it and retry."

### 5. `twitter.Status` — check auth marker

Before: checks keyring for `twitter:cookies` key (presence only, no validation).

After: checks keyring for `twitter:auth` key containing `"true"`.

Functionally equivalent but no longer stores stale cookie data.

### 6. `twitter.chromeOpts()` — pass profile dir

Update the `chromeOpts()` helper to populate `UserDataDir` from `cfg.Browser.ProfileDir` (with default fallback).

## Migration

**Users must re-run `megahorn auth twitter` after upgrading.** The new code checks for `twitter:auth` (not `twitter:cookies`), so `Status()` will return `not_configured` until re-auth.

During `Auth()`, the old `twitter:cookies` keyring entry is deleted to clean up stale session tokens from the OS keychain.

## Files Changed

| File | Change |
|------|--------|
| `internal/browser/chrome.go` | Add `UserDataDir` to `ChromeOpts`, pass to allocator, mkdir |
| `internal/config/config.go` | Add `ProfileDir` to `BrowserConfig`, default via `config.Dir()` |
| `internal/platform/twitter.go` | Simplify `Auth()`, remove `restoreCookies()`, update `chromeOpts()`, `Post()`, `Status()`. Remove `network` and `json` imports. |

## Files NOT Changed

- `internal/platform/platform.go` — interface unchanged
- `internal/auth/keyring.go` — unchanged
- LinkedIn/Reddit platforms — unaffected
- Post typing/clicking logic — unchanged

## Trade-offs

- **Disk**: Chrome profile is ~50-100MB. Acceptable for a CLI tool.
- **Concurrency**: Chrome locks the profile dir — can't run two megahorn instances simultaneously. Acceptable since megahorn is single-user CLI.
- **Portability**: Profile is machine-local (not synced via keyring). Acceptable — same as browser sessions everywhere.
- **Security**: Session data moves from encrypted OS keychain to unencrypted files on disk in the profile directory. Mitigated by `0700` directory permissions (owner-only access). This is the same security model as every browser on the system. Acceptable trade-off for session reliability.

## Future

- Apply for Twitter developer account → migrate to OAuth2 API (Approach C)
- The `Platform` interface stays the same, so swapping browser automation for API calls is a clean replacement
