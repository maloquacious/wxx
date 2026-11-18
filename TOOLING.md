# Tool Creation vs. Shell Execution

Prioritize building reusable, dedicated tools over running ad-hoc shell commands (`bash`, `sh`, `exec`).

## When to Create a Tool
* **Repeated Tasks:**
  If you need to run a command or multi-step workflow more than once, encapsulate it into a structured tool or script with clear inputs and outputs.
* **Complex Logic:**
  Use a proper programming language (e.g., Go, Python, TypeScript) to build a tool whenever a task requires JSON parsing, error handling, retries, or state management.
* **Data Validation:**
  Build tools with strict input schemas so arguments are validated before execution, preventing silent failures or malformed commands.

## When Shelling Out is Acceptable
* **One-Off Inspection:**
  Checking basic system state (e.g., `ls`, `pwd`, `git status`) during initial exploration.
* **Running Built Tools:**
  Executing a newly created tool, test suite, or build pipeline that does not yet have a native wrapper.

## Rules for Tool Design
1. **Atomic & Predictable:**
   Each tool should do one job well and return structured output (like JSON) rather than raw terminal text.
2. **Safe by Default:**
   Avoid destructive side effects. Use explicit parameters instead of passing raw, unescaped strings to a shell.
3. **Self-Documenting:**
   Include clear docstrings, parameter descriptions, and usage examples so you can reliably invoke the tool in future steps.
