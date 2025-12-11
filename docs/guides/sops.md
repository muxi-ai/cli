# SOPs Guide

This guide covers creating and managing Standard Operating Procedures (SOPs) in MUXI formations.

## CLI Commands

### List SOPs

View all SOPs defined in a running formation:

```bash
muxi sops
```

**Example output:**

```
    NAME                TYPE        DESCRIPTION
    customer-onboard    template    Guide new customers through setup
    bug-triage          guide       Categorize and prioritize bugs
    content-review      template    Review and approve content
```

### Show SOP Details

View detailed information about a specific SOP:

```bash
muxi sops show customer-onboard
```

**Example output:**

```
  SOP: customer-onboard

    Description:  Guide new customers through account setup
    Type:         template
    Steps:        4
    Agents:       [support-agent, setup-assistant]

    Content:
    ─────────────────────────────────────────
    # Customer Onboarding

    ## Step 1: Welcome and Verification
    ...
    ─────────────────────────────────────────
```

**Flags (both commands):**

| Flag | Short | Description |
|------|-------|-------------|
| `--formation` | `-F` | Formation ID (default: from formation.afs) |
| `--profile` | `-p` | Server profile (default: from .muxi or global) |

Requires **client API key** (from `secrets.enc` or `MUXI_CLIENT_KEY`).

---

## What is an SOP?

An SOP (Standard Operating Procedure) defines a workflow that agents follow to complete complex tasks. SOPs include:
- Step-by-step instructions
- Agent assignments for each step
- Tool references
- Decision points and branches

## Creating an SOP

Must be run inside a formation directory.

### Interactive Mode (Recommended)

```bash
muxi new sop
```

The wizard guides you through:
1. **SOP ID** - Unique identifier (e.g., `customer-onboarding`)
2. **Title** - Human-friendly name
3. **Description** - What this workflow accomplishes

### With Name

```bash
muxi new sop customer-onboarding
```

Spaces are normalized: `"Customer Onboarding"` becomes `customer-onboarding`.

### Non-Interactive

```bash
muxi new sop customer-onboarding --no-wizard
```

## File Structure

SOPs are Markdown files stored in `sops/`:

```
my-formation/
└── sops/
    ├── customer-onboarding.md
    ├── bug-triage.md
    └── content-review.md
```

## SOP Format

SOPs use Markdown with YAML frontmatter:

```markdown
---
name: "Customer Onboarding"
description: "Guide new customers through account setup and initial configuration"
mode: sequential
tags:
  - customer-success
  - onboarding
---

# Customer Onboarding

This SOP guides new customers through the onboarding process.

## Step 1: Welcome and Verification

**Agent:** @support-agent

1. Greet the customer warmly
2. Verify their account details
3. Confirm their subscription tier

## Step 2: Initial Setup

**Agent:** @setup-assistant
**Tools:** @database, @email-sender

1. Create user profile in the system
2. Set default preferences based on tier
3. Send welcome email with credentials

## Step 3: Training Session

**Agent:** @training-bot

1. Walk through key features
2. Answer questions about functionality
3. Provide relevant documentation links

## Step 4: Handoff

**Agent:** @support-agent

1. Confirm customer is comfortable
2. Provide support contact information
3. Schedule follow-up check-in
```

## Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Display name for the SOP |
| `description` | Yes | Brief explanation of the workflow |
| `mode` | No | Execution mode: `sequential` (default) or `parallel` |
| `tags` | No | Categories for organization |

## Agent References

Use `@agent-id` to assign steps to specific agents:

```markdown
**Agent:** @research-assistant

1. Search for relevant information
2. Compile findings into summary
```

Multiple agents can be assigned:

```markdown
**Agents:** @researcher, @writer

1. Research collaboratively
2. Draft content together
```

## Tool References

Use `@mcp-id` to specify which tools a step should use:

```markdown
**Tools:** @web-search, @database

1. Search the web for recent articles
2. Store findings in the database
```

## Execution Modes

### Sequential (Default)

Steps execute one after another:

```yaml
mode: sequential
```

### Parallel

Steps can execute simultaneously when possible:

```yaml
mode: parallel
```

## Decision Points

Use conditional formatting for branching logic:

```markdown
## Step 3: Evaluate Request

**Agent:** @coordinator

1. Review the request type
2. **If** high priority:
   - Escalate to @senior-agent
   - Notify management
3. **If** standard priority:
   - Proceed to Step 4
4. **If** low priority:
   - Queue for batch processing
```

## Secrets in SOPs

Reference secrets using the standard syntax:

```markdown
## Step 2: API Integration

**Agent:** @integration-bot
**Tools:** @external-api

1. Authenticate using API credentials
2. Fetch customer data
   - Use key: `${{ secrets.EXTERNAL_API_KEY }}`
```

## Common SOP Patterns

### Customer Support Workflow

```markdown
---
name: "Customer Support"
description: "Handle incoming customer inquiries"
mode: sequential
tags: [support, customer-service]
---

## Step 1: Classify Request
**Agent:** @classifier
- Determine request type (bug, question, feature)
- Assess priority level

## Step 2: Route Request
**Agent:** @coordinator
- **Bug:** Route to @technical-support
- **Question:** Route to @faq-bot
- **Feature:** Route to @product-team

## Step 3: Resolve
**Agent:** (assigned in Step 2)
- Address the customer's needs
- Document resolution

## Step 4: Follow Up
**Agent:** @support-agent
- Confirm satisfaction
- Close ticket
```

### Content Creation Pipeline

```markdown
---
name: "Content Pipeline"
description: "Create and publish blog content"
mode: sequential
tags: [content, marketing]
---

## Step 1: Research
**Agent:** @researcher
**Tools:** @web-search
- Research topic thoroughly
- Gather supporting data

## Step 2: Draft
**Agent:** @writer
- Create initial draft
- Include key points from research

## Step 3: Review
**Agent:** @editor
- Check for accuracy
- Improve readability
- Suggest edits

## Step 4: Publish
**Agent:** @publisher
**Tools:** @cms-api
- Format for platform
- Schedule publication
```

## Best Practices

1. **Clear step names** - Make each step self-explanatory
2. **Assign agents** - Every step should have an owner
3. **Specify tools** - List required MCPs for clarity
4. **Handle errors** - Include fallback steps
5. **Keep it focused** - One workflow per SOP
6. **Use tags** - Organize SOPs by category
7. **Test the flow** - Walk through manually first
