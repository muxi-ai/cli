# Users Guide

Manage user identifiers and mappings.

## Overview

User identifiers map external IDs (email, phone, Slack ID, etc.) to MUXI user IDs. This allows users to interact with formations from different platforms while maintaining a single identity.

## List Identifiers

List all identifiers for a user:

```bash
muxi users identifiers -u alice
```

Output:
```
Identifiers for user 'alice':

IDENTIFIER                          TYPE            CREATED
alice@company.com                   email           Jan 15, 2024
+1234567890                         phone           Jan 16, 2024
U12345ABC                           slack           Jan 17, 2024
```

## Link Identifier

Link an external identifier to a MUXI user:

```bash
muxi users link -u alice "alice@company.com"
muxi users link -u alice "+1234567890" --type phone
muxi users link -u alice "U12345ABC" --type slack
```

### Identifier Types

- `email` - Email address
- `phone` - Phone number
- `slack` - Slack user ID
- `telegram` - Telegram user ID
- `whatsapp` - WhatsApp number
- `external_id` - Custom external ID

## Unlink Identifier

Remove an identifier mapping:

```bash
muxi users unlink "alice@company.com"
muxi users unlink "+1234567890"
```

## Resolve Identifier

Look up which MUXI user an identifier belongs to:

```bash
muxi users resolve "alice@company.com"
```

Output:
```
Identifier: alice@company.com
User ID:    alice
Type:       email
```

## Options

```
-f, --formation    Formation ID
-p, --profile      Server profile
-u, --user         User ID (required for identifiers/link)
    --type         Identifier type (for link command)
```

## Use Cases

### Multi-Platform Users

Link identifiers from different platforms to a single user:

```bash
muxi users link -u alice "alice@company.com" --type email
muxi users link -u alice "U12345ABC" --type slack
muxi users link -u alice "+1234567890" --type whatsapp
```

Now when Alice messages from any platform, the formation recognizes her as the same user.

### Resolve Unknown Identifiers

When you receive a message from an unknown identifier:

```bash
muxi users resolve "+1987654321"
```

If found, you'll see the mapped user. If not found, link it:

```bash
muxi users link -u bob "+1987654321" --type phone
```

### Audit User Identifiers

List all identifiers to audit a user's linked accounts:

```bash
muxi users identifiers -u alice
```

## Best Practices

1. **Use meaningful user IDs** - Use consistent, human-readable user IDs
2. **Specify types** - Always specify identifier types for clarity
3. **Clean up old identifiers** - Unlink outdated or incorrect mappings
4. **One user per identity** - Avoid linking the same identifier to multiple users
