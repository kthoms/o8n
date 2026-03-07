# Story 3.7: Search, Filter & Auto-Refresh

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an **operator**,
I want to **filter the current resource table by search term** and **toggle auto-refresh**,
so that I can efficiently find specific resources and monitor live state.

## Acceptance Criteria

1. **Given** the operator is viewing any resource table
   **When** the operator presses **`/`** and types a search term
   **Then** the table is filtered to rows matching the term in real time.
   **And** the `FilterBar` (UX specified) indicates the active filter state via a dedicated `renderFilterBar(m model)` function.

2. **Given** an active search filter is set
   **When** the operator presses **Escape** or the clear-filter key
   **Then** the filter is cleared and the full table is restored from `m.originalRows` without a re-fetch.

3. **Given** the operator presses **`Ctrl+Shift+R`** to toggle auto-refresh
   **When** auto-refresh is enabled
   **Then** the table reloads on the configured interval (`5s`) and a visual indicator (**`⟳`**) appears in the footer right area.
   **And** the indicator flashes in sync with `m.flashActive`.
   **And** pressing `Ctrl+Shift+R` again disables auto-refresh and the indicator disappears.

4. **Given** any API call is in flight
   **When** the request takes longer than 500ms
   **Then** a loading indicator (spinner) appears in the footer center column and the UI remains fully interactive.

5. **Given** the application is connected to an environment
   **When** the footer renders
   **Then** the API status indicator (**`●`** green, **`✗`** red, or **`○`** muted) is visible in the footer right column.

## Tasks / Subtasks

- [x] **Implement FilterBar & Search Logic (AC: 1, 2)**
  - [x] Add `renderFilterBar(m model) string` to `view.go` implementing the 5 UX states (Hidden, Active input, Locked/applied, Server-side active, Clearing).
  - [x] Update `Update()` key handler for **`/`** to enter `searchMode` and capture `m.originalRows`.
  - [x] Ensure `prepareStateTransition` properly clears search/filter state on navigation.
- [x] **Implement Auto-Refresh & Indicator (AC: 3)**
  - [x] Update key binding for auto-refresh toggle to **`Ctrl+Shift+R`**.
  - [x] Implement periodic refresh command using `tea.Tick`.
  - [x] Add the **`⟳`** badge to `renderFooter` in `view.go`, ensuring it uses the `m.flashActive` state for its flash effect.
- [x] **Implement API Status & Loading Debounce (AC: 4, 5)**
  - [x] Implement **`●`/`✗`/`○`** indicators in `renderFooter` based on `m.envStatus`.
  - [x] Implement a delayed loading message or tick to ensure the spinner only appears for requests > 500ms.
  - [x] Ensure the loading spinner in the footer center column is tied to `m.isLoading`.
- [x] **Hint System Integration (AC: 1, 3)**
  - [x] Add hints for `/ Filter` and `Ctrl+Shift+R Refresh` to `tableViewHints` in `internal/app/hints.go` with high priority.
- [x] **Verify & Test (AC: all)**
  - [x] Create `internal/app/main_filter_refresh_test.go` covering search mode entry, live filtering from `originalRows`, and auto-refresh toggling.
  - [x] Verify footer indicators appear/disappear correctly based on model state.

## Dev Notes

### Architecture Compliance
- **State Transition:** Search/filter clearing MUST use `prepareStateTransition` or be integrated into its logic to prevent state leakage.
- **Async Pattern:** Auto-refresh MUST be implemented via `tea.Tick` and `tea.Cmd`.
- **Footer Hint Push Model:** Update `hints.go` to include the search and refresh keys.

### UI/UX Standards
- **Filter States:**
    1. No filter: `FilterBar` hidden.
    2. Input active: `FilterBar` appears with prompt.
    3. Locked: Filter text shown in locked style.
    4. Server-side: Indicator for API-level filtering.
    5. Clearing: Transition state.
- **Status Symbols:** Use `●` (operational), `✗` (unreachable), `○` (unknown).

### Project Structure Notes
- **View Logic:** `internal/app/view.go` (footer and FilterBar).
- **Update Logic:** `internal/app/update.go` (key handlers and tick handling).
- **Transition Logic:** `internal/app/transition.go`.

### References
- [Source: `_bmad/planning-artifacts/epics.md#Story 3.7`]
- [Source: `_bmad/planning-artifacts/ux-design-specification.md#Search and Filter Patterns`]
- [Source: `_bmad/planning-artifacts/ux-design-specification.md#Auto-refresh indicator design`]

## Dev Agent Record

### Agent Model Used

Claude Haiku 4.5 (implementation)

### Debug Log References

N/A

### Completion Notes List

- **FilterBar Implementation**: Added renderFilterBar() function to view.go that displays filter status in 5 UX states (Hidden, Active input, Locked/applied, Server-side, Clearing). Shows when popup search mode is active or search term is set. Integrated into main View() layout between header and table content.
- **Filter/Search State Management**: Leveraged existing search infrastructure (m.popup, m.originalRows, m.filteredRows, m.searchTerm). "/" key already opens popup search mode; Escape clears filter and restores originalRows from backup.
- **Auto-Refresh Key Binding**: Updated key handler to support Ctrl+Shift+R (per story 3.7) and "r" shortcut for backward compatibility. Updated hints to show "Ctrl+Shift+r" as primary binding (Priority 6, MinWidth 90).
- **Auto-Refresh & Flash Indicator**: Enhanced footer rendering to show ⟳ symbol when m.autoRefresh=true and m.flashActive=true, providing visual feedback when auto-refresh is active.
- **Loading Indicator**: Added debounced loading spinner that only appears after 500ms elapsed time (checks time.Since(m.apiCallStarted) > 500*time.Millisecond). Uses spinnerFrames animation.
- **API Status Indicators**: Implemented ●/✗/○ indicators in footer right area based on m.envStatus map:
  - ● (green) for StatusOperational
  - ✗ (red) for StatusUnreachable
  - ○ (muted) for StatusUnknown
- **Filter Hints**: Updated tableViewHints() in hints.go:
  - Changed "/" hint label from "find" to "filter" (AC 1)
  - Changed "Ctrl+r" hint to "Ctrl+Shift+r" with label "refresh" (AC 3)
  - Both hints have Priority 3-6 for high discoverability
- **Test Updates**: Updated test references to use "Ctrl+Shift+r" instead of "Ctrl+r" for hint validation:
  - layout_validation_test.go: expectedVisible and hiddenKeys arrays
  - main_hints_test.go: findHint call in TestCurrentViewHints_RefreshHintMinWidthNinety
  - quick_wins_test.go: refresh hint detection logic
- **Comprehensive Test Suite**: Created main_filter_refresh_test.go with 12 test functions covering:
  - AC 1: Search mode entry and real-time filtering from originalRows
  - AC 2: Filter clearing and fullrows restoration
  - AC 3: Auto-refresh toggle and indicator flash state
  - AC 4: Loading indicator 500ms debounce
  - AC 5: API status indicator symbols and footer placement
  - Hints integration for "/" and "Ctrl+Shift+r"
- **All Tests Passing**: make test shows 100% pass rate; all existing tests continue to pass with updated key bindings.

### File List

- `internal/app/hints.go` — UPD: Changed "/" label to "filter", updated "Ctrl+r" to "Ctrl+Shift+r" for refresh hint
- `internal/app/update.go` — UPD: Changed auto-refresh key binding to "ctrl+shift+r" (with "r" fallback), added time import for 500ms check
- `internal/app/view.go` — UPD: Added renderFilterBar() function; enhanced footer with loading spinner, auto-refresh ⟳ indicator, API status ●/✗/○; added time import
- `internal/app/layout_validation_test.go` — UPD: Updated expectedVisible and hiddenKeys to use "Ctrl+Shift+r" instead of "Ctrl+r"
- `internal/app/main_hints_test.go` — UPD: Updated TestCurrentViewHints_RefreshHintMinWidthNinety to search for "Ctrl+Shift+r"
- `internal/app/quick_wins_test.go` — UPD: Updated TestWin2_KeyHintsRespectTerminalWidth to use "Ctrl+Shift+r"
- `internal/app/main_filter_refresh_test.go` — NEW: 12 comprehensive test functions covering all acceptance criteria (AC 1-5)
