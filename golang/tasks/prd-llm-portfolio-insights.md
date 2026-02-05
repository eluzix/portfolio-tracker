# PRD: LLM Portfolio Insights

## Introduction

Add AI-powered portfolio analysis using OpenAI's API to provide insights, recommendations, and comparisons against market benchmarks. Users can trigger analysis with `Ctrl+S` from either the accounts list (full portfolio) or account detail view (single account). Results display in a scrollable floating window with copy-to-clipboard functionality.

## Goals

- Provide AI-generated insights on portfolio performance and composition
- Compare holdings against current market benchmarks (S&P 500, sector performance)
- Identify potential risks and opportunities based on transaction history
- Zero new dependencies—implement OpenAI API client with standard library only
- Configurable model via `OPENAI_MODEL` env var (default: `gpt-5.2`)

## User Stories

### US-001: OpenAI API Client Implementation
**Description:** As a developer, I need a zero-dependency OpenAI client so the app can request completions without adding external libraries.

**Acceptance Criteria:**
- [ ] Create `llm/openai.go` with HTTP client using `net/http`
- [ ] Implement chat completions endpoint (`POST /v1/chat/completions`)
- [ ] Read API key from `OPENAI_KEY` environment variable
- [ ] Read model from `OPENAI_MODEL` env var, default to `gpt-5.2`
- [ ] Handle streaming disabled (simple JSON response)
- [ ] Return structured error on API failure (rate limit, auth, network)
- [ ] Typecheck passes (`go build ./...`)

### US-002: Portfolio Context Builder
**Description:** As a developer, I need to build a structured prompt with portfolio data so the LLM has full context for analysis.

**Acceptance Criteria:**
- [ ] Create `llm/prompt.go` with context builder functions
- [ ] For single account: include account name, holdings, all transactions, performance metrics (gain, yield, dividends)
- [ ] For full portfolio: aggregate all accounts with same data
- [ ] Format data as structured text (not JSON) for better LLM comprehension
- [ ] Include current date for temporal context
- [ ] Keep prompt under 8000 tokens estimate (truncate oldest transactions if needed)
- [ ] Typecheck passes

### US-003: Add Summarize Keybinding
**Description:** As a user, I want to press `Ctrl+S` to trigger AI analysis of my portfolio or current account.

**Acceptance Criteria:**
- [ ] Add `Summarize` key binding to `KeyMap` struct: `ctrl+s`
- [ ] Add to help text: "Ctrl+S" → "AI insights"
- [ ] Works in `ViewAccounts` (analyzes full portfolio)
- [ ] Works in `ViewAccountDetail` (analyzes selected account only)
- [ ] Shows loading state in status bar while API call runs
- [ ] Typecheck passes

### US-004: Scrollable Insights Modal
**Description:** As a user, I want to view AI insights in a scrollable floating window so I can read long responses.

**Acceptance Criteria:**
- [ ] Create `tui/views/insights.go` with `InsightsView` component
- [ ] Floating modal centered on screen (80% width, 70% height)
- [ ] Scrollable content area using viewport (bubbles/viewport or manual implementation)
- [ ] Styled header showing "Portfolio Insights" or "Account: {name} Insights"
- [ ] Footer showing scroll position and available actions
- [ ] `↑/↓/j/k` for scrolling, `Esc` to close
- [ ] Typecheck passes

### US-005: Copy to Clipboard Button
**Description:** As a user, I want to copy the AI insights to clipboard so I can save or share them.

**Acceptance Criteria:**
- [ ] Add `c` keybinding in insights modal: "copy to clipboard"
- [ ] Copy full markdown content to system clipboard
- [ ] Use `golang.design/x/clipboard` OR shell fallback (`pbcopy`/`xclip`)
- [ ] Show confirmation in status bar: "Copied to clipboard"
- [ ] Handle clipboard errors gracefully (show error, don't crash)
- [ ] Typecheck passes

### US-006: Integrate Insights Modal into TUI
**Description:** As a developer, I need to wire the insights modal into the main TUI model and handle the async API flow.

**Acceptance Criteria:**
- [ ] Add `ModalInsights` to `ModalType` enum in `messages.go`
- [ ] Add `insightsView` field to `Model` struct
- [ ] Add `InsightsLoadedMsg` and `InsightsErrorMsg` message types
- [ ] On `Ctrl+S`: show loading spinner, dispatch async API call
- [ ] On success: populate and show insights modal
- [ ] On error: show error message in the modal (not toast)
- [ ] Typecheck passes

### US-007: System Prompt Engineering
**Description:** As a developer, I need a well-crafted system prompt so the LLM provides actionable financial insights.

**Acceptance Criteria:**
- [ ] Create system prompt in `llm/prompt.go`
- [ ] Instruct LLM to act as financial analyst
- [ ] Request analysis of: diversification, sector allocation, risk exposure, performance vs benchmarks
- [ ] Request actionable insights and potential concerns
- [ ] Format output as markdown with headers
- [ ] Include disclaimer about not being financial advice
- [ ] Typecheck passes

## Functional Requirements

- FR-1: Read `OPENAI_KEY` from environment; if missing, show error in modal when triggered
- FR-2: Read `OPENAI_MODEL` from environment; default to `gpt-5.2`
- FR-3: `Ctrl+S` in accounts view triggers full portfolio analysis
- FR-4: `Ctrl+S` in account detail view triggers single account analysis
- FR-5: Show spinner/loading indicator in status bar during API call
- FR-6: Display results in centered floating modal (80% × 70% of screen)
- FR-7: Modal content is scrollable with `↑/↓/j/k` keys
- FR-8: `c` key copies full content to system clipboard
- FR-9: `Esc` closes the modal and returns to previous view
- FR-10: API errors display inside the modal, not as separate dialogs
- FR-11: Add `Ctrl+S` to help overlay under "General" section

## Non-Goals

- No streaming responses (use simple request/response)
- No caching of insights (always fresh analysis)
- No model selection UI (env var only)
- No token usage display or cost tracking
- No support for other LLM providers (OpenAI only)
- No persistent storage of generated insights

## Technical Considerations

- **HTTP Client**: Use `net/http` with 60-second timeout for API calls
- **JSON Handling**: Use `encoding/json` for request/response marshaling
- **Clipboard**: Use `pbcopy` on macOS, `xclip` on Linux via `os/exec`
- **Viewport**: Implement simple scroll logic or use `charmbracelet/bubbles/viewport`
- **Existing Patterns**: Follow `Modal` component pattern from `tui/components/modal.go`
- **Async Pattern**: Follow existing `tea.Cmd` pattern used in `loadData()` and `loadExchangeRate()`

## File Structure

```
llm/
  openai.go      # HTTP client for OpenAI API
  prompt.go      # Context builder and system prompt
  types.go       # Request/response structs
tui/
  keys.go        # Add Summarize binding
  messages.go    # Add InsightsLoadedMsg, InsightsErrorMsg
  model.go       # Wire up modal and async flow
  views/
    insights.go  # Scrollable insights modal
    help.go      # Update with new keybinding
```

## Success Metrics

- Analysis completes in under 30 seconds for typical portfolios
- Modal renders correctly at various terminal sizes
- Copy to clipboard works on macOS and Linux
- No panics or crashes on API errors or missing env vars

## Open Questions

- Should we add a loading animation inside the modal while waiting?
- Should `Ctrl+S` be disabled if `OPENAI_KEY` is not set, or show error on trigger?
