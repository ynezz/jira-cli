# AGENTS.md — pfd (prplOS Firmware Downloader)

## RULE 0 — THE FUNDAMENTAL OVERRIDE PREROGATIVE

If I tell you to do something, even if it goes against what follows below, YOU MUST LISTEN TO ME. I AM IN CHARGE, NOT YOU.

---

## RULE 1 — ABSOLUTE (DO NOT EVER VIOLATE THIS)

You may NOT delete any file or directory unless I explicitly give the exact command **in this session**.

- This includes files you just created (tests, tmp files, scripts, etc.).
- You do not get to decide that something is "safe" to remove.
- If you think something should be removed, stop and ask. You must receive clear written approval **before** any deletion command is even proposed.

Treat "never delete files without permission" as a hard invariant.

---

## Rules & Safety

### Irreversible Git & Filesystem Actions

Absolutely forbidden unless I give the **exact command and explicit approval** in the same message:

- `git reset --hard`
- `git clean -fd`
- `rm -rf`
- Any command that can delete or overwrite code/data

### Rules

1. If you are not 100% sure what a command will delete, do not propose or run it. Ask first.
2. Prefer safe tools: `git status`, `git diff`, `git stash`, copying to backups, etc.
3. After approval, restate the command verbatim, list what it will affect, and wait for confirmation.
4. When a destructive command is run, record in your response:
   - The exact user text authorizing it
   - The command run
   - When you ran it

If that audit trail is missing, then you must act as if the operation never happened.

---

### Git Safety Protocol

#### Safe Commands (Always OK)

```bash
git status
git diff
git log
git branch -a
git stash list
git remote -v
```

#### Requires Confirmation

```bash
git checkout <file>      # Discards uncommitted changes
git restore <file>       # Same as checkout
git stash drop           # Permanently deletes stash
git branch -d <branch>   # Deletes local branch
```

#### FORBIDDEN Without Explicit Approval

```bash
git reset --hard         # Discards all uncommitted work
git clean -fd            # Deletes untracked files
git push --force         # Rewrites remote history
rm -rf                   # Recursive delete
```

---

### Approval Format

When you receive approval for a destructive command:

1. Quote the exact approval text
2. State the command you will run
3. List files/directories that will be affected
4. Wait for final confirmation
5. Execute and record the timestamp

Example:

```
User approval: "yes, go ahead and reset hard"
Command: git reset --hard HEAD~1
Affected: All uncommitted changes will be lost
Executing at: 2025-01-15T10:30:00Z
```

---

## Commit Policy

Agents are expected to commit their changes after completing each task or logical unit of work. Do not leave uncommitted changes.

### Git workflow after QA

- Based on your knowledge of the project, commit all changed files in a series of logically connected groupings with super detailed commit messages for each and then push. Take your time to do it right. Don't edit the code at all. Don't commit obviously ephemeral files.
- Do not sign commits. Add my configured sign-off `git commit -s`
- Use `git -c commit.gpgsign=false commit -s -m` to avoid signing
- Avoid one `-m` per wrapped line (that inserts blank lines); use a single body with embedded newlines or a message file instead
- Commit subject should include `prefix: ...` that matches the top-level tree/area being changed
  - Makefile: ...
  - tools: ...
  - AGENTS: ...
- Commit description should include:
  - what is currently wrong/missing
  - why is this change needed
  - "Currently ..."
  - "So lets fix ..." or "So lets add ..." (use the verb that matches the change)
  - Wrap commit description lines to 72 characters
- Add `Co-authored-by: <email> [model]`
  - example `Co-authored-by: codex <codex@openai.com> [gpt-5.2-codex high]`

---

## Issue Tracking with br (beads_rust) & bv

All issue tracking goes through **br**. No other TODO systems.

### Key Invariants

- `.beads/` is authoritative state and **must always be committed** with code changes.
- Do not edit `.beads/*.jsonl` directly; only via `br`.
- `br` is non-invasive and never executes git commands directly. You must manually run git operations after `br sync --flush-only`.

---

### br Basics

#### Check Ready Work

```bash
br ready --json
```

#### Create Issues

```bash
br create "Issue title" -t bug|feature|task -p 0-4 --json
br create "Issue title" -p 1 --deps discovered-from:br-123 --json
```

#### Update Status

```bash
br update br-42 --status in_progress --json
br update br-42 --priority 1 --json
```

#### Complete Work

```bash
br close br-42 --reason "Completed" --json
br close br-42 br-43 --reason "Batch complete"  # Close multiple
```

#### Other Commands

```bash
br list --status=open         # All open issues
br show <id>                  # Full issue details with dependencies
br sync --flush-only          # Export to JSONL (does NOT run git)
```

---

### Types & Priorities

**Types:** `bug`, `feature`, `task`, `epic`, `chore`, `question`, `docs`

**Priorities:**
- `0` critical (security, data loss, broken builds)
- `1` high
- `2` medium (default)
- `3` low
- `4` backlog

---

### Agent Workflow

1. `br ready` to find unblocked work
2. Claim: `br update <id> --status in_progress`
3. Implement + test
4. If you discover new work, create a new bead with `discovered-from:<parent-id>`
5. Close when done: `br close <id>`
6. Commit `.beads/` in the same commit as code changes

#### Sync Workflow

```bash
br sync --flush-only                    # Export to JSONL (does NOT run git)
git add .beads/ && git commit -m "Update beads" && git push
```

---

### bv — Graph-Aware Triage

bv is a graph-aware triage engine for Beads projects. It handles *what to work on* (triage, priority, planning).

**CRITICAL: Use ONLY `--robot-*` flags. Bare `bv` launches an interactive TUI that blocks your session.**

#### The Mega-Command

```bash
bv --robot-triage    # THE ENTRY POINT: start here
```

Returns:
- `quick_ref`: at-a-glance counts + top 3 picks
- `recommendations`: ranked actionable items with scores, reasons, unblock info
- `quick_wins`: low-effort high-impact items
- `blockers_to_clear`: items that unblock the most downstream work
- `project_health`: status/type/priority distributions, graph metrics
- `commands`: copy-paste shell commands for next steps

#### Other Commands

```bash
bv --robot-next          # Minimal: just the single top pick + claim command
bv --robot-plan          # Parallel execution tracks with unblocks lists
bv --robot-priority      # Priority misalignment detection
bv --robot-insights      # Full metrics: PageRank, betweenness, cycles, etc.
bv --robot-alerts        # Stale issues, blocking cascades, priority mismatches
```

#### Filtering

```bash
bv --robot-plan --label backend              # Scope to label's subgraph
bv --recipe actionable --robot-plan          # Pre-filter: ready to work
bv --recipe high-impact --robot-triage       # Pre-filter: top PageRank scores
```

#### Understanding Output

All robot JSON includes:
- `data_hash` — Fingerprint of source beads.jsonl
- `status` — Per-metric state: `computed|approx|timeout|skipped`

Two-phase analysis:
- **Phase 1 (instant):** degree, topo sort, density
- **Phase 2 (async, 500ms timeout):** PageRank, betweenness, cycles — check `status` flags

#### jq Quick Reference

```bash
bv --robot-triage | jq '.quick_ref'                        # At-a-glance summary
bv --robot-triage | jq '.recommendations[0]'               # Top recommendation
bv --robot-plan | jq '.plan.summary.highest_impact'        # Best unblock target
bv --robot-insights | jq '.Cycles'                         # Circular deps (must fix!)
```

---

### Never Do This

- Use markdown TODO lists
- Use other trackers
- Duplicate tracking
- Edit `.beads/*.jsonl` directly

---

## Reusable QA & Release Runbook

### Manual QA Plan (Canonical)

Use [`tests/manual/manual-test-plan.md`](tests/manual/manual-test-plan.md) as
the canonical reusable Jira Cloud verification procedure for
formatting/rendering changes.

Use [`tests/manual/golden-jira-issue.md`](tests/manual/golden-jira-issue.md)
as the static issue-description fixture and
[`tests/manual/golden-jira-issue.json`](tests/manual/golden-jira-issue.json)
as the stable case/check mapping source for manual verification reporting.

When a QA bead asks for manual verification (for example `br-258`), follow
the plan end-to-end and do not invent ad-hoc variants.

Minimum completion artifacts:
- Candidate commit/tag tested
- Jira project + probe issue key
- Forward-path result (`Markdown -> Wiki -> ADF -> Web UI`)
- Reverse-path result (`ADF -> jira issue view --plain/--raw`)
- Follow-up bead IDs for any failures (`discovered-from:<qa-bead-id>`)

### Release Process (Reusable)

For release beads (for example `br-157`), use this reusable sequence:

1. Confirm release blockers and required QA beads are closed.
2. Sync branch state: `git pull --rebase`.
3. Run quality gates for the candidate:
   - `go test -race ./...`
   - `ubs --diff --only=golang`
4. Execute the release bead's version/tag/build steps (project currently uses
   GoReleaser-based release flow).
5. Push commits/tags and verify: `git status` reports up to date with origin.
6. Post a release summary with:
   - version/tag
   - artifact/build output reference
   - linked manual QA result from
     `tests/manual/manual-test-plan.md` template

---

## Session Management

### Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

#### Mandatory Workflow

1. **File issues for remaining work** — Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) — Tests, linters, builds
3. **Update issue status** — Close finished work, update in-progress items
4. **PUSH TO REMOTE** — This is MANDATORY:

```bash
git pull --rebase
br sync --flush-only  # Export to JSONL (does NOT run git commands)
git add .beads/ && git commit -m "Update beads" && git push
git status  # MUST show "up to date with origin"
```

5. **Clean up** — Clear stashes, prune remote branches
6. **Verify** — All changes committed AND pushed
7. **Hand off** — Provide context for next session

#### Critical Rules

- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing — that leaves work stranded locally
- NEVER say "ready to push when you are" — YOU must push
- If push fails, resolve and retry until it succeeds

---

### Multi-Agent Awareness

#### Note for Codex/GPT-5.2 (and all agents)

You may notice unexpected changes in the working tree that you didn't make. These messages are common:

```
Unexpected changes (need guidance)

- Working tree still shows edits I did not make in go.mod, go.sum, internal/...
  Please advise whether to keep/commit/revert these before any further work.
```

**NEVER DO THIS.** The answer is always the same: those are changes created by other agents working on the project simultaneously. This happens multiple times per minute.

#### How to Handle Other Agents' Changes

1. **NEVER** stash, revert, overwrite, or disturb other agents' work
2. Treat those changes identically to changes you made yourself
3. Just fool yourself into thinking YOU made the changes and don't recall it
4. Continue your work normally

#### Merge Conflicts

If you encounter merge conflicts with other agents' work:

1. Pull latest changes: `git pull --rebase`
2. Resolve conflicts preserving both your work and theirs
3. If unsure which version is correct, ask the user
4. Never silently discard others' changes

---

### Handoff Template

When ending a session, provide this information:

```markdown
## Session Handoff

### Completed
- [List completed work]

### In Progress
- [List work started but not finished]
- [Include relevant issue IDs: br-123, br-456]

### Blocked/Needs Attention
- [List blockers or items needing decisions]

### Next Steps
- [Clear actionable items for next session]

### Notes
- [Any context the next agent needs]
```

---

## Tool Quick Reference

| Tool | When to Use | Example |
|------|-------------|---------|
| **br** | Issue tracking | `br ready --json`, `br create "..." --json` |
| **bv** | Graph-aware triage | `bv --robot-triage` (never bare `bv`) |
| **ubs** | Pre-commit bug scan | `ubs $(git diff --name-only --cached)` |
| **ast-grep** | Structural code changes | `ast-grep run -l Go -p 'fmt.Println($$$)'` |
| **rg** | Text search | `rg -n 'func.*Download' -t go` |
| **cass** | Cross-agent search | `cass search "..." --robot --limit 5` |
| **Agent Mail** | Multi-agent coordination | MCP tools: `register_agent`, `send_message`, `fetch_inbox` |

---

## Tool: UBS (Ultimate Bug Scanner)

UBS is the AI coding agent's secret weapon for flagging likely bugs early.

**Golden Rule:** `ubs <changed-files>` before every commit. Exit 0 = safe. Exit >0 = fix & re-run.

---

### Commands

```bash
ubs file.go file2.go                        # Specific files (< 1s) — USE THIS
ubs $(git diff --name-only --cached)        # Staged files — before commit
ubs --only=go internal/                     # Language filter (3-5x faster)
ubs --ci --fail-on-warning .                # CI mode — before PR
ubs --help                                  # Full command reference
ubs sessions --entries 1                    # Tail the latest install session log
ubs .                                       # Whole project (ignores .venv, node_modules)
```

---

### Output Format

```
⚠️  Category (N errors)
    file.go:42:5 – Issue description
    💡 Suggested fix
Exit code: 1
```

**Parse:** `file:line:col` → location | 💡 → how to fix | Exit 0/1 → pass/fail

---

### Fix Workflow

1. Read finding → category + fix suggestion
2. Navigate `file:line:col` → view context
3. Verify real issue (not false positive)
4. Fix root cause (not symptom)
5. Re-run `ubs <file>` → exit 0
6. Commit

---

### Speed Critical

Scope to changed files. Never full scan for small edits.

| Command | Time |
|---------|------|
| `ubs internal/cli/download.go` | < 1s |
| `ubs .` | 30s+ |

---

### Bug Severity

#### Critical (always fix)

- Nil pointer dereference
- Race conditions
- Goroutine leaks
- Unchecked errors

#### Important (production)

- Type narrowing issues
- Division-by-zero potential
- Resource leaks
- Unbounded allocations

#### Contextual (judgment)

- TODO/FIXME comments
- `fmt.Println` debug statements

---

### Go-Specific Examples

```bash
# Check single file after editing
ubs internal/download/client.go

# Check all Go files in directory
ubs --only=go internal/download/

# Check staged changes before commit
ubs $(git diff --name-only --cached | grep '\.go$')

# Full project scan (rare)
ubs --only=go .
```

---

### Anti-Patterns

| Don't | Do Instead |
|-------|------------|
| Ignore findings | Investigate each one |
| Full scan per edit | Scope to changed file |
| Fix symptom: `if x != nil { x.Y() }` | Fix root cause: ensure x never nil at callsite |
| Skip UBS before commit | Always run on staged files |

---

### Integration with pfd Workflow

```bash
# After making changes
go build ./cmd/pfd
go test ./...
ubs $(git diff --name-only --cached)  # Must exit 0
git commit -m "..."
```

---

## Tool: ast-grep vs ripgrep

### When to Use Which

**Use `ast-grep` when structure matters.** It parses code and matches AST nodes, ignoring comments/strings, and can safely rewrite code.

- Refactors/codemods: rename APIs, change patterns
- Policy checks: enforce patterns across a repo
- Structural searches: find all function calls, imports, etc.

**Use `ripgrep` when text is enough.** Fastest way to grep literals/regex.

- Recon: find strings, TODOs, log lines, config values
- Pre-filter: narrow candidate files before ast-grep
- Simple searches: known identifiers, error messages

---

### Rule of Thumb

| Need | Tool |
|------|------|
| Correctness or **applying changes** | `ast-grep` |
| Raw speed or **hunting text** | `rg` |
| Shortlist files, then match/modify | `rg` → `ast-grep` |

---

### Go Examples

#### ast-grep

```bash
# Find all fmt.Println statements
ast-grep run -l Go -p 'fmt.Println($$$)'

# Find all error returns without wrapping
ast-grep run -l Go -p 'return err'

# Find all http.Get calls
ast-grep run -l Go -p 'http.Get($URL)'

# Find function definitions
ast-grep run -l Go -p 'func $NAME($$$) { $$$ }'
```

#### ripgrep

```bash
# Quick textual hunt for function
rg -n 'func.*LoadConfig' -t go

# Find all TODO comments
rg -n 'TODO|FIXME' -t go

# Find specific error string
rg -n 'failed to download' -t go

# Count occurrences
rg -c 'fmt.Println' -t go
```

#### Combined Workflow

```bash
# 1. Find files with sync.Mutex (fast text search)
rg -l -t go 'sync.Mutex'

# 2. Then find structural pattern in those files
rg -l -t go 'sync.Mutex' | xargs ast-grep run -l Go -p 'mu.Lock()'
```

---

### Comparison Table

| Scenario | Tool | Why |
|----------|------|-----|
| "How is robot mode implemented?" | `warp_grep` or explore | Discovery, multiple files |
| "Where is `Download` defined?" | `rg -n 'func Download'` | Known identifier |
| "Replace `var` with `const`" | `ast-grep` | Structural rewrite |
| "Find all fmt.Printf calls" | `ast-grep` | Ignores comments/strings |
| "Find 'TODO' comments" | `rg` | Text search |
| "Rename function across codebase" | `ast-grep` | Safe refactor |

---

### ast-grep Patterns for Go

```bash
# Error handling
ast-grep run -l Go -p 'if err != nil { return err }'
ast-grep run -l Go -p 'if err != nil { $$$ }'

# HTTP calls
ast-grep run -l Go -p 'http.Get($$$)'
ast-grep run -l Go -p 'http.Post($$$)'

# Context usage
ast-grep run -l Go -p 'context.Background()'
ast-grep run -l Go -p 'context.TODO()'

# Defer patterns
ast-grep run -l Go -p 'defer $$.Close()'

# Goroutines
ast-grep run -l Go -p 'go func() { $$$ }()'
```

---

### Performance Notes

- `rg` is 10-100x faster for simple text searches
- `ast-grep` has parsing overhead but gives structural accuracy
- For large codebases, pre-filter with `rg -l` then pipe to `ast-grep`
- Both tools handle large repos well

---

## Tool: cass & cass-memory (Optional)

> **Note:** This tool may not be needed initially for pfd. Include when cross-agent search becomes valuable.

### cass — Cross-Agent Search

`cass` indexes prior agent conversations (Claude Code, Codex, Cursor, Gemini, ChatGPT, etc.) so you can reuse solved problems.

**Rule:** Never run bare `cass` (TUI). Always use `--robot` or `--json`.

#### Commands

```bash
cass health
cass search "authentication error" --robot --limit 5
cass view /path/to/session.jsonl -n 42 --json
cass expand /path/to/session.jsonl -n 42 -C 3 --json
cass capabilities --json
cass robot-docs guide
```

#### Tips

- Use `--fields minimal` for lean output
- Filter by agent with `--agent`
- Use `--days N` to limit to recent history

stdout is data-only, stderr is diagnostics. Exit 0 = success.

Treat cass as a way to avoid re-solving problems other agents already handled.

---

### cass-memory (cm)

The Cass Memory System gives agents effective memory by searching across previous coding agent sessions and extracting useful lessons.

#### Quick Start

```bash
# 1. Check status and see recommendations
cm onboard status

# 2. Get sessions to analyze (filtered by gaps in your playbook)
cm onboard sample --fill-gaps

# 3. Read a session with rich context
cm onboard read /path/to/session.jsonl --template

# 4. Add extracted rules
cm playbook add "Your rule content" --category "debugging"
# Or batch add:
cm playbook add --file rules.json

# 5. Mark session as processed
cm onboard mark-done /path/to/session.jsonl
```

#### Before Complex Tasks

Retrieve relevant context:

```bash
cm context "<task description>" --json
```

Returns:
- **relevantBullets**: Rules that may help with your task
- **antiPatterns**: Pitfalls to avoid
- **historySnippets**: Past sessions that solved similar problems
- **suggestedCassQueries**: Searches for deeper investigation

#### Protocol

1. **START**: Run `cm context "<task>" --json` before non-trivial work
2. **WORK**: Reference rule IDs when following them (e.g., "Following b-8f3a2c...")
3. **FEEDBACK**: Leave inline comments when rules help/hurt:
   - `// [cass: helpful b-xyz] - reason`
   - `// [cass: harmful b-xyz] - reason`
4. **END**: Just finish your work. Learning happens automatically.

#### Key Flags

| Flag | Purpose |
|------|---------|
| `--json` | Machine-readable JSON output (required!) |
| `--limit N` | Cap number of rules returned |
| `--no-history` | Skip historical snippets for faster response |

stdout = data only, stderr = diagnostics. Exit 0 = success.

---

## Tool: Morph Warp Grep (Optional)

> **Note:** This tool is optional. Use when you need AI-powered discovery across the codebase.

### What It Does

Use `mcp__morph-mcp__warp_grep` for "how does X work?" discovery across the codebase.

Warp Grep:
- Expands a natural-language query to multiple search patterns
- Runs targeted greps, reads code, follows imports
- Returns concise snippets with line numbers
- Reduces token usage by returning only relevant slices, not entire files

---

### When to Use

- You don't know where something lives
- You want data flow across multiple files (API → service → schema → types)
- You want all touchpoints of a cross-cutting concern

#### Example

```
mcp__morph-mcp__warp_grep(
  repoPath: "/data/projects/prplos-ci-build-artifacts",
  query: "How does the download caching work?"
)
```

---

### When NOT to Use

| Situation | Use Instead |
|-----------|-------------|
| You know the function/identifier name | `rg` |
| You know the exact file | Just open it |
| You only need yes/no existence check | `rg -l` |

---

### Comparison

| Scenario | Tool |
|----------|------|
| "How is caching implemented?" | warp_grep |
| "Where is `Download` defined?" | `rg` |
| "Replace `var` with `const`" | `ast-grep` |
| "Find all TODO comments" | `rg` |

---

## Tool: MCP Agent Mail

Agent Mail is a multi-agent coordination system available as an MCP server. Do not treat it as a CLI you must shell out to.

If Agent Mail is not available, flag to the user. They may need to start it using the `am` alias or by running:
```bash
cd "<agent_mail_install_dir>/mcp_agent_mail" && bash scripts/run_server_with_token.sh
```

### What Agent Mail Provides

- **Identities** — Register agents with names, track activity
- **Inbox/Outbox** — Message passing between agents
- **Searchable threads** — Conversation history with full-text search
- **File reservations** — Advisory leases to avoid agents clobbering each other
- **Persistent artifacts** — Git-backed, human-auditable

---

### Core Patterns

#### Same Repo Setup

1. **Register identity:**
   ```
   ensure_project(project_key="/abs/path/to/repo")
   register_agent(project_key, program="claude-code", model="opus-4")
   ```

2. **Reserve files before editing:**
   ```
   file_reservation_paths(
     project_key,
     agent_name,
     paths=["internal/**"],
     ttl_seconds=3600,
     exclusive=true
   )
   ```

3. **Communicate:**
   ```
   send_message(..., thread_id="FEAT-123")
   fetch_inbox(project_key, agent_name)
   acknowledge_message(project_key, agent_name, message_id)
   ```

4. **Fast reads (resources):**
   ```
   resource://inbox/{Agent}?project=<abs-path>&limit=20
   resource://thread/{id}?project=<abs-path>&include_bodies=true
   ```

5. **Optional environment:**
   ```bash
   export AGENT_NAME=YourAgent  # For pre-commit guard
   export WORKTREES_ENABLED=1   # If using git worktrees
   export AGENT_MAIL_GUARD_MODE=warn  # During trials
   ```

#### Multiple Repos (Same Product)

**Option A:** Same `project_key` for all; use specific reservations:
```
file_reservation_paths(..., paths=["frontend/**"])
file_reservation_paths(..., paths=["backend/**"])
```

**Option B:** Different projects linked via:
```
macro_contact_handshake(...)
# or
request_contact(...) / respond_contact(...)
```

Use a shared `thread_id` (e.g., ticket key) for cross-repo threads.

---

### Macros vs Granular Tools

**Prefer macros** when speed matters more than fine-grained control:

| Macro | Purpose |
|-------|---------|
| `macro_start_session` | Boot project, register agent, fetch inbox |
| `macro_prepare_thread` | Align with thread, get summary, fetch context |
| `macro_file_reservation_cycle` | Reserve files, optionally auto-release |
| `macro_contact_handshake` | Request contact + optional auto-approve |

**Use granular tools** when you need explicit control over each step.

---

### File Reservations

#### Reserve Files

```
file_reservation_paths(
  project_key="/abs/path",
  agent_name="YourAgent",
  paths=["app/api/*.py", "internal/download.go"],
  ttl_seconds=3600,
  exclusive=true,
  reason="implementing download feature"
)
```

#### Release Files

```
release_file_reservations(project_key, agent_name)
# Or release specific paths:
release_file_reservations(project_key, agent_name, paths=["app/api/*.py"])
```

#### Renew Reservations

```
renew_file_reservations(project_key, agent_name, extend_seconds=1800)
```

#### Handling Conflicts

If you get `FILE_RESERVATION_CONFLICT`:
1. Adjust patterns to non-overlapping paths
2. Wait for expiry
3. Use non-exclusive reservation if you only need to read

---

### Common Pitfalls

| Error | Cause | Fix |
|-------|-------|-----|
| "from_agent not registered" | Agent not registered | Call `register_agent` with correct `project_key` |
| `FILE_RESERVATION_CONFLICT` | Another agent holds the file | Adjust patterns, wait, or use non-exclusive |
| Auth issues with JWT+JWKS | Token mismatch | Bearer token with `kid` matching server JWKS |

---

### Quick Reference

```python
# Start session (macro)
macro_start_session(
  human_key="/abs/path/to/repo",
  program="claude-code",
  model="opus-4",
  task_description="Implementing download feature"
)

# Send message
send_message(
  project_key="/abs/path",
  sender_name="YourAgent",
  to=["OtherAgent"],
  subject="Question about cache",
  body_md="How should we handle cache invalidation?"
)

# Check inbox
fetch_inbox(project_key="/abs/path", agent_name="YourAgent", limit=10)

# Search messages
search_messages(project_key="/abs/path", query="cache AND invalidation")

# Summarize thread
summarize_thread(project_key="/abs/path", thread_id="FEAT-123")
```

---

## Built-in TODO Note

If asked to use built-in TODO functionality, comply. You can use built-in TODOs if explicitly instructed, even when beads is available.
