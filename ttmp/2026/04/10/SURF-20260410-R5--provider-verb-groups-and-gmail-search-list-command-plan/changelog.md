# Changelog

## 2026-04-10

- Initial workspace created.
- Added the main design document for provider grouping and Gmail command architecture.
- Added the implementation planning diary with target command tree and execution order.
- Replaced the placeholder task list with a detailed implementation backlog covering provider grouping, Gmail research, command implementation, tests, and docs.
- Expanded the task list into phased, developer-ready execution steps for grouped provider refactoring, Gmail research, Gmail implementation, documentation, and validation.
- Implemented grouped provider parents: `chatgpt`, `kagi`, and `gmail`.
- Moved provider commands under grouped paths: `chatgpt ask`, `chatgpt transcript`, `kagi search`, and `kagi assistant`.
- Added `gmail list --inbox` and `gmail search <query>` as dual-mode Glazed commands with embedded JS.
- Added ordered Gmail research scripts under this ticket's `scripts/` directory and validated selectors against the live Gmail session.
- Added a new Glazed help tutorial for stateful Gmail browser verbs and updated the existing browser-verb tutorial to grouped command names.
