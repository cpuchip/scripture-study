# Scratch: Agent-Driven Project Initialization

**Binding problem:** When creating projects in the brain UI, there's no way to tell the system *what the project is about* in enough detail for meaningful initialization. The current `ScaffoldProject()` creates a mechanical directory structure with placeholder comments, producing files nobody reads. The "Initialize" button only works for external projects. What's needed: an instructions field that captures intent, and an agent that reads those instructions to create *meaningful* governance/skills/context files — not just directory trees.

**Created:** 2026-04-08
**Decision from prior session:** Agent-driven with mechanical fallback (user chose this explicitly).

---

## Current State Inventory

### What Exists
- `ScaffoldProject()` in `scaffold.go` — mechanical scaffold, external-only
  - Creates: `.github/`, `.spec/proposals|scratch|memory/`, `docs/`, README.md, thin copilot-instructions.md
  - Runs: `git init`, initial commit, optional `gh repo create`
  - Limitation: `workspace_type != "external"` → returns error
  - Generated copilot-instructions.md has placeholder sections: `<!-- Add ... here -->`
- `POST /api/projects/{id}/scaffold` endpoint in server.go
- Frontend "🚀 Initialize" button — only shows for `workspace_type === 'external'`
- `ScaffoldResult` struct: `project_dir`, `git_inited`, `gh_created`, `error`

### What's Missing
1. **No `init_instructions` field** — the only input is name + description. Not enough for an agent to do meaningful work.
2. **No agent-driven initialization** — ScaffoldProject creates placeholder files, not content.
3. **External-only** — integrated/subfolder projects can't be initialized at all.
4. **No instructions field in create form** — the create form has name + description + emoji. No place to describe what you want.

### Pipeline Agent Capabilities
From `ai/agent.go`: `AgentConfig` supports `Model`, `SystemMessage`, `MCPServers`, `WorkingDir`, `AgentName`, `AllowedWritePaths`, `TokenWarningThreshold`, `PremiumRequestCost`, `SkillDirectories`, `InfiniteSessions`, `CustomAgents`.

Agents use `ai.NewAgent(p.pool.Client(), agentCfg)` → `agent.Ask(ctx, prompt)`. The agent can read/write files through the SDK's built-in tools. WorkingDir controls where file operations happen.

### Model Costs
- Sonnet: 1.0 premium request per Ask() — appropriate for one-shot initialization
- Haiku: 0.33 — too cheap to write good governance docs
- **Recommendation: Sonnet (1.0)** — this is a one-time cost per project

---

## Design Decisions

### D1: init_instructions vs enriching description
**Decision:** Separate `init_instructions TEXT` column. Description is what you show in lists ("Space Center dream business"). Init instructions are what you tell the agent ("Create a TypeScript monorepo with Next.js frontend, Go backend, PostgreSQL. Architecture: microservices with shared auth..."). Different audiences, different purpose.

### D2: Agent initialization for non-external projects
**Decision:** For integrated/subfolder projects, the agent creates a **context file** (the `context_file` field). This is the document that gets injected into pipeline agent system messages. The agent writes a meaningful project context doc based on the instructions. For external projects, the agent does the full scaffold (dirs + git + context + governance).

### D3: Replace scaffold endpoint vs add new
**Decision:** Replace. The scaffold endpoint was Phase 7c stopgap. `POST /api/projects/{id}/initialize` replaces `POST /api/projects/{id}/scaffold`. The mechanical scaffold becomes the fallback path within the new `InitializeProject()` function.

### D4: Initialize button availability
**Decision:** Available for ALL project types. External projects get full workspace scaffold + agent content. Integrated/subfolder projects get an agent-written context file. The button is useful any time init_instructions exist but the project hasn't been initialized.

### D5: Streaming vs one-shot
**Decision:** One-shot for v1. The agent runs, creates files, returns results. No streaming progress. This keeps it simple and consistent with how research/plan agents work.

---

## Implementation Spec

### 1. Schema: `init_instructions` column

**File:** `scripts/brain/internal/store/db.go`

Add `migrateInitInstructions()`:
```go
func (d *DB) migrateInitInstructions() error {
    cols := columnNames("projects")
    if !cols["init_instructions"] {
        _, err := d.db.Exec("ALTER TABLE projects ADD COLUMN init_instructions TEXT")
        return err
    }
    return nil
}
```

Chain from `migrate()`: `migrateProjectWorkspace → migrateInitInstructions`

### 2. Types: Add field

**File:** `scripts/brain/internal/store/types.go`

Add to Project struct:
```go
InitInstructions string `json:"init_instructions,omitempty"` // detailed instructions for agent initialization
```

### 3. CRUD: Wire through

**File:** `scripts/brain/internal/store/db.go`

- `CreateProject`: Add `init_instructions` to INSERT (12 columns now). Use `nullStr(p.InitInstructions)`.
- `GetProject`: Add `init_instructions` to SELECT + scan (NullString). Assign `p.InitInstructions = initInstr.String`.
- `ListProjects`: Same — add to SELECT + scan + assign.
- `UpdateProject`: Add `init_instructions` to UPDATE SET (10 mutable fields now).

### 4. API: Wire through

**File:** `scripts/brain/internal/web/server.go`

- `handleCreateProject`: Add `InitInstructions string` to request struct. Set on `p`.
- `handleUpdateProject`: Add `init_instructions` to partial-update case: `if v, ok := updates["init_instructions"].(string); ok { existing.InitInstructions = v }`

### 5. Pipeline: InitializeProject

**File:** `scripts/brain/internal/pipeline/scaffold.go` (evolve existing file)

New result struct:
```go
type InitResult struct {
    Method     string   `json:"method"`      // "agent" or "mechanical"
    FilesCreated []string `json:"files_created"`
    ProjectDir string   `json:"project_dir,omitempty"`
    GitInited  bool     `json:"git_inited"`
    GHCreated  bool     `json:"gh_created"`
    Error      string   `json:"error,omitempty"`
}
```

New function: `InitializeProject(project *store.Project) (*InitResult, error)`

Logic:
1. Resolve target directory:
   - External: `project.WorkspacePath` (absolute or relative to workspace)
   - Subfolder: `filepath.Join(p.workspace, project.WorkspacePath)`  
   - Integrated: `p.workspace` (project lives in the main workspace)

2. **If `p.pool != nil`** (agent pool available → agent-driven path):
   - Build system message (see below)
   - Build prompt with project details + init_instructions
   - Create agent with Sonnet model, WorkingDir = resolved dir
   - AllowedWritePaths: broad (`.`) since this is initialization
   - For external: also set up dirs + git init BEFORE agent runs (agent needs a directory to write to)
   - Agent.Ask() → agent creates files
   - Parse agent response for summary
   - For external: git add + commit after agent completes
   - Optional gh repo create

3. **If `p.pool == nil`** (mechanical fallback):
   - External: current ScaffoldProject logic (but use init_instructions in the copilot-instructions.md template if available)
   - Non-external: write a simple context file using init_instructions as content

4. Return InitResult

**System message for init agent:**
```
You are a project initialization agent. Your job is to create meaningful, project-specific files based on the instructions provided.

You will be given:
- A project name and description
- Initialization instructions describing the project's purpose, tech stack, and goals
- The workspace type (external = standalone repo, subfolder = within parent workspace, integrated = shares parent workspace)

Based on the workspace type, create appropriate files:

For EXTERNAL projects:
- .github/copilot-instructions.md — Project identity, architecture, tech stack, conventions. NOT a template with placeholders — write real content based on the instructions.
- README.md — Clear project description with getting started.
- .spec/memory/identity.md — Project purpose and values.
- Any other files the instructions suggest (e.g., docs/architecture.md).

For SUBFOLDER or INTEGRATED projects:
- A project context file (will be specified in the prompt) that captures project purpose, conventions, and architecture for injection into agent prompts.

Guidelines:
- Write real content, not placeholders. If the instructions say "TypeScript monorepo with Next.js", write actual TypeScript conventions.
- Reference the parent workspace (scripture-study) for base governance where appropriate.
- Keep files concise but substantive.
- Don't create files that aren't useful yet. A README and copilot-instructions.md are always useful. A CONTRIBUTING.md for a solo project is not.
```

**Prompt template:**
```
Initialize the project "{{name}}".

Description: {{description}}

Instructions:
{{init_instructions}}

Workspace type: {{workspace_type}}
{{if context_file}}Write the project context to: {{context_file}}{{end}}

Create the appropriate files for this project. Be specific to the instructions — this is not a generic template.
```

### 6. API: Replace scaffold endpoint

**File:** `scripts/brain/internal/web/server.go`

Replace route:
```go
// Replace: s.mux.HandleFunc("POST /api/projects/{id}/scaffold", ...)
s.mux.HandleFunc("POST /api/projects/{id}/initialize", s.cors(s.handleInitializeProject))
```

New handler: `handleInitializeProject` — same shape as handleScaffoldProject but calls `s.pipeline.InitializeProject(project)` and returns `InitResult`.

### 7. Frontend: api.ts

**File:** `scripts/brain/frontend/src/api.ts`

- Add `init_instructions?: string` to Project interface
- Add `init_instructions?: string` to createProject params
- Add `'init_instructions'` to updateProject Pick type
- Replace `scaffoldProject` with:
```typescript
initializeProject(id: number) {
    return request<InitResult>(`/projects/${id}/initialize`, { method: 'POST' })
}
```

### 8. Frontend: ProjectsView.vue (create form)

Add `init_instructions` textarea below description:
```vue
<textarea
  v-model="newInitInstructions"
  placeholder="Initialization instructions (optional) — tech stack, architecture, conventions, goals..."
  rows="3"
  class="w-full bg-gray-950 border border-gray-700 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-sky-500 resize-none"
/>
```

Pass `init_instructions: newInitInstructions.value.trim() || undefined` in createProject call.

### 9. Frontend: ProjectDetailView.vue

- Replace `doScaffold` → `doInitialize`, call `api.initializeProject(project.value.id)`
- Remove `v-if="project.workspace_type === 'external'"` from Initialize button (show for all)
- Update scaffoldResult type to match InitResult
- Add `init_instructions` to edit form (textarea in workspace settings section)

---

## Costs & Risks

| Item | Assessment |
|------|------------|
| Premium requests | 1.0 per init (Sonnet). One-time cost per project. Negligible. |
| Agent reliability | Sonnet is reliable for one-shot structured file creation. Low risk. |
| Fallback path | Mechanical scaffold ensures init works even without SDK. |
| Schema migration | Single column add. Zero-downtime in SQLite. |
| Breaking change | Scaffold endpoint replaced. Frontend updated in same release. |
| Scope creep risk | LOW — this is well-bounded. 7 files touched, all patterns established. |

---

## Verification Checklist

- [ ] `go vet ./...` clean
- [ ] `npx vue-tsc --noEmit` clean  
- [ ] `go test ./...` all pass
- [ ] Create project with init_instructions → field persists through create/edit/get
- [ ] Initialize external project → agent creates meaningful copilot-instructions.md (not placeholders)
- [ ] Initialize integrated project → context file created
- [ ] Initialize without agent pool → mechanical fallback works
- [ ] Initialize button visible for all workspace types
- [ ] Old scaffold endpoint removed, no 404s
