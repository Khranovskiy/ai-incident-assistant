# AI Incident Assistant — Spec

## 1. Problem

On-call engineers need a fast way to triage production incidents from a short free-form incident description.

The tool should use an LLM as a controlled component inside a small service architecture, not as a single black-box function.

## 2. Goal

Build a compact service that accepts an incident description and returns a structured incident triage result:
- incident category
- short summary
- likely affected users or components
- severity
- up to 3 hypotheses
- concrete next diagnostic steps

## 3. Input

A free-form incident description in English.

Example input:
- customer symptoms
- affected endpoint or service
- log snippets or observations
- metrics hints
- timing information

## 3.1 Auxiliary Knowledge

The solution uses structured static auxiliary knowledge for:
- system description
- past incidents

Past incidents are represented as structured auxiliary records derived from the assignment examples.

The implementation may enrich them with normalized fields (for example, services, keywords, or signals) to support deterministic context selection, while preserving the original incident meaning.

## 4. Output

Machine-readable JSON with a stable structure.

The result must include:
- incident category;
- a short summary of what is happening;
- who is likely affected;
- severity (`low`, `medium`, or `high`);
- up to 3 hypotheses;
- for each hypothesis, 2–3 concrete diagnostic next steps.

Example schema shape:

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

## 5. Functional Requirements

1. Accept incident description text.  
2. Parse the incident text into normalized signals.  
3. Use auxiliary context:  
   - system description  
   - past incidents  
4. Generate a structured triage result with the LLM.  
5. Validate the result against a JSON schema or equivalent typed structure.  
6. Recover from typical model failures:  
   - invalid JSON  
   - response does not match the expected output schema  
   - missing fields  
   - wrong enum values  
   - unexpected language  
7. Return a final valid structured response or a clear error.
8. Ensure the summary explicitly describes:
   - what is happening;
   - who is likely affected.
9. Ensure each hypothesis contains 2–3 concrete next diagnostic steps.
10. Load past incidents as structured auxiliary knowledge records.
11. During context construction, select a relevant subset of past incidents rather than always passing the full set. 

## 6. Non-Goals

- No real integration with ELK, PostgreSQL, Grafana, or provider APIs.  
- No authentication or multi-user support.  
- No persistent storage.  
- No advanced UI.  
- No full autonomous incident resolution.  
- No vector database or heavy RAG infrastructure.

## 7. Constraints

- Time budget: ~3 hours.  
- Keep the solution compact and readable.  
- Configuration must come from environment variables.  
- LLM usage must be explicitly staged, not hidden behind one "analyze()" call.

## 8. Mandatory Orchestration Requirement

The solution must explicitly separate these stages in code:

1. input parsing
2. use of auxiliary data:
   - system description
   - past incidents
3. structured answer generation
4. validation and recovery

## 9. Deterministic vs LLM-driven Responsibilities

### Deterministic

- request handling  
- loading static knowledge  
- schema validation  
- retry policy  
- language check / basic normalization  
- output formatting

### LLM-driven

- incident classification  
- concise summary generation  
- severity assessment  
- hypothesis generation  
- diagnostic step generation

## 10. Acceptance Criteria

The project is successful if:

- it runs locally with minimal steps;  
- the orchestration stages are visible in code;  
- output is machine-readable and validated;  
- invalid model output is handled with a recovery strategy;  
- at least 3–5 test incidents are documented in README.
- the solution uses system description and past incidents as part of the triage pipeline;

The recovery strategy should include retry and/or regeneration after validation failure.
A repair step may be used for near-valid malformed JSON.

## 11. Assumptions

- Input language is expected to be English.  
- Static knowledge is sufficient for this take-home task.  
- LLM quality is good enough for small incident triage when constrained by prompt \+ schema \+ retries.

## 12. Trade-offs Accepted

- Simpler retrieval based on static matching is acceptable.  
- Small set of categories is acceptable.  
- Perfect classification accuracy is not required.  
- Readability is preferred over framework-heavy extensibility.
