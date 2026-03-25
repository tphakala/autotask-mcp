# autotask-mcp-go Fix Plan

## Issues to Fix

### 1. Wire Lazy Loading in server.go (Critical)

**Problem:** `Config.LazyLoading` is loaded from env but never used. `buildServer()` always calls `tools.RegisterAll()`, ignoring the `LAZY_LOADING=true` setting. The lazy loading infrastructure in `tools/lazy.go` (RegisterLazyTools, ToolCategories, routing) is complete but unreachable.

**Fix:**
- Change `buildServer(client)` signature to `buildServer(client, lazyLoading bool)`
- Add conditional: if `lazyLoading`, call `tools.RegisterLazyTools(s)` instead of `tools.RegisterAll(s, ...)`
- **Skip MappingCache/PicklistCache initialization when lazy loading** — these are unused in lazy mode and would waste memory / potentially trigger API calls (Gemini review feedback)
- Resources should still be registered regardless of lazy loading mode
- Update all callers: `runStdio`, `runHTTP` (env mode reuse), `runHTTP` (gateway mode per-request)
- Update `server_test.go` to test both modes

**Files:** `server.go`, `server_test.go`

### 2. Add Lazy Loading Drift Detection Test (Medium)

**Problem:** `ToolCategories` in `tools/lazy.go` and `toolDescriptions` are hardcoded. If a new tool is added to `RegisterAll` in the future, a developer might forget to update `lazy.go`, causing the discovery meta-tools to be out of sync with actual tools. (Gemini review feedback)

**Fix:** Add a test in `tools/lazy_test.go` that collects all tool names from the `ToolCategories` map and verifies they match the tools registered by `RegisterAll`. This prevents drift.

**Files:** `tools/lazy_test.go`

### 3. Standardize JSON Field Name Casing (Not Fixing)

**Problem:** Input struct JSON tags mix `companyId` and `companyID`, `ticketId` and `ticketID` inconsistently.

**Decision: Keep as-is.** After review (confirmed by Gemini), the Go implementation faithfully reproduces the TypeScript version's field names. Changing them would break backward compatibility for users migrating from the TypeScript MCP server. MCP clients rely on these JSON schemas; external contract stability takes precedence over internal casing consistency.

### 4. Improve .gitignore (Low)

**Problem:** Only contains `autotask-mcp` binary name.

**Fix:** Add standard entries: `.env*`, `.DS_Store`, `*.log`, `dist/`, `*.swp`.

**Files:** `.gitignore`

### 5. Add Dockerfile (Low)

**Problem:** No container build support. The TypeScript version has Dockerfile + docker-compose.yml.

**Fix:** Add multi-stage Dockerfile (build + scratch/distroless) and docker-compose.yml for HTTP/gateway mode.

**Files:** `Dockerfile`, `docker-compose.yml`

## Out of Scope

- **Migrate Raw operations to typed entities** — blocked on go-autotask generator fixing `Id` vs `ID` field naming (being worked on in another session)
- **Resource handler error handling** — reviewed and confirmed correct; `jsonResult()` errors ARE propagated via the `(*mcp.ReadResourceResult, error)` return
- **Gateway client cleanup** — SDK limitation, already documented with TODO
- **Additional integration tests** — existing 5 tests provide adequate smoke coverage; more can be added incrementally
