# Story: Vim Mode Toggle

## Summary

Make vim-style keybindings opt-in via a `--vim` CLI flag or `vim_mode: true` config setting. Default navigation uses only standard keys (arrows, PgUp/PgDn, Home/End). This eliminates accidental triggers for non-vim users and makes `Ctrl+D` unambiguous.

## Motivation

Currently, vim keybindings (`gg`, `G`, `Ctrl+U`, `Ctrl+D` for scrolling) are always active. This creates problems:

1. **Accidental triggers:** Non-vim users typing `g` twice in quick succession jump to the top of the table unexpectedly
2. **`Ctrl+D` ambiguity:** In default mode `Ctrl+D` means half-page scroll OR delete depending on context — users can't predict which will happen
3. **Discoverability:** `Ctrl+U`/`Ctrl+D` for half-page scroll is a vim convention unknown to most users
4. **Convention alignment:** k9s requires explicit vim mode; o8n should follow the same pattern

## Acceptance Criteria

### Default Mode (no flag)

- [ ] **AC-1:** `gg` chord detection is disabled — pressing `g` twice does nothing
- [ ] **AC-2:** `G` does nothing (no jump to bottom)
- [ ] **AC-3:** `Ctrl+U` does nothing in table view
- [ ] **AC-4:** `Ctrl+D` is **only** delete/terminate (when action is configured for current resource), otherwise no-op. No half-page scroll behavior
- [ ] **AC-5:** `j`/`k` do not navigate table rows
- [ ] **AC-6:** `Home` jumps to first row in table
- [ ] **AC-7:** `End` jumps to last row in table
- [ ] **AC-8:** Detail view and help modal scroll with `Up`/`Down`/`PgUp`/`PgDn` only
- [ ] **AC-9:** `Home`/`End` work in detail view and help modal (jump to top/bottom)

### Vim Mode (`--vim` or config)

- [ ] **AC-10:** `j`/`k` navigate table rows (aliases for `Down`/`Up`)
- [ ] **AC-11:** `gg` jumps to first row (chord with 500ms timeout, existing implementation)
- [ ] **AC-12:** `G` jumps to last row
- [ ] **AC-13:** `Ctrl+U` scrolls half-page up in table view
- [ ] **AC-14:** `Ctrl+D` scrolls half-page down in table view (when no delete action exists for current resource); delete/terminate takes priority when configured
- [ ] **AC-15:** `j`/`k`/`Ctrl+U`/`Ctrl+D` work in detail view and help modal for scrolling
- [ ] **AC-16:** `Home`/`End` also work in vim mode (additive, not replaced)

### Activation

- [ ] **AC-17:** `--vim` CLI flag enables vim mode for the session
- [ ] **AC-18:** `vim_mode: true` in `o8n-cfg.yaml` enables vim mode persistently
- [ ] **AC-19:** CLI flag overrides config file setting
- [ ] **AC-20:** Vim mode state is accessible in the model as `m.vimMode bool`

### Help Screen

- [ ] **AC-21:** Help screen (`?`) reflects the active mode — shows only the keybindings that are actually active
- [ ] **AC-22:** In default mode, help screen shows `Home`/`End` for first/last row
- [ ] **AC-23:** In vim mode, help screen shows `gg`/`G`, `j`/`k`, `Ctrl+U`/`Ctrl+D` in addition to standard keys

### Header Hints

- [ ] **AC-24:** Header key hints adapt to vim mode (e.g., show `j/k nav` instead of `↑↓ nav` when vim mode is active)

## Tasks

### Task 1: Add vim mode flag and config

1. Add `--vim` to CLI flag parsing (alongside `--debug`, `--no-splash`, `--skin`)
2. Add `vim_mode: bool` field to `AppConfig` struct, loaded from `o8n-cfg.yaml`
3. Store resolved value in model: `m.vimMode = cliFlag || config.VimMode`
4. CLI flag takes priority over config

### Task 2: Add Home/End support (default mode)

1. Handle `home` key event in table view: set cursor to first row (index 0)
2. Handle `end` key event in table view: set cursor to last row
3. Handle `home`/`end` in detail view and help modal: jump to top/bottom of scroll
4. These keys work in both default and vim mode (always available)

### Task 3: Gate vim keys behind vimMode check

Wrap existing vim key handlers with `if m.vimMode`:

1. `gg` chord detection (the `g` key handler and `pendingG` timeout logic)
2. `G` jump to bottom
3. `Ctrl+U` half-page up in table view
4. `Ctrl+D` half-page down in table view (keep `Ctrl+D` delete behavior ungated — it always works)
5. Add `j`/`k` handlers (new) gated by `m.vimMode && !popup && !modal && !search`
6. Gate `j`/`k`/`Ctrl+U`/`Ctrl+D` in detail view and help modal scrolling

### Task 4: Update help screen rendering

1. Read `m.vimMode` when building help content
2. In default mode: show `Home`/`End` for first/last, omit `gg`/`G`/`Ctrl+U`/`Ctrl+D`/`j`/`k`
3. In vim mode: show `gg`/`G`, `j`/`k`, `Ctrl+U`/`Ctrl+D`, plus `Home`/`End`

### Task 5: Update header key hints

1. Modify `getKeyHints()` to check `m.vimMode`
2. Default: `↑↓ nav` hint
3. Vim: `j/k nav` hint (or `↑↓/jk nav`)

### Task 6: Update tests

1. Test that `gg`/`G`/`Ctrl+U`/`j`/`k` are no-ops when `vimMode=false`
2. Test that `gg`/`G`/`Ctrl+U`/`Ctrl+D`/`j`/`k` work when `vimMode=true`
3. Test that `Home`/`End` work in both modes
4. Test that `Ctrl+D` always triggers delete when action exists, regardless of vim mode
5. Test that `--vim` flag sets `m.vimMode=true`
6. Test that `vim_mode: true` in config sets `m.vimMode=true`
7. Test that CLI flag overrides config
8. Test help screen content differs between modes

## Technical Notes

- `home` and `end` key events in Bubble Tea are `tea.KeyHome` and `tea.KeyEnd`
- The `gg` chord uses `pendingG bool` + `gTimer` with 500ms timeout — this entire mechanism should be skipped when `vimMode=false`
- `Ctrl+D` for delete must remain **ungated** — it's a resource action, not a vim motion. Only the half-page-scroll behavior of `Ctrl+D` is gated
- In detail view, `Ctrl+D` currently scrolls down 10 lines — in default mode this should use `PgDn` instead; `Ctrl+D` in detail view is only available in vim mode
- The `q`/`Q` no-op guard should remain in both modes (only `Ctrl+C` quits)
