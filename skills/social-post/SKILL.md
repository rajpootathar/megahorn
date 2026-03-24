---
name: social-post
description: Write and publish social media posts using Megahorn MCP. Reads project context to craft platform-specific content for Twitter, LinkedIn, and Reddit.
---

# Social Post — Megahorn Skill

You have access to Megahorn MCP tools for posting to social media. This skill teaches you how to write great posts and publish them.

## Prerequisites

Check that Megahorn is available:
1. Call `megahorn_auth_status` to verify which platforms are connected
2. If no platforms are authenticated, tell the user to run `megahorn auth <platform>`
3. If `megahorn_auth_status` tool is not available, tell the user to install Megahorn: `go install github.com/rajpootathar/megahorn@latest && megahorn install`

## Workflow

When asked to promote, announce, or post about something:

### 1. Understand the Context
- Read the project's CLAUDE.md for product overview, voice, and audience
- Check recent git commits or changelogs for what was shipped
- If a specific feature/blog/update was mentioned, understand it thoroughly
- Look at docs/, README, or any marketing copy for tone reference

### 2. Write Platform-Specific Content

**Twitter (280 chars max):**
- Lead with a hook: surprising stat, bold claim, or relatable problem
- Use short, punchy sentences
- Include a link if relevant
- No hashtag spam (1-2 max, only if natural)
- No corporate speak — write like a builder shipping something cool

**LinkedIn (up to 3000 chars):**
- Thought-leadership tone, first-person
- Open with a hook paragraph that makes people stop scrolling
- Share insight, lesson learned, or behind-the-scenes
- Professional but not boring — no "I'm thrilled to announce"
- Can include bullet points for readability
- End with a soft CTA or question to drive engagement

**Reddit:**
- Anti-promotional tone — this MUST read as genuine discussion
- Frame as sharing learnings, asking questions, or contributing value
- NEVER use marketing language ("excited to announce", "game-changing")
- Title should be specific and useful, not clickbaity
- Always ask user to confirm the subreddit before posting
- Warn about subreddit rules if you know them (karma requirements, flair, self-promo rules)

### 3. Present for Review

ALWAYS show ALL posts to the user before publishing. Format:

```
## Twitter
> [tweet content here]

## LinkedIn
> [linkedin post here]

## Reddit (r/subreddit)
> Title: [title]
> Body: [body]

Approve all? (or tell me what to change)
```

Wait for explicit approval before calling any megahorn tools.

### 4. Publish

For each approved post, call `megahorn_post` with the platform and content.
Post to platforms independently — if one fails, continue with the others.

Report results:
- Success: show the URL
- Failure: show the error and suggest a fix (e.g., "Token expired — run: megahorn auth linkedin")

## Important Rules

1. **Never post without user approval** — always show drafts first
2. **Never use the same text on every platform** — adapt to each platform's culture
3. **Never use marketing buzzwords on Reddit** — be genuine or don't post
4. **Each platform gets its own `megahorn_post` call** — don't try to batch
5. **If megahorn_auth_status shows a platform as expired, tell the user before trying to post**
6. **Read the project context every time** — don't assume you know the product from memory
7. **For Reddit, always ask user to confirm the subreddit** — never auto-select
