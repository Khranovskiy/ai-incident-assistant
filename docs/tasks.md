# AI Incident Assistant — Tasks (Submission Scope)

## 1. Bootstrap CLI
- [ ] Create CLI entrypoint
- [ ] Support `--text` and `--file`
- [ ] Reject invalid flag combinations
- [ ] Print final JSON to stdout
- [ ] Add optional verbose diagnostics to stderr
- [ ] Load required config from env
- [ ] Add `.env.example`

## 2. Add knowledge and core types
- [ ] Add `data/system_description.json`
- [ ] Add `data/past_incidents.json`
- [ ] Define parser output type:
  - [ ] `services`
  - [ ] `endpoints`
  - [ ] `keywords`
  - [ ] `signals`
  - [ ] `time_hints`
- [ ] Define final output type:
  - [ ] `category`
  - [ ] `summary`
  - [ ] `affected`
  - [ ] `severity`
  - [ ] `hypotheses`
- [ ] Define hypothesis type
- [ ] Define category enum
- [ ] Define severity enum

## 3. Implement parser and retrieval
- [ ] Implement deterministic parser
- [ ] Extract services, endpoints, keywords, signals, and time hints
- [ ] Do not infer category, severity, or affected users
- [ ] Implement past-incident scoring
- [ ] Select up to top 2 relevant incidents
- [ ] Skip incident injection if relevance is too low

## 4. Build context and prompts
- [ ] Build context from:
  - [ ] raw incident text
  - [ ] parsed signals
  - [ ] system description
  - [ ] selected past incidents
- [ ] Build initial generation prompt
- [ ] Build refined regeneration prompt
- [ ] Include output contract, enums, and JSON-only instruction

## 5. Implement generation and validation
- [ ] Call the LLM with structured context
- [ ] Validate top-level JSON object
- [ ] Validate required fields
- [ ] Validate category enum
- [ ] Validate severity enum
- [ ] Validate hypotheses count
- [ ] Validate each hypothesis structure
- [ ] Validate 2–3 non-empty `next_steps`
- [ ] Validate expected language
- [ ] Validate JSON-only output

## 6. Implement recovery flow
- [ ] Repair near-valid mechanical issues when safe:
  - [ ] markdown fences
  - [ ] extra surrounding text
  - [ ] trivial enum normalization
- [ ] Re-validate after repair
- [ ] If still invalid, do one refined regeneration attempt
- [ ] Re-validate regenerated output
- [ ] Return controlled error if recovery fails

## 7. Wire orchestrator and diagnostics
- [ ] Wire the full pipeline:
  - [ ] input loading
  - [ ] parsing
  - [ ] knowledge loading
  - [ ] incident selection
  - [ ] context building
  - [ ] generation
  - [ ] validation
  - [ ] repair/regeneration
  - [ ] final output
- [ ] Keep stages explicit in code
- [ ] Add optional verbose stage logs

## 8. Add examples, checks, and README
- [ ] Add example inputs for the 4 canonical incidents
- [ ] Add one small parser check
- [ ] Add one small validator check
- [ ] Add one orchestrator smoke check
- [ ] Write README:
  - [ ] how to run
  - [ ] architecture summary
  - [ ] auxiliary data usage
  - [ ] validation and recovery behavior
  - [ ] 3–5 example incidents with expected behavior
  - [ ] short trade-offs
- [ ] Final polish and consistency check against spec/design
