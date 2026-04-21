# AI Incident Assistant

A CLI tool for triaging production incidents on a payment platform. It takes a free-text incident description and returns a structured JSON triage result: category, severity, affected scope, and up to three actionable hypotheses.

## Requirements

- Go 1.25+
- `ANTHROPIC_API_KEY` environment variable (required)
- `ANTHROPIC_MODEL` environment variable (optional, default: `claude-sonnet-4-20250514`)

## Usage

```bash
# Analyze incident text directly
go run ./cmd/incident-assistant/ --text "Payments are failing with 504 errors..."

# Analyze from a file
go run ./cmd/incident-assistant/ --file docs/examples/inc102_db_reporting.txt

# With verbose diagnostics on stderr
go run ./cmd/incident-assistant/ --file docs/examples/inc101_payment_provider.txt --verbose
```

The final triage result is printed to **stdout** as a JSON object. Verbose stage diagnostics go to **stderr** only.

## Example output

```json
{
  "category": "db_degradation_caused_by_reporting",
  "summary": "Reporting service queries saturating shared DB, causing payment timeouts.",
  "affected": "Customers making payments via /payments/create; api-gateway returning 504s.",
  "severity": "high",
  "hypotheses": [
    {
      "title": "Reporting queries starving payment transactions",
      "reasoning": "Long-running reporting queries hold DB locks and exhaust shared connection pool.",
      "next_steps": [
        "Check pg_stat_activity for long-running queries from reporting-service.",
        "Kill offending queries and verify payment response times recover."
      ]
    }
  ]
}
```

## Running tests

```bash
go test ./internal/ -v -short   # recommended: skips tests that require a network call
go test ./internal/ -v          # runs all tests including the orchestrator smoke test (makes one network call)
```

## Architecture

The pipeline is staged and deterministic — the LLM is a controlled component with explicit guardrails, not a single black-box call.

```
incident text
     │
     ▼
 1. Parse          deterministic signal extraction (services, endpoints, keywords, signals, time hints)
     │
     ▼
 2. Load knowledge  data/system_description.json + data/past_incidents.json
     │
     ▼
 3. Select          overlap scoring → top-2 relevant past incidents (skipped if score < threshold)
     │
     ▼
 4. Build prompt    raw text + parsed signals + system description + selected incidents
     │
     ▼
 5. Generate        single Anthropic API call
     │
     ▼
 6. Validate        JSON structure, enums, cardinality, English language
     │
     ├─ valid ──────► stdout
     │
     ▼
 7. Repair          extract JSON from fences/prose; normalize enum casing
     │
     ├─ valid ──────► stdout
     │
     ▼
 8. Regenerate      one retry with refined prompt (includes failure reason + original output)
     │
     ├─ valid ──────► stdout
     │
     └─ still invalid ► controlled error (non-zero exit)
```

## Auxiliary data

`data/system_description.json` — describes the 6 platform services, shared infrastructure (PostgreSQL, ELK), and known operational risk patterns. It is included in the generation prompt as structured context.

`data/past_incidents.json` — 4 structured past incidents derived from the assignment scenarios, with normalized signals and diagnostic hints. The retrieval stage scores overlap between parsed incident signals and past-incident fields; the top-2 matches (if above threshold) are included in the prompt as contextual historical analogies.

## Output contract

| Field | Type | Constraints |
|---|---|---|
| `category` | enum | `external_payment_provider_issue`, `db_degradation_caused_by_reporting`, `notification_delivery_issue`, `user_authentication_errors`, `unknown_or_mixed` |
| `summary` | string | non-empty, English |
| `affected` | string | non-empty |
| `severity` | enum | `low`, `medium`, `high` |
| `hypotheses` | array | 1–3 items |
| `hypotheses[].title` | string | non-empty |
| `hypotheses[].reasoning` | string | non-empty |
| `hypotheses[].next_steps` | array | 2–3 non-empty strings |

## Validation and recovery

The output is validated against the expected response schema.
The implementation handles invalid JSON, unexpected output structure, missing fields, wrong enum values, and unexpected response language, and retries with a refined prompt when validation fails.

1. **Repair**: extract JSON from markdown fences or surrounding prose; normalize enum casing variants (e.g. `"High"` → `"high"`)
2. **Regenerate**: one retry with a refined prompt that includes the validation error and the original invalid output
3. **Controlled error**: if still invalid after regeneration, exit with a non-zero status and a descriptive error message

Recovery budget: 1 generation + 1 repair + 1 regeneration.

## Example incidents

The repository includes four example test incidents in `docs/examples/`.

For each example incident, the expected behavior is:
- correct top-level category
- valid JSON output
- severity within the allowed enum
- 1–3 plausible hypotheses with 2–3 concrete next steps each

| File | Expected category | Key signals |
|---|---|---|
| `inc101_payment_provider.txt` | `external_payment_provider_issue` | PayGate timeouts, card payment failures, timing hints |
| `inc102_db_reporting.txt` | `db_degradation_caused_by_reporting` | high DB CPU, long queries from reporting-service, 504s |
| `inc103_notification.txt` | `notification_delivery_issue` | SMS and email confirmations not delivered, SMTP connection failures |
| `inc104_auth.txt` | `user_authentication_errors` | auth failures across web and mobile, invalid token signatures after key rotation |


## Trade-offs

**Deterministic parser over LLM extraction** — signal extraction (services, endpoints, keywords) uses string matching and regex, not an LLM call. This makes retrieval scoring reproducible and avoids hallucinated service names polluting the context.

**Closed category enum** — the 5-value enum keeps downstream consumers simple. `unknown_or_mixed` is the escape hatch for incidents that don't fit a known pattern.

**Repair before regeneration** — many LLM formatting failures are mechanical (markdown fences, capitalized enums). Fixing these locally avoids a second API call in the common case.

**Static knowledge files** — `data/` files are checked in as JSON. No database, no embeddings, no vector search. This keeps the tool self-contained and auditable.
