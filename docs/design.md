# AI Incident Assistant — Design

## 1. Overview

This project implements a compact incident triage assistant for on-call engineers.

Given a free-form incident description, the system produces a structured triage result that includes:
- incident category
- short summary
- likely affected users or components
- severity
- up to 3 hypotheses
- concrete next diagnostic steps

The implementation is intentionally small and explicit. The LLM is used as a constrained reasoning component inside a deterministic orchestration pipeline, not as a single black-box function.

## 2. Design Goals

The design aims to satisfy the assignment requirements while staying compact and readable.

Key goals:
- make orchestration stages explicit in code
- use auxiliary context, not just raw user input
- return machine-readable validated output
- show a visible recovery strategy for model failures
- keep the implementation small enough for a take-home task

## 3. Alternatives Considered

### Option A — Single LLM call

Flow:
incident text -> one prompt -> final JSON

Pros:
- fastest to implement
- minimal code

Cons:
- too close to a single black-box LLM function
- weak visibility of orchestration stages
- harder to demonstrate auxiliary-data usage clearly
- weaker control over malformed output and recovery

Decision:
Rejected because the assignment explicitly asks for staged processing and an agent-style architecture.

### Option B — Staged orchestrator with static context retrieval

Flow:
parse input -> select context -> generate structured answer -> validate -> repair/retry

Pros:
- directly matches the assignment requirements
- compact enough for the time budget
- easy to explain in README
- keeps deterministic control around the LLM
- simple to test with example incidents

Cons:
- retrieval is intentionally lightweight
- does not optimize for ambiguous or large-scale knowledge bases

Decision:
Chosen as the best trade-off for the take-home.

### Option C — Multi-agent framework with specialized agents

Flow:
parser agent -> retrieval agent -> triage agent -> validator agent

Pros:
- can look more “agentic”
- extensible for larger systems

Cons:
- unnecessary framework overhead for this task
- higher implementation cost
- adds ceremony without improving the signal much

Decision:
Rejected due to time budget and readability goals.

### Option D — Framework-based orchestration (e.g. LangGraph)

Flow:
framework graph -> LLM generation -> validation -> retry

Pros:
- fast to bootstrap
- compact prototype
- convenient state transitions

Cons:
- easier to collapse parsing, context usage, and generation into one prompt-centric step
- weaker visibility of deterministic orchestration stages
- adds framework dependency without materially improving the signal for this take-home

Decision:
Rejected in favor of a custom minimal orchestrator, because the assignment emphasizes explicit staged architecture, auxiliary-data usage, validation, and recovery more than framework usage.

## 4. Chosen Architecture

The chosen architecture is an explicit orchestration pipeline with five stages:

1. `Parser`
2. `ContextBuilder`
3. `LLMCaller`
4. `Validator`
5. `Repair/Retry`

This is the smallest design that still keeps the LLM inside a controlled flow and makes all required stages visible in code.

## 5. Component Responsibilities

### 5.1 Parser

The parser accepts raw incident text and extracts lightweight normalized signals such as:
- mentioned services
- likely symptoms
- error-related keywords
- timing hints

Examples of signals:
- `payment-service`
- `timeout`
- `401`
- `invalid token signature`
- `504 Gateway Timeout`
- `starting from 12:05 UTC`

This stage is deterministic and intentionally simple. It uses lightweight matching rules rather than ML.

### Parser Output Shape

The parser produces a compact normalized signal object with the following fields:
- `services`
- `endpoints`
- `keywords`
- `signals`
- `time_hints`

Example shape:

```json
{
  "services": ["payment-service"],
  "endpoints": ["/payments/create"],
  "keywords": ["timeout", "paygate", "high cpu"],
  "signals": [
    "customers cannot pay by card",
    "504 gateway timeout"
  ],
  "time_hints": ["12:05 UTC"]
}
```

The parser does not infer category, severity, or affected users.  
Its role is signal normalization, not reasoning.


### 5.2 ContextBuilder

The context builder loads and prepares auxiliary context from static knowledge:
- system description
- past incidents
- parsed incident signals

Its goal is to give the LLM the minimum relevant background needed for classification and hypothesis generation.

Relevance is determined using lightweight heuristics such as:
- service-name overlap
- keyword overlap
- symptom overlap

The ContextBuilder does not inject all historical incidents into the prompt.

Instead, it:
1. loads all past incident records
2. scores them against parsed incident signals
3. selects the top 1-2 most relevant incidents
4. passes only those compact summaries into the generation context

Relevance is determined using simple deterministic matching, such as:
- service-name overlap
- endpoint overlap
- keyword overlap
- symptom/signal overlap

### 5.2.2 Past Incident Selection

The ContextBuilder selects up to 2 most relevant past incidents using deterministic overlap scoring across:

* services  
* endpoints  
* keywords  
* signals  
* provider or protocol terms

Recommended scoring:

* service overlap: +3  
* endpoint overlap: +3  
* keyword overlap: +2  
* provider/protocol mention overlap: +2  
* signal overlap: +1

If relevance is too low, no past incident is injected and generation proceeds with system description and the raw incident only.

This keeps retrieval explicit, deterministic, and compact.

### 5.2.3 Category Enum

The structured response uses the following canonical category values:

* `external_payment_provider_issue`  
* `db_degradation_caused_by_reporting`  
* `notification_delivery_issue`  
* `user_authentication_errors`  
* `unknown_or_mixed`

These values are stable, machine-readable identifiers used by validation and recovery.


### 5.3 LLMCaller

The LLM stage generates the structured triage result:
- category
- summary
- affected
- severity
- hypotheses
- diagnostic next steps

The caller requests structured JSON output and includes:
- task instructions
- allowed categories and field expectations
- selected auxiliary context
- current incident text

The LLM is treated as useful but unreliable. Its output is never trusted without validation.

### 5.4 Validator

The validator checks that the model output matches the expected output contract.

It validates:
- schema and required fields
- enum values
- cardinality constraints
- expected language
- JSON-only output

### 5.5 Repair / Retry

If validation fails, the system performs an explicit recovery step.

Possible recovery actions:
- repair for near-valid mechanical issues
- retry with a refined regeneration prompt
- return a controlled error if recovery fails

This recovery path is explicit in code and is part of the intended design, not a hidden exception-handling detail.

## 6. Data Flow

The end-to-end flow is:

1. Receive incident description
2. Parse it into lightweight normalized signals
3. Load system description and past incidents
4. Select the most relevant auxiliary context
5. Build generation prompt
6. Call the LLM
7. Parse and validate the returned output
8. If invalid, run repair/retry
9. Return final structured JSON or a controlled error

## 7. Knowledge Model

The solution uses local static knowledge files.

### 7.1 System Description

A compact representation of the simplified payment platform:
- `api-gateway`
- `auth-service`
- `payment-service`
- `billing-service`
- `notification-service`
- `reporting-service`

General notes from the assignment are also included, such as:
- centralized log storage
- PostgreSQL usage
- provider-related payment errors
- notification provider degradation
- reporting-service DB pressure

### 7.2 Past Incidents

Historical incidents from the assignment are stored as structured local records.

They are used as:
- retrieval context
- domain grounding
- compact historical analogies for classification and hypothesis generation

### 7.3 Past Incident Record Shape

Past incidents are stored as local structured records in `data/past_incidents.json`.

Each record preserves the original incident summary and may include normalized retrieval fields such as:
- `services`
- `endpoints`
- `signals`
- `keywords`
- `likely_affected`

These derived fields are used only for context selection and prompt construction.
They do not introduce new historical incidents or unsupported facts.

This keeps retrieval deterministic and avoids relying on long raw incident prose during prompt construction.

## 8. Output Contract

The system returns machine-readable JSON with a structure similar to:

```json
{
  "category": "string",
  "summary": "string",
  "affected": "string",
  "severity": "low|medium|high",
  "hypotheses": [
    {
      "title": "string",
      "reasoning": "string",
      "next_steps": ["string", "string"]
    }
  ]
}
```

Recommended category enum:

* `external_payment_provider_issue`  
* `db_degradation_caused_by_reporting`  
* `notification_delivery_issue`  
* `user_authentication_errors`  
* `unknown_or_mixed`

This keeps the output predictable and easy to validate.

## 9. Prompting Strategy

### 9.1 Generation Prompt

The main prompt includes:

* system role  
* concise task instructions  
* expected output fields  
* allowed severity values  
* expected category values  
* selected system description  
* selected past incidents  
* parsed incident signals  
* raw incident text  
* instruction to return JSON only

### 9.2 Regeneration Prompt

The regeneration prompt is narrower and more constrained.

It includes:

* the original invalid model output  
* the validation failure reason  
* the expected schema  
* instruction to return corrected JSON only

The repair prompt exists to demonstrate an explicit recovery strategy.

## 10. Validation and Recovery

The model output is never trusted directly.

After each generation attempt, the response is validated against the expected output contract. The validator checks:

- the top-level response is a JSON object
- required fields are present:
  - `category`
  - `summary`
  - `affected`
  - `severity`
  - `hypotheses`
- `severity` is one of: `low`, `medium`, `high`
- `hypotheses` is an array with at most 3 items
- each hypothesis contains:
  - `title`
  - `reasoning`
  - `next_steps`
- each `next_steps` list contains 2 to 3 non-empty strings
- the response language matches the expected language
- the output contains JSON only, without unsupported surrounding prose

### Recovery Policy

The CLI performs:

1. one normal generation attempt
2. validation
3. if validation fails:
   - try a small deterministic repair for near-valid mechanical issues
   - otherwise perform one refined regeneration attempt
4. if validation still fails:
   - return a controlled CLI error

The retry count is intentionally limited to one refined regeneration attempt to keep the behavior compact and deterministic.

### Repair vs Regeneration

Repair is used only for near-valid mechanical issues, for example:
- markdown fences around JSON
- extra surrounding text before or after the JSON object
- trivial enum normalization such as `"High"` -> `"high"`

Regeneration is used for structural or semantic failures, for example:
- invalid JSON that cannot be safely extracted
- missing required fields
- wrong top-level shape
- wrong field types
- wrong language
- wrong enum values that are not safely normalizable
- too many hypotheses
- invalid `next_steps` cardinality

### Recovery Decision Table

| Failure case | Validation result | Recovery action |
|---|---|---|
| Markdown fences around JSON | JSON payload is present but wrapped | Repair |
| Extra prose before/after JSON | JSON object can be extracted | Repair |
| Enum normalization issue | Schema almost valid, enum mismatch | Repair |
| Missing required field | Output contract not satisfied | Regenerate with refined prompt |
| Wrong top-level shape | Response schema mismatch | Regenerate with refined prompt |
| Wrong field type | Type validation failed | Regenerate with refined prompt |
| Wrong language | Language validation failed | Regenerate with refined prompt |
| Invalid JSON syntax | JSON parsing failed | Repair if clearly extractable, otherwise regenerate |
| Too many hypotheses | Cardinality constraint failed | Regenerate with refined prompt |
| Wrong `next_steps` count | Cardinality constraint failed | Regenerate with refined prompt |
| Retry also fails | Recovery exhausted | Return controlled CLI error |

### Refined Regeneration Prompt

When regeneration is needed, the second prompt is narrower and more constrained than the initial generation prompt.

It includes:
- the reason validation failed
- the expected output schema
- enum restrictions
- the instruction to return JSON only
- the instruction to keep the response in the expected language

This makes the recovery path explicit and demonstrates that malformed model output is handled intentionally rather than ignored.

## 11. Chosen Scope

This implementation intentionally focuses on the smallest architecture that still satisfies the assignment requirements.

### Included in scope

#### Explicit staged orchestration

The solution uses an explicit orchestration pipeline:

* `Parser`  
* `ContextBuilder`  
* `LLMCaller`  
* `Validator`  
* `Repair/Retry`

#### Deterministic parser

The parser extracts simple normalized signals from the incident text.

#### Auxiliary context usage

The generation stage uses:

* system description  
* past incidents  
* parsed signals

#### Structured JSON generation

The LLM produces a typed, machine-readable incident triage response.

#### Validation and recovery

Model output is validated and retried or repaired if needed.

#### Single simple interface

The project exposes a CLI interface only.

Why:
- explicitly allowed by the assignment
- fastest way to demonstrate the orchestration flow
- keeps the solution focused on parsing, context selection, validation, and recovery

#### Minimal documentation and checks

The repository includes:
- minimal run instructions
- a short architecture summary
- example incidents
- a short trade-offs section
- a very small set of checks aligned with the take-home scope

## 12. Out of Scope

The following are intentionally excluded from this solution:

* real integrations with ELK, PostgreSQL, Grafana, or provider APIs  
* persistent storage or a database  
* vector search, embeddings, or semantic retrieval  
* advanced UI  
* authentication or multi-user support  
* streaming responses  
* heavy agent frameworks  
* production-grade observability infrastructure  
* enterprise boilerplate such as DI containers or excessive abstraction layers

These exclusions are intentional and consistent with the task format and time budget.

## 13. Trade-offs

### Static matching instead of embeddings

Context retrieval uses lightweight keyword and service-name matching instead of embeddings or vector search.

Why:

* the domain is small and fixed  
* the assignment provides a compact system description and a few known historical incidents  
* this keeps the solution small and easy to explain

With more time:

* replace retrieval with embedding-based similarity search for broader or noisier incident sets

### One main LLM generation step inside a staged pipeline

The design uses one main generation step rather than multiple specialized model agents.

Why:

* the assignment requires explicit stages, not necessarily multiple LLM agents  
* one generation step is enough for this scale  
* this minimizes complexity while preserving architectural clarity

With more time:

* explore specialized generation, repair, or ranking stages for broader domains

### Static files instead of a database

Knowledge is stored in local static files.

Why:

* the tool is stateless  
* persistence is not required  
* local files reduce setup overhead

With more time:

* move knowledge sources into a small datastore or retrieval service

### Minimal interface over full product surface

The design uses only CLI or minimal HTTP API, not both.

Why:

* the assignment allows choosing one interface  
* the main signal is in orchestration and validation, not presentation  
* this keeps the implementation aligned with the time budget

### Minimal checks over broad automated coverage

The project prioritizes a very small set of deterministic checks and documented example scenarios over a broad automated test suite.

Why:
- the take-home explicitly asks for documented incident examples in the README
- the implementation is intentionally scoped to stay compact
- time is better spent demonstrating the core pipeline and recovery behavior

Included:
- one small parser check
- one small validator check
- one orchestrator smoke check

With more time:
- add a slightly broader suite for parser, retrieval, and validator edge cases

### Structured past incidents instead of raw examples

Historical incidents are stored as compact structured JSON records rather than copied into prompts as long free-form text.

Why:
- easier deterministic matching
- easier prompt construction
- more visible separation between retrieval and generation

With more time:
- retrieval could evolve from static matching to embedding-based similarity search

### Custom orchestrator instead of framework-based graph

The design uses a small custom orchestrator rather than an agent framework.

Why:
- explicit stages are easier to see in code
- deterministic control over parsing, context selection, validation, and retry is clearer
- this keeps the solution closer to the assignment requirements and avoids unnecessary framework ceremony

With more time:
- a graph-based framework could be reasonable if the workflow became more dynamic or required multiple branching agent steps

## 14. Operational Visibility

The implementation may include minimal optional verbose stage logs such as:
- parsing started/completed
- selected past incidents
- validation failed
- retry triggered

These logs are intended only to make the orchestration flow visible during local runs.
They are not meant to be a full observability layer.

## 15. Why This Design Fits the Assignment

This design intentionally optimizes for:

* explicit orchestration over hidden magic  
* deterministic guardrails around model output  
* clear use of auxiliary data  
* compact and readable implementation  
* visible engineering trade-offs

It is not intended to be a full production incident-management platform.  
It is intended to be a clear, credible, and scoped demonstration of controlled LLM usage in a service architecture.

It intentionally favors assignment fit and clarity over completeness within the available time budget.
