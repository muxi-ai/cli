# Complete CLI Command Reference

**Date:** 2025-11-25  
**Timeline:** 1 Week  
**Total Commands:** ~80

---

## Implementation Priority

**Day 1-2: Core** - Formation dev, secrets, profiles  
**Day 3: Server** - Lifecycle, management  
**Day 4-5: Formation API** - Agents, MCPs, chat, monitoring  
**Day 6-7: Polish** - Testing, docs, UX

---

## Complete Command Tree

```
muxi
│
├── 🏗️  FORMATION DEVELOPMENT (Local)
│   ├── init <name>                    # Create formation (wizard)
│   ├── validate                       # Validate formation files
│   └── deploy [--profile <name>]      # Deploy to server(s)
│
├── 🔐 SECRET MANAGEMENT (Local)
│   ├── secrets set <key>              # Set secret (prompt for value)
│   ├── secrets list                   # List secret keys (no values)
│   ├── secrets delete <key>           # Delete secret (with validation)
│   └── secrets setup                  # Setup wizard (from secrets.example)
│
├── 🌐 REGISTRY
│   ├── login [registry]               # Authenticate (default: registry.muxi.org)
│   ├── logout [registry]              # Logout
│   ├── push [--registry <url>]        # Publish formation
│   ├── pull <ref> [--registry <url>]  # Download formation
│   └── search <query> [--registry <url>] # Search formations
│
├── 🖥️  PROFILE MANAGEMENT
│   ├── profile add <name>             # Add server profile (wizard)
│   ├── profile list                   # List all profiles
│   ├── profile use <name>             # Set default profile
│   └── profile remove <name>          # Remove profile
│
├── 📦 FORMATION LIFECYCLE (Server API)
│   ├── formation list [--profile <name>]      # List formations on server
│   ├── formation stop <id> [--profile <name>] # Stop formation
│   ├── formation restart <id> [--profile <name>] # Restart formation
│   ├── formation rollback <id> [--profile <name>] # Rollback version
│   └── formation delete <id> [--profile <name>] # Delete formation
│
├── 🖥️  SERVER MANAGEMENT (Server API)
│   ├── server status [--profile <name>]  # Server status & stats
│   ├── server logs [--profile <name>]    # Server audit logs
│   └── server ping [--profile <name>]    # Test reachability
│
├── ⚙️  FORMATION CONFIGURATION (Formation API)
│   ├── config show [--formation <id>] [--profile <name>] # Full config
│   ├── status [--formation <id>] [--profile <name>]      # Runtime status
│   │
│   ├── agent
│   │   ├── list [--formation <id>] [--profile <name>]
│   │   ├── add [--formation <id>] [--profile <name>] [--file <yaml>]
│   │   ├── get <agent_id> [--formation <id>] [--profile <name>]
│   │   ├── update <agent_id> [--formation <id>] [--profile <name>]
│   │   └── delete <agent_id> [--formation <id>] [--profile <name>]
│   │
│   ├── mcp
│   │   ├── list [--formation <id>] [--profile <name>]
│   │   ├── add [--formation <id>] [--profile <name>] [--file <yaml>]
│   │   ├── get <server_id> [--formation <id>] [--profile <name>]
│   │   ├── update <server_id> [--formation <id>] [--profile <name>]
│   │   └── delete <server_id> [--formation <id>] [--profile <name>]
│   │
│   ├── overlord
│   │   ├── show [--formation <id>] [--profile <name>]     # Get overlord config
│   │   └── persona [--formation <id>] [--profile <name>]  # Get persona
│   │
│   ├── llm
│   │   ├── settings [--formation <id>] [--profile <name>] # Get LLM settings
│   │   └── models [--formation <id>] [--profile <name>]   # List configured models
│   │
│   └── memory
│       ├── config [--formation <id>] [--profile <name>]   # Get memory config
│       └── buffer list [--formation <id>] [--profile <name>] # List buffer entries
│
├── 💬 INTERACTION (Formation API)
│   ├── chat [--formation <id>] [--profile <name>]         # Interactive chat
│   ├── avchat [--formation <id>] [--profile <name>] --file <path> # Audio/video chat
│   │
│   ├── session
│   │   ├── list [--formation <id>] [--profile <name>] [--user <user_id>]
│   │   ├── get <session_id> [--formation <id>] [--profile <name>]
│   │   ├── messages <session_id> [--formation <id>] [--profile <name>]
│   │   └── delete <session_id> [--formation <id>] [--profile <name>]
│   │
│   ├── trigger
│   │   ├── list [--formation <id>] [--profile <name>]
│   │   ├── get <name> [--formation <id>] [--profile <name>]
│   │   └── invoke <name> [--formation <id>] [--profile <name>] [--data <json>]
│   │
│   ├── job
│   │   ├── list [--formation <id>] [--profile <name>] [--user <user_id>]
│   │   └── cancel <job_id> [--formation <id>] [--profile <name>]
│   │
│   └── request
│       ├── status <request_id> [--formation <id>] [--profile <name>]
│       └── cancel <request_id> [--formation <id>] [--profile <name>]
│
├── 📊 MONITORING & LOGS (Formation API)
│   ├── logs [--formation <id>] [--profile <name>] [--follow]
│   │
│   ├── logging
│   │   ├── config [--formation <id>] [--profile <name>]
│   │   ├── destinations list [--formation <id>] [--profile <name>]
│   │   ├── destinations add [--formation <id>] [--profile <name>]
│   │   ├── destinations update <id> [--formation <id>] [--profile <name>]
│   │   └── destinations delete <id> [--formation <id>] [--profile <name>]
│   │
│   └── audit
│       ├── list [--formation <id>] [--profile <name>]
│       └── clear [--formation <id>] [--profile <name>] --confirm
│
├── 🗓️  SCHEDULER (Formation API)
│   ├── scheduler config [--formation <id>] [--profile <name>]
│   ├── scheduler list [--formation <id>] [--profile <name>]
│   ├── scheduler add [--formation <id>] [--profile <name>]
│   ├── scheduler update <id> [--formation <id>] [--profile <name>]
│   └── scheduler delete <id> [--formation <id>] [--profile <name>]
│
├── 🤝 A2A CONFIGURATION (Formation API)
│   ├── a2a config [--formation <id>] [--profile <name>]
│   └── a2a outbound list [--formation <id>] [--profile <name>]
│
├── 📚 SOPS & KNOWLEDGE (Formation API - Read-only)
│   ├── sop list [--formation <id>] [--profile <name>]
│   └── sop get <name> [--formation <id>] [--profile <name>]
│
└── 👥 USERS (Formation API)
    ├── user identifiers list [--formation <id>] [--profile <name>]
    ├── user identifiers add [--formation <id>] [--profile <name>]
    ├── user get <identifier> [--formation <id>] [--profile <name>]
    └── user delete <identifier> [--formation <id>] [--profile <name>]
```

---

## ALL Commands Designed ✅

**Total:** ~80 commands covering all API endpoints

**Implementation:** Priority order (6 weeks), but ALL designed now

---

**Status:** Complete command reference  
**Next:** Implement in priority order
