# o8n — UX Concept

> **Audience:** Designers, developers, and contributors.  
> **Scope:** All interaction patterns, visual conventions, and layout rules in o8n as of v0.1.0.

---

## 1. Design Philosophy

o8n is modelled after **k9s** — a keyboard-first, terminal-native management UI. The core principles:

| Principle | Expression |
|---|---|
| **Keyboard first** | Every action has a key binding. Mouse is irrelevant. |
| **Information hierarchy** | Screen space is scarce; only what matters is shown. |
| **Consistency** | Same patterns across views reduce cognitive load. |
| **Progressive disclosure** | Show the right data at the right drill depth. |
| **Immediate feedback** | Every action confirms itself within the same frame. |
| **Safe destructive actions** | Two-step confirmation before irreversible operations. |

---

## 2. Screen Layout

```
┌──────────────────────────────────────────────────────────────────┐  ← Row 1–3 (always)
│  o8n v0.1.0 │ local ●              ? help  : switch  ↑↓ nav ...  │
│  (compact header: status line + key hints)                        │
└──────────────────────────────────────────────────────────────────┘
                                                                        ← Context popup (`:`) — shown/hidden
┌──────────────────────────────────────────────────────────────────┐
│  proc…ess-definitions▒▒▒                                         │  input + ghost completion
│  ↑↓:select  Tab/Enter:switch  Esc:cancel                         │
│  ▸ process-definitions                                           │  match list (up to 8 rows)
│    process-instances                                             │
└──────────────────────────────────────────────────────────────────┘
                                                                        ← Search bar (/) — shown/hidden
┌──────────────────────────────────────────────────────────────────┐
│ /search-term                                                     │
└──────────────────────────────────────────────────────────────────┘
┌──────process-definitions — 42 items──────────────────────────────┐  ← Content box (fills remaining height)
│  KEY           NAME                     VERSION   RESOURCE        │
│  ▶ my-proc     My Process               3         my-proc.bpmn   │  ← selected row (inverted)
│    other-proc  Other Process            1         other.bpmn     │
│    ...                                                            │
└──────────────────────────────────────────────────────────────────┘
[1] <process-definitions> | ✗ error message here     | [2/5] ⚡ 34ms   ← Footer (1 row)
```

### 2.1 Vertical Regions (in order)

| Region | Height | Always visible? |
|---|---|---|
| Compact header | 3 rows | ✓ |
| Context popup | 0 or 4–13 rows | Only when `:` pressed |
| Search bar | 0 or 1 row | Only when `/` active |
| Content box | Fills remainder | ✓ |
| Footer | 1 row | ✓ |

### 2.2 Content Box Title

The content box title (embedded in the top border) adapts to state:

```
process-definitions — 42 items             ← normal
process-definitions [/proc/ — 3 of 42]    ← search active
process-instances(my-key)                  ← drilled into instances
```

### 2.3 Footer Columns

```
[1] <process-definitions> | ✗ error message — Ctrl+r to retry | [2/5] ⚡ 34ms
│                            │                                    │
└─ Breadcrumb (colored)     └─ Status message                   └─ Pagination + Remote indicator
```

- **Column 1 (breadcrumb):** Each crumb is colored by depth level (env accent → lavender → mint → peach). Numbered `[1]`, `[2]`… for direct navigation.
- **Column 2 (status):** Error `✗` (red), success `✓` (green), info `ℹ` (blue), loading spinner (yellow). Auto-clears after 5 s. Errors append `— Ctrl+r to retry`.
- **Column 3 (remote):** `⚡` flashes for 200 ms after each API call. Shows API latency (`34ms`) and page position `[2/5]` in neutral grey.

---

## 3. Splash Screen

On startup: animated ASCII logo reveal over 15 frames (100 ms each ≈ 1.5 s total). Logo lines are revealed top-to-bottom. Version number fades in at frame 8. Centered over full terminal. Bypassed with `--no-splash`.

```
         ____
  ____  ( __ )____
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /
\____/\____/_/ /_/

         v0.1.0
```

---

## 4. Navigation Patterns

### 4.1 Drill-Down Hierarchy

```
Context Switch (:)
      │
      ▼
process-definitions  ──Enter──▶  process-instances  ──Enter──▶  variables
       ◀──────────── Esc ──────────────────────────── Esc ──────────────────
```

- **Enter** drills down to the next level.
- **Esc** pops the navigation stack (one level back).
- **1–4** jump directly to a breadcrumb level.
- Each drill-down pushes a `viewState` snapshot (rows, cursor, columns, params) for lossless back navigation.

### 4.2 Vim Motion Layer

| Key | Action |
|---|---|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g g` | Jump to top |
| `G` | Jump to bottom |
| `Ctrl+u` | Half-page scroll up |
| `Ctrl+d` | Half-page scroll down (or terminate in instances view) |

### 4.3 Pagination

- Server-side pagination via `firstResult` / `maxResults`.
- `PgDn` / `Ctrl+f` advances; `PgUp` / `Ctrl+b` retreats.
- At first/last page: no-op fetch is skipped; a brief hint appears.
- Page position shown neutrally in footer: `[2/5]` (grey `#888888`, not flash colour).

### 4.4 Context Switcher (`:`)

```
┌───────────────────────────────────────────────────┐
│ proc▌ess-definitions                              │  typed + ghost suffix
│ ↑↓:select  Tab/Enter:switch  Esc:cancel           │
│ ▸ process-definitions                             │  cursor highlight
│   process-instances                               │
│   … 4 more                                        │
└───────────────────────────────────────────────────┘
```

- Ghost completion in grey `#666666` after first character typed.
- `Tab` completes to first match.
- `↑`/`↓` moves popup cursor.
- `Enter` switches context (exact match required).
- `Esc` cancels; clears input.
- On switch: resets viewMode, clears navigation stack, fetches new root.
- Content pane shrinks exactly to accommodate popup height (no layout break).

---

## 5. Search & Filter

- `/` activates search mode; a search bar appears below the header.
- Typing filters table rows client-side; matches shown immediately.
- Title updates: `[/term/ — 3 of 47]` showing match count vs. total.
- `Enter` locks filter; `Esc` clears filter and exits search mode.

---

## 6. Modal Dialogs

All modals are **overlays** on top of the main UI (never blank-canvas replacements). Two rendering approaches:

| Modal | Rendering |
|---|---|
| Confirm Delete | `overlayCenter(baseView, modal)` |
| Confirm Quit | `overlayCenter(baseView, modal)` |
| Edit | `overlayCenter(baseView, modal)` |
| Help | Full-screen `lipgloss.Place` (scrollable) |
| Sort | Full-screen `lipgloss.Place` (centered) |
| Detail View | Full-screen `lipgloss.Place` (centered, scrollable) |
| Environment Picker | Full-screen `lipgloss.Place` (centered) |

### 6.1 Confirm Delete (Two-Step Pattern)

```
╔══════════════════════════════════════════════════════╗
║  ⚠️  DELETE PROCESS-INSTANCE                         ║
║                                                      ║
║  You are about to DELETE this item:                  ║
║                                                      ║
║  ID:            abc-123-def                          ║
║  Name:          My Process                           ║
║                                                      ║
║  ⚠️  WARNING: This action CANNOT be undone!          ║
║                                                      ║
║  Ctrl+d  Confirm Delete    Esc  Cancel               ║
╚══════════════════════════════════════════════════════╝
```

- First `Ctrl+d` opens modal.
- Second `Ctrl+d` executes.
- Any other key cancels with "Cancelled" feedback.

### 6.2 Confirm Quit

```
╭──────────────────────────────────────╮
│  Quit o8n?                           │
│                                      │
│  Press Ctrl+c again to confirm.      │
│                                      │
│  Ctrl+c  Quit    Esc  Cancel         │
╰──────────────────────────────────────╯
```

Same two-step pattern. Rendered as overlay over live UI.

### 6.3 Edit Modal

```
╭──────────────────────────────────────────────────────────╮
│  Editing: value  (type: json)                            │
│  ─────────────────────────────────                       │
│                                                          │
│  > {"key": "val"}█                                       │
│                                                          │
│  ⚠ Must be valid JSON                                    │
│                                                          │
│  [ Save ]   [ Cancel ]                                   │
╰──────────────────────────────────────────────────────────╯
```

- Tab cycles focus: Input → Save → Cancel → Input.
- Save button is **disabled** (grey) when validation fails.
- Bool fields use `Space` to toggle `true`/`false`.
- Type-aware validation: `bool`, `int`, `float64`, `json`, `text`, `user`.
- "user" type shows content-assist suggestions from cache.

### 6.4 Help Screen (Scrollable)

Full-screen three-column layout. Scrollable with `j`/`k`/`↑`/`↓`/`Ctrl+d`/`Ctrl+u`. Scroll indicators `↑ more above` / `↓ more below`. Any other key closes.

### 6.5 Detail View

Scrollable JSON viewer with line numbers and syntax highlighting:
- Keys: `#00BCD4` (cyan)
- Strings: `#50C878` (green)
- Numbers: `#FFD700` (gold)
- Booleans/null: `#FF69B4` (pink)

### 6.6 Sort Popup

Column list with `▸` cursor and `▲`/`▼` indicator on active sort column. Includes "— clear sort —" option when sort is active.

### 6.7 Environment Picker

Lists all configured environments with:
- Connection status: `●` green (operational), `○` yellow (unknown), `✗` red (unreachable)
- Environment accent colour applied to name
- `✓` marker on currently active environment
- URL shown inline

---

## 7. Key Bindings Reference

### 7.1 Global

| Key | Action |
|---|---|
| `?` | Open help |
| `:` | Open context switcher |
| `Ctrl+c` | Quit (with confirmation overlay) |
| `Ctrl+e` | Open environment picker |
| `Ctrl+r` | Toggle auto-refresh (5 s interval) |
| `/` | Enter search mode |

### 7.2 Navigation

| Key | Action |
|---|---|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Drill down |
| `Esc` | Go back |
| `PgDn` / `Ctrl+f` | Next page |
| `PgUp` / `Ctrl+b` | Prev page |
| `g g` | Jump to top |
| `G` | Jump to bottom |
| `Ctrl+u` | Half-page up |
| `1`–`4` | Jump to breadcrumb level N |

### 7.3 Context-Specific

| Context | Key | Action |
|---|---|---|
| Any | `s` | Sort popup |
| Any | `Space` | Actions menu |
| Any | `y` | Detail view (JSON) |
| Instances | `Ctrl+d` | Terminate (2-step confirm) |
| Variables | `e` | Edit value |

### 7.4 Hint Priority System

Hints are shown/hidden based on terminal width:

| Priority | Width threshold | Example |
|---|---|---|
| 1 | Always | `? help` |
| 2 | Always | `: switch` |
| 3 | Always | `↑↓ nav`, `PgDn/PgUp page` |
| 4 | 85+ | `Enter drill`, `e edit var` |
| 5 | 88+ | `s sort`, `Esc back` |
| 6 | 100+ | `Space actions`, `Ctrl+r refresh` |
| 7 | 100+ | `Ctrl+d terminate` |
| 8 | 110+ | `Ctrl+c quit` |
| 9 | 105+ | `Ctrl+e env` |

---

## 8. Visual Design System

### 8.1 Colour Roles

| Role | Source | Usage |
|---|---|---|
| Accent / border | `env.UIColor` (per-env) | Borders, focus highlight, breadcrumb L1 |
| Body fg | Skin `body.fgColor` | Table content |
| Body bg | Skin `body.bgColor` | Terminal background / selected row contrast |
| Focus bg | Skin `border.focusColor` | Selected row background |
| Ghost text | `#666666` | Completion suffix, inactive elements |
| Page counter | `#888888` | Pagination `[N/M]` (neutral, not accent) |

### 8.2 Status Colours

| Status | Colour | Symbol | Usage |
|---|---|---|---|
| Operational / Running | `#00FF00` green | `●` | Engine reachable, instance running |
| Suspended | `#FFAA00` yellow | `●` | Instance suspended |
| Unknown | `#FFAA00` yellow | `○` | Health not checked yet |
| Unreachable / Failed | `#FF0000` red | `✗` | Engine down, instance failed |
| Ended | dim | `○` | Instance completed |

### 8.3 Footer Status Icons

| Kind | Icon | Colour |
|---|---|---|
| Error | `✗` | `#FF6B6B` red + bold |
| Success | `✓` | `#50C878` green + bold |
| Info | `ℹ` | `#00A8E1` blue |
| Loading | `⠋⠙⠹…` (Braille spinner) | `#FFD700` yellow |

### 8.4 Skin System

36 built-in skins (Dracula, Nord, Gruvbox, etc.) in `skins/*.yaml`. Active skin set in `o8n-env.yaml`. Skin controls: body fg/bg, logo colour, border fg, border focus colour. Falls back to `stock.yaml` if skin not found. Environment `UIColor` always overrides border accent.

---

## 9. Tables

### 9.1 Column System

- Widths specified as percentages in `o8n-cfg.yaml`; remaining distributed equally.
- Columns auto-hide below width thresholds (hide-able flag + terminal width).
- Low-priority columns shrink to header width + 1 space before hiding.
- All rows normalised to configured column count (pad empty / truncate excess).

### 9.2 Selection

- Selected row: inverted colours (focus bg from skin / accent dark shade).
- Drillable rows: prefixed with `▶ ` focus indicator (stripped when resolving IDs).
- Editable columns: cell value shown with `[E]` suffix.

### 9.3 Sorting

- Client-side sort on any visible column.
- Sort direction toggles on re-select: `▲` ascending → `▼` descending.
- Sort state resets on context switch or data refresh.

---

## 10. Feedback & Responsiveness

### 10.1 API Call Feedback

1. `⚡` appears in footer right for **200 ms** on every REST call.
2. API latency shown alongside: `⚡ 34ms`.
3. Braille spinner `⠋⠙…` in footer centre during loading operations.

### 10.2 Error Handling

- All API errors go to footer status column in red.
- Message is truncated to available column width.
- Auto-clears after **5 seconds** via scheduled `clearErrorMsg`.
- Error messages include `— Ctrl+r to retry` retry hint.
- Friendly translations for common network errors (connection refused, timeout, cert error, unknown host).

### 10.3 Action Feedback

| Action | Feedback |
|---|---|
| Context switch | Footer clears, new rows load |
| Terminate success | `"Terminated: <id>"` in footer (green) |
| Edit save success | Modal closes; `"Saved"` briefly in footer |
| Refresh toggle on | `↺` badge in header (accent colour) |
| Cancel confirmation | `"Cancelled"` for 2 s |
| Pagination boundary | `"First page"` / `"Last page"` for 2 s |

---

## 11. Environment & Health

- All configured environments polled on startup and every 60 s.
- Status stored per-env: `operational` / `unreachable` / `unknown`.
- Header shows current env name + colour-coded `●`/`✗`/`○` symbol.
- Environment picker shows all envs with status, colour, and URL.
- Switching env: `Ctrl+e` cycles, or `Ctrl+e` opens picker modal.
- Active env persisted to `o8n-env.yaml` on switch.

---

## 12. Responsive Behaviour

| Terminal width | Behaviour |
|---|---|
| < 80 | Minimum viable; most hints hidden |
| 80–99 | Core nav hints visible |
| 100–109 | Actions menu hint + terminate hint appear |
| 110–119 | Quit hint + detail view hint appear |
| 120–139 | Env hint visible |
| 140+ | All hints visible |

Content box always fills remaining height after header + popup + footer.

---

## 13. Actions Menu

`Space` opens a context-sensitive actions menu (config-driven from `o8n-cfg.yaml`). "View as JSON" is always appended as the last item. Actions can require confirmation (triggers `ModalConfirmDelete`).

```
╭────────────────────────────────────╮
│  Actions: process-instance         │
│  my-process-id                     │
│                                    │
│  ▸ [s] Suspend                     │
│    [r] Resume                      │
│    [y] View as JSON                │
│                                    │
│  Esc: Close                        │
╰────────────────────────────────────╯
```

---

## 14. Open UX Gaps

The following areas have been identified as improvement opportunities (not yet implemented):

| ID | Area | Description |
|---|---|---|
| G1 | Error recovery | After error clears, no affordance for retry without manual Ctrl+r |
| G2 | Multi-select | No bulk action support; all operations are single-row |
| G3 | Column resize | No horizontal scroll or runtime column width adjustment |
| G4 | Sort persistence | Sort resets on every fetch; could be persisted per context |
| G5 | Keyboard map conflicts | `Ctrl+d` is both half-page scroll (vim) and terminate — context-dependent disambiguation works but is not discoverable |
| G6 | Empty state messaging | Empty tables show blank content rather than a contextual hint |
| G7 | Loading indication | Spinner only in footer; no indication inside the content box during load |

---

*Last updated: 2026-02-23. Source of truth: `view.go`, `model.go`, `nav.go`, `styles.go`, `util.go`.*
