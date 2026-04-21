# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Test

```bash
go build ./...                          # build everything
go run ./cmd/incident-assistant/ --text "incident description"
go run ./cmd/incident-assistant/ --file path/to/incident.txt
go run ./cmd/incident-assistant/ --text "..." --verbose   # diagnostics to stderr
go test ./internal/ -v                  # run all tests
go test ./internal/ -run TestParse -v   # run a single test
```

Requires `ANTHROPIC_API_KEY` env var. Optional `ANTHROPIC_MODEL` (default: `claude-sonnet-4-20250514`).

## Architecture

Staged orchestration pipeline — the LLM is a controlled component inside deterministic guardrails, not a single black-box call.

**Pipeline stages** (in `internal/orchestrator.go` `Run()`):
1. **Parse** (`parser.go`) — deterministic signal extraction: services, endpoints, keywords, signals, time hints. Does NOT infer category, severity, or affected users.
2. **Load knowledge** (`knowledge.go`) — reads `data/system_description.json` and `data/past_incidents.json`
3. **Select incidents** (`retrieval.go`) — deterministic overlap scoring, top-2 selection, skips injection if relevance too low
4. **Build prompt** (`context.go`) — assembles generation prompt from raw text + parsed signals + system description + selected incidents
5. **Generate** (`llm.go`) — single Anthropic API call via `anthropic-sdk-go`
6. **Validate** (`validator.go`) — checks JSON structure, enums, cardinality, language
7. **Repair** (`repair.go`) — fixes markdown fences, extracts JSON from prose, normalizes enums
8. **Regenerate** — one retry with refined prompt if repair insufficient; controlled error if that also fails

**Output contract** (`types.go`): `TriageResult` with `category` (5-value enum), `summary`, `affected`, `severity` (low/medium/high), `hypotheses` (max 3, each with 2-3 `next_steps`).

## Key Constraints

- Category enum is closed: `external_payment_provider_issue`, `db_degradation_caused_by_reporting`, `notification_delivery_issue`, `user_authentication_errors`, `unknown_or_mixed`
- Recovery budget: 1 generation + 1 repair attempt + 1 regeneration attempt, then fail
- Final JSON goes to stdout; diagnostics only to stderr when `--verbose`
- Auxiliary knowledge lives in `data/` as static JSON files

## Workflow Rules

- Work strictly one task at a time from `docs/tasks.md`
- Before starting a task: state which task, files to touch, expected result
- After finishing a task: explain what changed, verify against spec/design/decisions, run checks, propose commit
- Do not continue to the next task without explicit user approval
- Use small logical commits grouped by change intent, not by file type
- Commit message format: `feat(scope): concise description` (conventional commits)
- Document priority: `input-task.md` > `spec.md` > `implementation_decisions.md` > `design.md` > `tasks.md`
- Treat `implementation_decisions.md` as fixed — do not reopen settled choices
- Prefer the smallest implementation that satisfies the assignment
- Do not add frameworks, persistence, embeddings, HTTP API, UI, or unnecessary abstractions


Execution protocol:

Work strictly one task at a time.

Rules:
1. Do not implement multiple top-level tasks from tasks.md in one step.
2. At any moment, work on exactly one current task or one clearly named subtask of the current task.
3. Before making changes, state:
   - which task you are starting
   - what files you expect to touch
   - what the expected observable result is
4. After finishing that task, stop and do not continue automatically to the next one.
5. After finishing that task, do all of the following before moving on:
   - explain what changed
   - verify that the result matches spec.md, design.md, and implementation_decisions.md
   - run the smallest relevant checks
   - summarize check results
   - propose a commit plan for this task only
6. Commit only the changes for the completed logical task.
7. Use small logical commits. Group by change intent, not by file type.
8. If changes include unrelated cleanup, propose splitting them into separate commits.
9. After the commit, stop and wait for my approval before starting the next task.

Important:
- Do not “batch implement” multiple tasks at once.
- Do not continue to the next task automatically.
- Prefer the smallest complete increment over broad progress.

Commit protocol:

After completing one task:
1. Run git status and inspect the diff.
2. Summarize the logical change group.
3. Propose exactly one commit for that completed task, unless the diff clearly contains unrelated changes.
4. If unrelated changes exist, propose split commits before committing.
5. Use a conventional commit message with a clear scope.

Commit message examples:
- feat(cli): add input loading and flag validation
- feat(parser): extract services, keywords, and time hints
- feat(retrieval): select top-2 relevant past incidents
- feat(validation): enforce output contract and recovery flow
- docs(readme): add usage and example incidents

For each task, respond using this format:

Current task:
- <task name>

Plan:
- <1-3 bullets>

Files to change:
- <file list>

Expected result:
- <observable behavior>

After implementation, report:
- what changed
- what checks were run
- check results
- proposed commit message
- waiting for approval

Hard stop rule:
After each completed task and proposed commit, stop and wait for my explicit confirmation before proceeding.

Start with Task 1 only.
Do not touch any other task.
When Task 1 is complete, validate it, propose one commit, and stop.
