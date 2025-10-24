# CLI Implementation Plan

**Status:** Ready to Start  
**Version:** 2.0  
**Last Updated:** 2025-10-24

---

## 🎯 Implementation Scope

Based on approved design decisions from CLI-DESIGN-IMPROVEMENTS.md:

**✅ Implementing:**
- #2: `~/.muxi/cli/config.yaml` - Global CLI settings
- #3: Global flags (shorthands + flexible placement)
- #6: Wizard commands (interactive)
- #8: Registry push/pull (TBD placeholders)
- #9: Remove `local` prefix, auto-pack
- #12: `.muxi` project file support

**⏳ Deferred:**
- #5: Multi-server sync (discuss strategy later)

---

## 📋 Implementation Phases

### Phase 1: Foundation (Week 1)

#### 1.1 Configuration Management

**Files to create:**
- `pkg/config/config.go` - Load/save config files
- `pkg/config/paths.go` - Path constants and helpers

**Config file structure:**
```go
// ~/.muxi/cli/config.yaml
type Config struct {
    DefaultProfile  string `yaml:"default_profile"`
    DefaultRegistry string `yaml:"default_registry"`
    OutputFormat    string `yaml:"output_format"`    // text, json, yaml
    NoColor         bool   `yaml:"no_color"`
    Debug           bool   `yaml:"debug"`
}

// ~/.muxi/cli/profiles.yaml
type ProfilesFile struct {
    Version        string              `yaml:"version"`
    DefaultProfile string              `yaml:"default_profile"`
    Profiles       map[string]*Profile `yaml:"profiles"`
}

type Profile struct {
    Servers []Server `yaml:"servers"`
}

type Server struct {
    ID   string      `yaml:"id"`           // default, us-east-1, etc.
    URL  string      `yaml:"url"`
    Auth ProfileAuth `yaml:"auth"`
}

type ProfileAuth struct {
    KeyID     string `yaml:"key_id"`
    SecretKey string `yaml:"secret_key"`
}

// Helper methods for future multi-server support
func (p *Profile) GetDefaultServer() *Server {
    if len(p.Servers) == 0 {
        return nil
    }
    // For now: just return first server
    // Future: support server selection
    return &p.Servers[0]
}

// ~/.muxi/cli/formations.yaml
type FormationsFile struct {
    Version    string                  `yaml:"version"`
    Formations map[string]*FormationCreds `yaml:"formations"`
}

type FormationCreds struct {
    URL       string `yaml:"url,omitempty"`
    AdminKey  string `yaml:"admin_key"`
    ClientKey string `yaml:"client_key"`
    AddedAt   string `yaml:"added_at"`
    Notes     string `yaml:"notes,omitempty"`
}

// .muxi (per-formation)
type FormationConfig struct {
    Profile  string            `yaml:"profile"`
    Registry string            `yaml:"registry"`
    Env      map[string]string `yaml:"env,omitempty"`
}
```

**Tasks:**
- [x] Design config structs ✅
- [ ] Implement `config.LoadConfig()` → `~/.muxi/cli/config.yaml`
- [ ] Implement `config.LoadProfiles()` → `~/.muxi/cli/profiles.yaml`
- [ ] Implement `config.LoadFormations()` → `~/.muxi/cli/formations.yaml`
- [ ] Implement `config.LoadFormationConfig()` → `.muxi` (optional)
- [ ] Implement `config.SaveConfig()`, `SaveProfiles()`, `SaveFormations()`
- [ ] Implement `config.EnsureConfigDir()` → create `~/.muxi/cli/` if missing
- [ ] Add migration from old `~/.muxi/profiles.yaml` to `~/.muxi/cli/profiles.yaml`

**Acceptance criteria:**
- Can load/save all config files
- Auto-creates directories if missing
- Validates YAML structure
- Handles missing files gracefully
- Migrates old profile locations

---

#### 1.2 Global Flags (Cobra Integration)

**File:** `cmd/root.go`

**Shorthands to add:**
```go
rootCmd.PersistentFlags().StringP("profile", "p", "", "Server profile to use")
rootCmd.PersistentFlags().StringP("output", "o", "text", "Output format (text, json, yaml)")
rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug logging")
rootCmd.PersistentFlags().BoolP("no-color", "c", false, "Disable colored output")
// -h is already help in Cobra
```

**Tasks:**
- [ ] Add shorthand flags to root command
- [ ] Test flag placement (before/after/middle)
- [ ] Implement flag precedence:
  1. CLI flags
  2. Environment variables (`MUXI_PROFILE`, `MUXI_DEBUG`, etc.)
  3. `.muxi` file
  4. `~/.muxi/cli/config.yaml`
- [ ] Add `--formation` flag for hybrid commands

**Acceptance criteria:**
- Flags work anywhere: `muxi -p prod formation list`, `muxi formation list -p prod`
- Shorthands work: `-p`, `-o`, `-d`, `-c`
- Flag precedence correct
- Environment variables override config file
- CLI flags override everything

---

#### 1.3 Formation Context Detection

**File:** `pkg/context/formation.go`

**Detection logic:**
```go
type FormationContext struct {
    Path        string                 // Formation directory path
    FormationID string                 // From formation.yaml
    Config      *FormationYAML         // Parsed formation.yaml
    Secrets     *FormationSecrets      // Decrypted secrets.enc
    LocalConfig *FormationConfig       // From .muxi (optional)
}

func DetectFormationContext() (*FormationContext, error) {
    // Walk up directory tree looking for formation.yaml
    cwd, _ := os.Getwd()
    for {
        formationPath := filepath.Join(cwd, "formation.yaml")
        if fileExists(formationPath) {
            ctx := &FormationContext{Path: cwd}
            
            // Parse formation.yaml
            ctx.Config = parseFormationYAML(formationPath)
            ctx.FormationID = ctx.Config.ID
            
            // Decrypt secrets.enc (if exists)
            ctx.Secrets = decryptSecrets(cwd)
            
            // Load .muxi (if exists)
            ctx.LocalConfig = loadFormationConfig(cwd)
            
            return ctx, nil
        }
        
        parent := filepath.Dir(cwd)
        if parent == cwd {
            break // Reached root
        }
        cwd = parent
    }
    
    return nil, errors.New("not in formation directory")
}
```

**Tasks:**
- [ ] Implement `DetectFormationContext()`
- [ ] Implement `parseFormationYAML()` (basic parser)
- [ ] Implement `decryptSecrets()` (use runtime's encryption logic)
- [ ] Implement `loadFormationConfig()` for `.muxi`
- [ ] Add caching to avoid repeated file reads

**Acceptance criteria:**
- Detects formation.yaml up to 5 levels up
- Correctly parses formation ID
- Decrypts secrets.enc with .key
- Loads .muxi config if present
- Returns clear error if not in formation dir

---

### Phase 2: Core Commands (Week 1-2)

#### 2.1 `muxi init` - Formation Scaffolding

**File:** `cmd/init.go`

**Wizard flow:**
```go
func initWizard() error {
    // Prompt for details
    id := promptString("Formation ID", "my-bot")
    name := promptString("Formation name", "My Bot")
    desc := promptString("Description", "")
    template := promptSelect("Select template", []string{
        "Basic (single agent)",
        "Multi-agent (orchestrator + specialists)",
        "Workflow (SOP-driven)",
    })
    
    // Generate formation
    generator.CreateFormation(id, name, desc, template)
    
    // Success message
    fmt.Printf("✓ Created %s/\n", id)
    fmt.Println("\nNext steps:")
    fmt.Printf("  cd %s\n", id)
    fmt.Println("  muxi validate")
    fmt.Println("  muxi deploy --profile <server-profile>")
}
```

**Generated files:**
- `formation.yaml` - Formation config
- `agents/main.yaml` - Basic agent
- `secrets.enc` - Encrypted secrets (ADMIN_KEY, CLIENT_KEY auto-generated)
- `.key` - Encryption key
- `.gitignore` - Auto-configured
- `.muxi` - Optional defaults
- `README.md` - Usage instructions

**Tasks:**
- [ ] Create `cmd/init.go` command
- [ ] Create `pkg/generator/formation.go` - formation scaffolding
- [ ] Add templates (basic, multi-agent, workflow)
- [ ] Generate random ADMIN_KEY and CLIENT_KEY
- [ ] Implement encryption (use runtime's secrets logic)
- [ ] Auto-create `.gitignore` with `.key` excluded
- [ ] Add security warning after creation

**Acceptance criteria:**
- `muxi init` runs wizard
- `muxi init my-bot` creates formation non-interactively
- Generated formation.yaml is valid
- secrets.enc is encrypted
- .gitignore excludes .key
- Security warning displayed

---

#### 2.2 `muxi validate` - Formation Validation

**File:** `cmd/validate.go`

**Tasks:**
- [ ] Create `cmd/validate.go`
- [ ] Require formation context (error if not in dir)
- [ ] Validate formation.yaml against schema
- [ ] Check agents/*.yaml files
- [ ] Check mcp/*.yaml files
- [ ] Check secrets.enc can be decrypted
- [ ] Display validation results

**Acceptance criteria:**
- Works only in formation directory
- Validates all YAML files
- Shows clear error messages
- Returns exit code 0 on success, 1 on failure

---

#### 2.3 `muxi deploy` - Deploy Formation

**File:** `cmd/deploy.go`

**Workflow:**
```go
func deploy() error {
    // 1. Require formation context
    ctx := mustBeInFormationDir()
    
    // 2. Validate
    validate(ctx)
    
    // 3. Get profile
    profile := getProfile() // From --profile flag or .muxi or config
    
    // 4. Package formation
    bundle := packageFormation(ctx)
    
    // 5. Deploy via Server API
    client := NewServerClient(profile)
    result := client.DeployFormation(bundle)
    
    // 6. Show result
    fmt.Printf("✓ Formation '%s' deployed\n", ctx.FormationID)
    fmt.Printf("  URL: %s/api/%s\n", profile.URL, ctx.FormationID)
}
```

**Tasks:**
- [ ] Create `cmd/deploy.go`
- [ ] Implement `packageFormation()` - create .tar.gz
- [ ] Implement Server API client for `/rpc/formations`
- [ ] Add HMAC authentication
- [ ] Show progress during upload
- [ ] Display deployment result

**Acceptance criteria:**
- Works only in formation directory
- Validates before deploying
- Uses profile from --profile or .muxi or config
- Uploads bundle to server
- Shows deployment status

---

#### 2.4 Profile Management Commands

**Files:** `cmd/profile.go`, `cmd/profile_*.go`

**Commands:**
- `muxi profile add <name>` - Wizard to add server profile
- `muxi profile list` - List all profiles
- `muxi profile use <name>` - Set default (prompt for permanent)
- `muxi profile remove <name>` - Remove profile
- `muxi profile current` - Show current profile

**Wizard for `profile add`:**
```bash
muxi profile add production

Server URL: https://muxi.company.com:7890

Authentication:
  Key ID: MUXI_PROD_KEY
  Secret Key: ••••••••••••

Test connection? [Y/n]: y
✓ Connection successful

Set as default profile? [y/N]: y

✓ Profile 'production' added and set as default

To add more servers, edit ~/.muxi/cli/profiles.yaml
```

**Generated profile (single server with id="default"):**
```yaml
production:
  servers:
    - id: default
      url: https://muxi.company.com:7890
      auth:
        key_id: MUXI_PROD_KEY
        secret_key: sk_...
```

**Users can manually edit to add more servers:**
```yaml
production:
  servers:
    - id: us-east-1  # Changed from 'default'
      url: https://api-east.company.com:7890
      auth:
        key_id: MUXI_EAST_KEY
        secret_key: sk_east_...
    - id: us-west-1  # Added manually
      url: https://api-west.company.com:7890
      auth:
        key_id: MUXI_WEST_KEY
        secret_key: sk_west_...
```

**Tasks:**
- [ ] Create `cmd/profile.go` parent command
- [ ] Implement `muxi profile add` wizard
  - Create single server with `id: default`
  - Save in multi-server array format (future-proof)
  - Show hint about editing YAML for more servers
- [ ] Implement `muxi profile list` (table output)
- [ ] Implement `muxi profile use` (prompt for permanent)
- [ ] Implement `muxi profile remove` (with confirmation)
- [ ] Implement `muxi profile current`
- [ ] Add `--auto-detect` flag for localhost

**Acceptance criteria:**
- Wizard prompts for all details
- Tests connection before saving
- Saves to `~/.muxi/cli/profiles.yaml` with servers array
- Profile format is future-proof (supports multi-server when needed)
- `profile use` prompts for permanent vs session
- `profile list` shows table with default marker

---

#### 2.5 Formation Credential Management Commands

**Files:** `cmd/formation_add.go`, `cmd/formation_list_creds.go`

**Commands:**
- `muxi formation add <id>` - Wizard to add formation credentials
- `muxi formation list-creds` - List configured formations
- `muxi formation remove-creds <id>` - Remove credentials

**Wizard for `formation add`:**
```bash
muxi formation add my-bot

Formation ID: my-bot

Connection type:
  1. Direct connection (standalone formation)
  2. Via server (server-managed formation)
  3. Both (hybrid - can use either)
> 1

Formation URL: https://my-bot.company.com

Admin key: fma_...
Client key: fmc_...

Where should credentials be stored?
  1. OS Keychain (secure, recommended)
  2. Encrypted file (portable)
> 1

✓ Formation 'my-bot' added
  URL: https://my-bot.company.com
  Storage: macOS Keychain
```

**Tasks:**
- [ ] Create `cmd/formation_add.go` wizard
- [ ] Implement keychain storage (OS-specific)
  - macOS: Use `security` command
  - Windows: Use Windows Credential Manager API
  - Linux: Use Secret Service API
- [ ] Error if inside formation directory
- [ ] Save reference to `~/.muxi/cli/formations.yaml`
- [ ] Implement `formation list-creds` table view
- [ ] Implement `formation remove-creds` with keychain cleanup

**Acceptance criteria:**
- Wizard guides through all options
- Stores keys in OS keychain
- Errors if in formation directory
- `list-creds` shows all configured formations
- `remove-creds` cleans up keychain entries

---

### Phase 3: Hybrid Commands (Week 2)

#### 3.1 Connection Resolution

**File:** `pkg/client/resolver.go`

**Resolution logic:**
```go
type Connection struct {
    URL            string
    ServerAuth     *HMACAuth     // For Server API
    FormationAuth  *FormationAuth // For Formation API
}

func ResolveConnection(cmd *cobra.Command) (*Connection, error) {
    // 1. Detect formation context
    formationCtx, inDir := DetectFormationContext()
    
    // 2. Get formation ID
    formationID := cmd.Flag("formation").Value.String()
    if formationID == "" && inDir {
        formationID = formationCtx.FormationID
    }
    
    // 3. Get profile (if specified)
    profileName := cmd.Flag("profile").Value.String()
    
    // 4. Resolve connection
    if profileName != "" {
        // Use server (proxied)
        profile := LoadProfile(profileName)
        formationAuth := getFormationAuth(formationID, formationCtx)
        
        return &Connection{
            URL:           fmt.Sprintf("%s/api/%s", profile.URL, formationID),
            ServerAuth:    profile.Auth,
            FormationAuth: formationAuth,
        }, nil
    }
    
    if inDir {
        // Use localhost
        port := formationCtx.Config.Port
        return &Connection{
            URL:           fmt.Sprintf("http://localhost:%d", port),
            FormationAuth: formationCtx.Secrets,
        }, nil
    }
    
    if formationID != "" {
        // Use formations.yaml URL
        formationCreds := LoadFormationCreds(formationID)
        if formationCreds.URL == "" {
            return nil, errors.New("formation has no URL, use --profile")
        }
        return &Connection{
            URL:           formationCreds.URL,
            FormationAuth: formationCreds,
        }, nil
    }
    
    return nil, errors.New("could not determine connection")
}
```

**Tasks:**
- [ ] Implement `ResolveConnection()` logic
- [ ] Implement HMAC authentication for Server API
- [ ] Implement Formation API key authentication
- [ ] Add connection caching
- [ ] Add helpful error messages

**Acceptance criteria:**
- Correctly resolves all 4 connection scenarios
- Uses correct auth for each scenario
- Clear error messages when missing info

---

#### 3.2 Hybrid Command Implementation

**Pattern for all hybrid commands:**
```go
func statusCmd() error {
    // Resolve connection
    conn := ResolveConnection(cmd)
    
    // Call Formation API
    client := NewFormationClient(conn)
    status := client.GetStatus()
    
    // Display
    printStatus(status)
}
```

**Commands to implement:**
- `muxi status` - Formation runtime status
- `muxi logs` - Formation logs
- `muxi chat` - Interactive chat
- `muxi agent list/get/update/delete`
- `muxi secret list/add/update/delete`
- `muxi mcp list/add/get/update/delete`

**Tasks:**
- [ ] Implement Formation API client
- [ ] Implement all hybrid commands
- [ ] Add `--formation` flag to all
- [ ] Test in all 4 connection scenarios
- [ ] Add proper error handling

**Acceptance criteria:**
- All commands work in formation dir (auto-detect)
- All commands work with `--formation` flag
- All commands work with `--profile` flag
- Correct auth used in each case

---

### Phase 4: Wizards & UX (Week 3)

#### 4.1 Interactive Wizards

**Commands with wizards:**
- `muxi init` - Formation creation ✅ (Phase 2)
- `muxi profile add` - Server profile ✅ (Phase 2)
- `muxi formation add` - Formation credentials ✅ (Phase 2)
- `muxi registry login` - Registry authentication
- `muxi agent add` - Agent creation (when in formation dir)
- `muxi mcp add` - MCP server creation (when in formation dir)

**Shared wizard utilities:**
```go
// pkg/wizard/wizard.go
func PromptString(label, defaultValue string) string
func PromptPassword(label string) string
func PromptSelect(label string, options []string) int
func PromptConfirm(label string, defaultYes bool) bool
func PromptMultiSelect(label string, options []string) []int
```

**Tasks:**
- [ ] Create `pkg/wizard/` package
- [ ] Implement prompt utilities (use bubbletea or promptui)
- [ ] Add wizards to remaining commands
- [ ] Add `--non-interactive` flag to skip wizards
- [ ] Add validation to all prompts

**Acceptance criteria:**
- All wizards are user-friendly
- Clear instructions and defaults
- Validation prevents invalid input
- `--non-interactive` flag works
- Ctrl+C cancels gracefully

---

#### 4.2 `profile use` with Prompt

**Enhanced behavior:**
```bash
muxi profile use production

Current default: localhost
Set 'production' as default? [y/N]: y

✓ Using profile 'production' (default: yes)

# Or for session only:
✓ Using profile 'production' (session only)
```

**Tasks:**
- [ ] Implement session-only profile override
- [ ] Prompt for permanent vs session
- [ ] Update config.yaml only if permanent
- [ ] Store session override in environment or memory

**Acceptance criteria:**
- Prompts for permanent vs session
- Session-only works (env var or cache)
- Permanent updates config.yaml

---

### Phase 5: Registry & Cache (Week 3-4)

#### 5.1 Registry Commands (TBD Placeholders)

**Commands:**
- `muxi registry login <url>` - Authenticate
- `muxi registry push` - Push formation/agent (context-aware)
- `muxi registry pull <ref>` - Pull schema
- `muxi registry search <term>` - Search
- `muxi registry list` - List user's schemas

**For now:**
- [ ] Create command stubs
- [ ] Add `TODO: Registry API not yet defined` messages
- [ ] Design expected workflow in comments
- [ ] Wait for registry API spec

**Acceptance criteria:**
- Commands exist but show "coming soon" message
- Basic structure in place for future implementation

---

#### 5.2 Cache Management

**Commands:**
- `muxi cache list` - List cached schemas (was: `muxi local list`)
- `muxi cache prune` - Prune cache (was: `muxi local prune`)

**Cache location:** `~/.muxi/cli/cache/`

**Tasks:**
- [ ] Remove `muxi local` commands
- [ ] Create `muxi cache list`
- [ ] Create `muxi cache prune`
- [ ] Implement cache storage format
- [ ] Add auto-prune on pull

**Acceptance criteria:**
- `muxi local` commands removed
- `muxi cache` commands work
- Cache stored in `~/.muxi/cli/cache/`

---

#### 5.3 Remove `pack`, Auto-pack on Deploy

**Tasks:**
- [ ] Remove `muxi local pack` command
- [ ] Add auto-pack to `muxi deploy`
- [ ] Add auto-pack to `muxi registry push` (when ready)
- [ ] Pack includes all referenced files

**Acceptance criteria:**
- `muxi local pack` command doesn't exist
- `muxi deploy` automatically packs formation
- All referenced files included in bundle

---

### Phase 6: Polish & Testing (Week 4)

#### 6.1 Error Messages

**Improve all error messages:**

```bash
# Bad
Error: formation not found

# Good
✗ Formation 'my-bot' not found

Available formations:
  - support-bot (server-managed)
  - chat-bot (direct: https://chat-bot.company.com)

Add a new formation:
  muxi formation add my-bot
```

**Tasks:**
- [ ] Review all error messages
- [ ] Add helpful hints
- [ ] Show available options
- [ ] Use colored output (red for errors, yellow for warnings)
- [ ] Add exit codes (0=success, 1=error, etc.)

---

#### 6.2 `.muxi` File Support

**Features:**
- Profile default
- Registry default  
- Env overrides

**Tasks:**
- [ ] Load `.muxi` in formation context
- [ ] Apply profile default (if no --profile flag)
- [ ] Apply registry default
- [ ] Apply env overrides (for local dev)
- [ ] Document `.muxi` format in README

**Acceptance criteria:**
- `.muxi` file loaded when in formation dir
- Profile default works
- Env overrides work for local dev
- Documented

---

#### 6.3 Testing

**Unit tests:**
- [ ] Config loading/saving
- [ ] Formation context detection
- [ ] Connection resolution
- [ ] HMAC authentication
- [ ] Keychain storage (mocked)

**Integration tests:**
- [ ] `muxi init` creates valid formation
- [ ] `muxi deploy` uploads to server
- [ ] Hybrid commands work in all scenarios

**Manual testing:**
- [ ] Test all wizards
- [ ] Test all error cases
- [ ] Test on macOS, Linux, Windows

---

## 🚀 Implementation Order

### Sprint 1 (Week 1)
1. Configuration management (#1.1)
2. Global flags (#1.2)
3. Formation context detection (#1.3)
4. `muxi init` (#2.1)
5. `muxi validate` (#2.2)

### Sprint 2 (Week 2)
6. `muxi deploy` (#2.3)
7. Profile management (#2.4)
8. Formation credential management (#2.5)
9. Connection resolution (#3.1)
10. Basic hybrid commands (status, logs) (#3.2)

### Sprint 3 (Week 3)
11. All hybrid commands (#3.2)
12. Wizards & UX improvements (#4.1)
13. `profile use` prompt (#4.2)
14. Cache management (#5.2)
15. Remove pack (#5.3)

### Sprint 4 (Week 4)
16. Registry command stubs (#5.1)
17. Error messages (#6.1)
18. `.muxi` file support (#6.2)
19. Testing (#6.3)
20. Documentation

---

## ✅ Success Criteria

**Phase 1 Complete:**
- [ ] All config files load/save correctly
- [ ] Global flags work with shorthands
- [ ] Formation context detected correctly

**Phase 2 Complete:**
- [ ] Can create formation with `muxi init`
- [ ] Can validate formation
- [ ] Can deploy formation to server
- [ ] Can manage profiles
- [ ] Can add formation credentials

**Phase 3 Complete:**
- [ ] All hybrid commands work
- [ ] Connection resolution works in all scenarios
- [ ] Auth works for both Server and Formation APIs

**Phase 4 Complete:**
- [ ] All wizards are user-friendly
- [ ] `profile use` prompts correctly
- [ ] UX is polished

**Phase 5 Complete:**
- [ ] Cache management works
- [ ] Auto-pack on deploy works
- [ ] Registry stubs in place

**Phase 6 Complete:**
- [ ] All error messages are helpful
- [ ] `.muxi` file support works
- [ ] Tests pass
- [ ] Documentation complete

---

## 📝 Notes

**Deferred:**
- Multi-server support (#5) - Format is ready, implementation deferred
  - Profile format supports multiple servers (array with id, url, auth)
  - For now: Only use first/default server
  - Future: Add `--server <id>` flag and iteration logic
  - No breaking changes needed when implementing

**Dependencies:**
- Runtime's secrets encryption logic (for secrets.enc)
- Server API must be running for testing
- Registry API spec (for registry commands)

**Nice to have:**
- Shell completions (bash/zsh/fish)
- Man pages
- Progress bars for uploads
- Colored diff for changes

---

**Last Updated:** 2025-10-24  
**Status:** Ready to start implementation
