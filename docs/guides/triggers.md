# Triggers Guide

This guide covers creating and managing webhook triggers in MUXI formations.

## CLI Commands

### List Triggers

View all triggers defined in a running formation:

```bash
muxi triggers
```

**Example output:**

```
    NAME                 DESCRIPTION
    github-issue         Handle GitHub issue events
    linear-ticket        Process Linear ticket updates
    slack-message        Respond to Slack messages
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID (default: from formation.afs) |
| `--profile` | `-p` | Server profile (default: from .muxi or global) |

Requires **client API key** (from `secrets.enc` or `MUXI_CLIENT_KEY`).

### Fire a Trigger

Invoke a trigger with JSON data:

```bash
# Simple invocation
muxi trigger github-issue --data '{"issue": {"title": "Bug"}}'

# From a file
muxi trigger github-issue --file payload.json

# Async mode (returns job ID)
muxi trigger github-issue --data '...' --async
```

---

## What is a Trigger?

A trigger is a prompt template that gets invoked via webhooks. When an external service sends data to your formation's webhook endpoint, the trigger template is rendered with that data and sent to an agent.

Use cases:
- GitHub webhook notifications
- Slack message handling
- Form submissions
- Scheduled tasks (via cron webhooks)
- Third-party integrations

## Creating a Trigger

Must be run inside a formation directory.

### Interactive Mode (Recommended)

```bash
muxi new trigger
```

The wizard guides you through:
1. **Trigger ID** - Unique identifier (e.g., `github-issue`)
2. **Title** - Human-friendly name
3. **Description** - What this trigger handles

### With Name

```bash
muxi new trigger github-issue
```

Spaces are normalized: `"GitHub Issue"` becomes `github-issue`.

### Non-Interactive

```bash
muxi new trigger github-issue --no-wizard
```

## File Structure

Triggers are Markdown files stored in `triggers/`:

```
my-formation/
└── triggers/
    ├── github-issue.md
    ├── slack-message.md
    └── daily-report.md
```

## Trigger Format

Triggers use Markdown with YAML frontmatter:

```markdown
---
name: "GitHub Issue Handler"
description: "Process new GitHub issues"
agent: issue-handler
---

# New GitHub Issue

A new issue was created in repository **${{ data.repository.full_name }}**.

## Issue Details

- **Title:** ${{ data.issue.title }}
- **Number:** #${{ data.issue.number }}
- **Author:** @${{ data.issue.user.login }}
- **Labels:** ${{ data.issue.labels | map: "name" | join: ", " }}

## Description

${{ data.issue.body }}

---

Please analyze this issue and:
1. Categorize it (bug, feature, question, etc.)
2. Suggest appropriate labels
3. Draft an initial response
```

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Display name for the trigger |
| `description` | Yes | Brief explanation |
| `agent` | No | Target agent ID (uses default if not specified) |
| `sop` | No | SOP to execute instead of direct agent call |

## Data Interpolation

Use `${{ data.* }}` syntax to access webhook payload values:

### Simple Values

```markdown
User: ${{ data.user.name }}
Email: ${{ data.user.email }}
```

### Nested Objects

```markdown
Repository: ${{ data.repository.full_name }}
Owner: ${{ data.repository.owner.login }}
```

### Arrays

```markdown
First item: ${{ data.items[0].name }}
All items: ${{ data.items | map: "name" | join: ", " }}
```

### Conditionals

```markdown
${{ if data.priority == "high" }}
⚠️ HIGH PRIORITY - Respond immediately!
${{ endif }}
```

### Default Values

```markdown
Status: ${{ data.status | default: "pending" }}
```

## Routing to Agents

### Single Agent

```yaml
---
agent: support-bot
---
```

### Dynamic Routing

Use the payload to determine the agent:

```yaml
---
agent: ${{ data.department }}-agent
---
```

### Using SOPs

Route to a workflow instead of a single agent:

```yaml
---
sop: customer-onboarding
---
```

## Common Trigger Patterns

### GitHub Issue

```markdown
---
name: "GitHub Issue"
description: "Handle new GitHub issues"
agent: issue-triage
---

# New Issue: ${{ data.issue.title }}

**Repository:** ${{ data.repository.full_name }}
**Author:** @${{ data.issue.user.login }}
**URL:** ${{ data.issue.html_url }}

## Body

${{ data.issue.body }}

---

Analyze and triage this issue.
```

### GitHub PR

```markdown
---
name: "Pull Request"
description: "Review new pull requests"
agent: code-reviewer
---

# PR: ${{ data.pull_request.title }}

**Author:** @${{ data.pull_request.user.login }}
**Branch:** ${{ data.pull_request.head.ref }} → ${{ data.pull_request.base.ref }}
**Files Changed:** ${{ data.pull_request.changed_files }}

## Description

${{ data.pull_request.body }}

---

Review this PR for:
1. Code quality
2. Security issues
3. Test coverage
```

### Slack Message

```markdown
---
name: "Slack Message"
description: "Respond to Slack messages"
agent: slack-bot
---

# Message from ${{ data.user.name }}

**Channel:** #${{ data.channel.name }}
**Time:** ${{ data.timestamp }}

## Message

${{ data.text }}

---

Respond helpfully to this message.
```

### Form Submission

```markdown
---
name: "Contact Form"
description: "Handle contact form submissions"
agent: support-agent
---

# New Contact Form Submission

**From:** ${{ data.name }} (${{ data.email }})
**Subject:** ${{ data.subject }}

## Message

${{ data.message }}

---

Draft a response to this inquiry.
```

### Scheduled Task

```markdown
---
name: "Daily Report"
description: "Generate daily summary"
agent: reporter
sop: daily-report-workflow
---

# Daily Report Request

**Date:** ${{ data.date }}
**Type:** ${{ data.report_type | default: "standard" }}

Generate the daily report covering:
- Key metrics
- Notable events
- Action items
```

## Webhook Setup

Once deployed, your formation exposes webhook endpoints:

```
POST https://your-formation.muxi.cloud/triggers/{trigger-id}
```

### Authentication

Webhooks can require authentication:
- API key in header
- HMAC signature verification
- Bearer token

Configure in `formation.afs`:

```yaml
triggers:
  auth:
    type: api_key
    header: X-Webhook-Secret
    key: ${{ secrets.WEBHOOK_SECRET }}
```

## Best Practices

1. **Validate data** - Check required fields exist
2. **Use defaults** - Handle missing optional fields
3. **Clear prompts** - Tell the agent exactly what to do
4. **Log context** - Include relevant metadata
5. **Handle errors** - What if data is malformed?
6. **Secure webhooks** - Always use authentication
7. **Test thoroughly** - Send sample payloads before going live
