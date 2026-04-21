# Implementation Decision Log

## D1. Interface choice

The tool is implemented as a CLI.

Rationale:
- explicitly allowed by the assignment;
- keeps the project focused on orchestration, validation, and recovery;
- avoids unnecessary HTTP/UI boilerplate within the ~3 hour scope.

Decision:
- support `--text` and `--file`
- print final JSON to stdout
- print diagnostic logs only in verbose mode

---

## D2. Orchestration style

The solution uses a small custom orchestrator instead of an agent framework.

Rationale:
- makes parsing, context selection, generation, validation, and recovery explicit in code;
- keeps deterministic guardrails visible;
- better matches the assignment than a framework-centric or prompt-centric design.

Decision:
- explicit stages:
  1. parse input
  2. load auxiliary data
  3. select relevant past incidents
  4. generate structured answer
  5. validate
  6. repair/regenerate if needed

---

## D3. Knowledge representation

System description and past incidents are stored as local structured JSON files.

Rationale:
- keeps auxiliary knowledge separate from prompts and code;
- makes retrieval deterministic and inspectable;
- avoids unnecessary persistence or infrastructure.

Decision:
- `data/system_description.json`
- `data/past_incidents.json`

---

## D4. Parser scope

The parser is a deterministic signal normalizer, not a classifier.

Rationale:
- the assignment requires explicit parsing as a separate stage;
- deterministic extraction is sufficient for this task;
- keeps reasoning responsibilities with the LLM.

Decision:
the parser returns:
- `services`
- `endpoints`
- `keywords`
- `signals`
- `time_hints`

The parser does not infer:
- category
- severity
- affected users

---

## D5. Past-incident retrieval

The ContextBuilder selects up to 2 relevant past incidents using deterministic overlap scoring.

Rationale:
- satisfies the auxiliary-data requirement without embeddings or vector search;
- avoids dumping all incidents into the prompt;
- keeps context compact and relevant.

Decision:
- top-k = 2
- score based on:
  - service overlap
  - endpoint overlap
  - keyword overlap
  - provider/protocol overlap
  - signal overlap
- if relevance is too low, inject no past incidents

---

## D6. Output contract

The final result is machine-readable JSON with a stable validated structure.

Rationale:
- explicitly required by the assignment;
- improves validator clarity;
- makes recovery and testing easier.

Decision:
top-level fields:
- `category`
- `summary`
- `affected`
- `severity`
- `hypotheses`

Category enum:
- `external_payment_provider_issue`
- `db_degradation_caused_by_reporting`
- `notification_delivery_issue`
- `user_authentication_errors`
- `unknown_or_mixed`

Validation rules:
- `severity` in `low|medium|high`
- up to 3 hypotheses
- each hypothesis must contain `title`, `reasoning`, `next_steps`
- each `next_steps` list must contain 2–3 non-empty strings

---

## D7. Recovery policy

Recovery is explicit and limited.

Rationale:
- the assignment explicitly asks for a recovery strategy;
- a small, deterministic policy is enough for this take-home;
- avoids infinite retry loops or excessive complexity.

Decision:
- one normal generation attempt
- validate output
- if validation fails:
  - repair near-valid mechanical issues when safe
  - otherwise do one refined regeneration attempt
- if validation still fails:
  - return a controlled CLI error

Repair examples:
- markdown fences around JSON
- extra text around JSON
- trivial enum normalization

Regeneration examples:
- missing required fields
- wrong top-level shape
- wrong field types
- wrong language
- unrecoverable invalid JSON
