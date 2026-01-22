# TUI Migration: tview → Bubbletea

## Overview

Migrate the portfolio tracker TUI from `tview` to `bubbletea` with the Charm ecosystem.
The new implementation will live in `btui/` alongside the existing `tui/` directory.

## Goals

- Modern Bubbletea implementation following Elm architecture (Model/View/Update)
- Dark theme with Bubbletea aesthetic using Lipgloss
- New features: spinners, status bar, help menu with key mappings
- Maintain backward compatibility with tview version during migration

## Tech Stack

| Purpose | Library |
|---------|---------|
| Core TUI | [bubbletea](https://github.com/charmbracelet/bubbletea) |
| Styling | [lipgloss](https://github.com/charmbracelet/lipgloss) |
| Components | [bubbles](https://github.com/charmbracelet/bubbles) (table, spinner, help, textinput) |
| Forms | [huh](https://github.com/charmbracelet/huh) |

## Architecture

```
btui/
├── main.go           # Entry point, tea.NewProgram
├── model.go          # Root model, state management
├── theme.go          # Lipgloss styles and color scheme
├── keys.go           # Key bindings definitions
├── messages.go       # Custom messages for async operations
├── views/
│   ├── accounts.go   # Accounts list view (main screen)
│   ├── account.go    # Single account detail view
│   └── help.go       # Help overlay with key mappings
├── components/
│   ├── table.go      # Styled table wrapper
│   ├── statusbar.go  # Bottom status bar
│   ├── header.go     # Top header component
│   └── modal.go      # Modal dialogs (add/delete transaction)
└── forms/
    ├── transaction.go # Add transaction form using huh
    └── confirm.go     # Confirmation dialogs
```

## Color Scheme (Dark Theme - Bubbletea Style)

| Element | Color | Hex |
|---------|-------|-----|
| Background | Dark Gray | `#1a1a2e` |
| Foreground | Light Gray | `#eaeaea` |
| Primary/Accent | Indigo | `#7c3aed` |
| Secondary | Cyan | `#22d3ee` |
| Header BG | Deep Purple | `#312e81` |
| Header FG | White | `#ffffff` |
| Selected BG | Purple | `#6366f1` |
| Selected FG | White | `#ffffff` |
| Positive | Green | `#22c55e` |
| Negative | Red | `#ef4444` |
| Muted | Gray | `#6b7280` |
| Border | Slate | `#475569` |

---

## Tasks

### Phase 1: Foundation ✅
- [x] **1.1** Add bubbletea dependencies (`go get` bubbletea, bubbles, lipgloss, huh)
- [x] **1.2** Create `btui/` directory structure
- [x] **1.3** Implement `theme.go` with Lipgloss styles
- [x] **1.4** Implement `keys.go` with key bindings using bubbles/key
- [x] **1.5** Implement `messages.go` for custom tea.Msg types
- [x] **1.6** Implement root `model.go` with state enum and basic Init/Update/View

### Phase 2: Core Components ✅
- [x] **2.1** Implement `components/statusbar.go` - bottom bar with mode, hints, loading
- [x] **2.2** Implement `components/header.go` - top title bar
- [x] **2.3** Implement `components/table.go` - styled table wrapper using bubbles/table
- [x] **2.4** Implement `views/help.go` - overlay showing all key bindings

### Phase 3: Main Views ✅
- [x] **3.1** Implement `views/accounts.go` - accounts list with table, tag filter, currency toggle
- [x] **3.2** Implement `views/account.go` - single account view with transactions table
- [x] **3.3** Wire up navigation between accounts list and account detail

### Phase 4: Forms & Modals ✅
- [x] **4.1** Implement `forms/transaction.go` - add transaction form using huh
- [x] **4.2** Implement `forms/confirm.go` - delete/abandon confirmation dialogs
- [x] **4.3** Implement `components/modal.go` - modal container/overlay

### Phase 5: Integration & Polish ✅
- [x] **5.1** Implement `main.go` entry point with tea.NewProgram
- [x] **5.2** Add spinner for loading states (initial data load, currency switch)
- [x] **5.3** Add command-line flag to choose TUI backend (`-tui=bubble` vs `-tui=tview`)
- [x] **5.4** Test all functionality matches tview version
- [x] **5.5** Update AGENTS.md with new run commands

### Phase 6: Cleanup (Future - After Validation)
- [ ] **6.1** Make bubbletea the default TUI
- [ ] **6.2** Remove tview dependency (optional, keep for reference)

---

## Key Mappings (New)

| Key | Context | Action |
|-----|---------|--------|
| `↑/k` | Table | Move up |
| `↓/j` | Table | Move down |
| `Enter` | Accounts | Open account detail |
| `Esc` | Account detail | Back to accounts |
| `Esc` | Modal | Close modal |
| `Tab` | Any | Cycle focus |
| `n` | Account detail | New transaction |
| `d` | Account detail | Delete selected transaction |
| `1` | Accounts | View in USD |
| `2` | Accounts | View in NIS |
| `t` | Accounts | Cycle tag filter |
| `h` | Account detail | Toggle dividends |
| `?` | Any | Toggle help |
| `q` | Any | Quit |

---

## Progress Log

_Updated as tasks complete_

| Date | Task | Status |
|------|------|--------|
| 2026-01-22 | Phase 1: Foundation | ✅ Complete |
| 2026-01-22 | Phase 2: Core Components | ✅ Complete |
| 2026-01-22 | Phase 3: Main Views | ✅ Complete |
| 2026-01-22 | Phase 4: Forms & Modals | ✅ Complete |
| 2026-01-22 | Phase 5: Integration & Polish | ✅ Complete |

