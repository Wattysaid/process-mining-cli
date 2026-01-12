# User Profiles and Prompt Adaptation

**Document version:** R1.01 (2026-01-12)

## 1. Purpose
PM Assist adapts its prompts and defaults based on the user's aptitude level and preferences. The tool remains decision-neutral and only assists the user in making choices.

## 2. Profile storage
- Location: `.profiles/` in the project directory
- Format: YAML, one file per user (e.g., `.profiles/jane-doe.yaml`)
- Editable by the user; the assistant may update profiles when instructed
- Filenames are sanitized (spaces become `-`, non-alphanumeric removed); the CLI should accept the original name when selecting profiles

## 3. Required fields
- `name`
- `role`
- `aptitude` (beginner|intermediate|expert)

## 4. Optional fields
- `preferences.prompt_depth` (short|standard|detailed)
- `preferences.default_output_formats` (notebook|html|both)
- `preferences.llm_provider` (openai|anthropic|gemini|ollama|none)
- `preferences.decisions.require_confirmation` (always true)

## 5. Behavioural rules
- The CLI must never auto-select algorithms, thresholds, or destructive actions.
- Use the profile to tailor explanations and the amount of guidance.
- Always surface the CLI flags that map to interactive choices.

## 6. Current implementation notes
- Active profile name is stored in `pm-assist.yaml`.
- Profile creation is integrated into `pm-assist init` and `pm-assist profile init`.

## 6. Example profile
```yaml
name: Jane Doe
role: Process Mining Analyst
aptitude: intermediate
preferences:
  prompt_depth: standard
  default_output_formats: both
  llm_provider: openai
  decisions:
    require_confirmation: true
```
