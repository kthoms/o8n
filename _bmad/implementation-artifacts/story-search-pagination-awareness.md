# Story: Search Pagination Awareness

## Summary

Improve the search experience when dealing with paginated data. Make the scope limitation explicit and consider offering server-side search for results beyond the current page.

## Motivation

Currently, `/` search filters rows client-side within the current page only. On a table with 500 items and a page size of 25, the user is searching only 5% of the data. A footer warning exists but the UX could be clearer. k9s handles this well by making the search scope always visible.

## Acceptance Criteria

- [ ] **AC-1:** When search is active on a paginated view, the search popup or content title clearly indicates scope: e.g., `[/term/ — 3 of 25 on page, 500 total]`
- [ ] **AC-2:** When total items exceed page size and search is active, a hint appears suggesting the user can press `Enter` then `Ctrl+A` to search all pages (server-side)
- [ ] **AC-3:** `Ctrl+A` (search-All) when a search filter is locked triggers a server-side filtered fetch (using the locked search term as a query filter parameter where the API supports it)
- [ ] **AC-4:** If the API does not support server-side filtering for the current resource, `Ctrl+A` shows a footer message: "Server-side search not available for this resource"
- [ ] **AC-5:** The search scope indicator disappears when search is cleared

## Tasks

### Task 1: Enhance search scope indicator

1. When search is active and `pageTotals[currentRoot]` exceeds the current page size, update the content title to include total count
2. Show a persistent hint in the search popup footer: `Enter to lock, then Ctrl+A: search all pages`

### Task 2: Implement server-side search trigger

1. After a search filter is locked (Enter pressed), `Ctrl+A` triggers a server-side fetch with the search term as a query parameter
2. Use the table's existing API path with an additional filter parameter (resource-dependent)
3. Fallback: show footer message if the API doesn't support text-based filtering
4. On success, replace the current rows with server-side results and update title to indicate "all pages"

### Task 3: Update tests

1. Test search scope indicator formatting
2. Test server-side search trigger behavior
3. Test fallback behavior for unsupported resources
4. Test that search state is properly cleared

## Technical Notes

- Operaton REST API supports various filter parameters per resource (e.g., `processDefinitionKeyLike`, `nameLike`)
- Not all resources support text filtering — this needs a per-table config flag (e.g., `search_param: nameLike`)
- **Key conflict resolved:** `Ctrl+F` stays as page-forward (consistent with vim/k9s convention). Server-side search uses `Ctrl+A` (search-All) — only active after a search filter is locked via `Enter`. This avoids any context-dependent key remapping and keeps the interaction flow clear: `/term` -> `Enter` (lock) -> `Ctrl+A` (widen to all pages)
- `Ctrl+A` is currently unbound in o8n, so no conflict
