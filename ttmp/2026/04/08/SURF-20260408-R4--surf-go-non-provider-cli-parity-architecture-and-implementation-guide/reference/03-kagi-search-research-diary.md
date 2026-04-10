# Kagi Search Research Diary

## 2026-04-10

Objective:
- apply the browser-verb process from the new help page to a concrete `kagi-search` implementation

What I checked:
- loaded `https://kagi.com/search?q=llm+transcript+attribution` in Playwright
- inspected both the accessibility snapshot and live DOM structure
- identified the stable result containers and snippet fields

Selectors and structure confirmed:
- main result blocks: `main div._0_SRI.search-result`
- grouped sibling result blocks: `main div.__srgi`
- title link: `h3 a[href^="http"]`
- display URL line: `.__sri-url-box`, `.__sri-url`, `.__sri_url_path_box`
- snippet/body: `._0_DESC.__sri-desc`, `.__sri-desc`
- quick answer box: `main .qa-container-box .qa-content`

Observed page behavior:
- Kagi includes a Quick Answer box above normal results for this query
- some result groups contain nested sibling results; deduping by final URL is necessary
- the class names used for result bodies are visible and sufficiently specific for a first implementation
- on the inspected results page, the Quick Answer box exposed only UI labels (`Quick Answer`, `References`, `Continue in Assistant`) and not a useful answer body, so the first command version suppresses Quick Answer unless substantive text is present

Implementation decisions:
- `kagi-search` should navigate directly to `https://kagi.com/search?q=<escaped query>`
- the embedded browser script should wait for either `._0_SRI.search-result` or `.__srgi`
- the scraper should ignore the Quick Answer box as a normal result row, but expose it separately in the returned payload
- the command should be dual-mode Glazed:
  - default Markdown writer output
  - structured rows behind `--with-glaze-output`

Validation status:
- Go unit tests and mock-host integration tests pass
- live DOM selector validation was done in Playwright
- native-host-backed live command validation was blocked during this pass because the local Surf socket was down
