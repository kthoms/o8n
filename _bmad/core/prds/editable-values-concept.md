# Editable Values Concept

## Overview
Enable in-table editing for specific columns via config-driven metadata, modal input dialogs, and visual indicators. Users can edit process variables, task assignments, and other mutable fields directly from the TUI.

---

## Configuration Schema

### Extended ColumnDef
```yaml
- name: value
  visible: true
  width: 50%
  align: left
  editable: true              # NEW: marks column as editable
  type: string                # NEW: data type (string, int, bool, json, date)
  validationPattern: ".*"     # NEW: optional regex for validation
  maxLength: 1000             # NEW: optional max length
```

### Extended TableDef
```yaml
- name: process-variables
  columns:
    - name: name
      visible: true
      editable: false
    - name: value
      visible: true
      editable: true
      type: string
      maxLength: 5000
    - name: type
      visible: true
      editable: false
  # NEW: define which API endpoint/method to call on edit
  editEndpoint: "/process-instance/{processInstanceId}/variables/{name}"
  editMethod: PUT
  # NEW: parameter mapping (row column -> API param)
  editParams:
    - rowColumn: name         # which row column contains the variable name
      apiParam: name          # which API parameter it maps to
```

---

## Keyboard Interaction Flow

### 1. Visual Indicator (No key press needed)
- **Editable columns**: render cells with a subtle marker (e.g., `[E]` suffix or different color)
- **Current row's editable columns**: highlight with a distinct background when row is selected

### 2. Enter Edit Mode
**Key**: `e` (edit) or `F2`

**Behavior when pressed**:
- If current cell is editable → open modal for that cell
- If current cell is not editable but row has editable columns:
  - Show quick-pick menu: "Edit which field: [1] value [2] assignee [Esc] cancel"
  - User types `1` or `2` → opens modal for that column
- If row has no editable columns → show footer error: "No editable fields in this row"

### 3. Modal Input Dialog
**Layout** (centered overlay):
```
┌─ Edit Variable Value ────────────────────────┐
│                                               │
│  Variable: myVariable                         │
│  Type: String                                 │
│                                               │
│  Current Value:                               │
│  ┌───────────────────────────────────────┐   │
│  │ old-value-here                        │   │
│  └───────────────────────────────────────┘   │
│                                               │
│  New Value:                                   │
│  ┌───────────────────────────────────────┐   │
│  │ █                                     │   │ ← cursor here
│  └───────────────────────────────────────┘   │
│                                               │
│  <ctrl>+s  Save    Esc  Cancel                │
└───────────────────────────────────────────────┘
```

**Key bindings in modal**:
- `Ctrl+S` or `Enter` → validate & save
- `Esc` → cancel (no changes)
- `Ctrl+C` → cancel
- Standard text editing: arrow keys, backspace, delete, home/end, Ctrl+A (select all)

### 4. Save Flow
1. User presses `Ctrl+S`
2. Validate input (type check, regex, length)
3. If invalid → show inline error in modal: `⚠ Invalid: must be an integer`
4. If valid:
   - Close modal
   - Show footer spinner: "Saving..."
   - Call API (PUT/PATCH endpoint)
   - On success: refresh table, show flash + footer: "✓ Saved"
   - On error: show footer error: "⚠ Failed to save: {reason}"

---

## Visual Indicators

### In Table View (Normal)
```
┌─ process-variables ────────────────────────────┐
│ NAME          │ VALUE           │ TYPE        │
├───────────────┼─────────────────┼─────────────┤
│ myVar         │ 123 [E]         │ Integer     │ ← [E] = editable
│ readonly      │ abc             │ String      │
│ flag          │ true [E]        │ Boolean     │
```

### Row Selected (Editable Columns Highlighted)
```
┌─ process-variables ────────────────────────────┐
│ NAME          │ VALUE           │ TYPE        │
├───────────────┼─────────────────┼─────────────┤
│ myVar         │ 123 [E]         │ Integer     │
│►readonly      │ abc             │ String      │ ← selected, no highlight
│ flag          │ true [E]        │ Boolean     │
```

```
┌─ process-variables ────────────────────────────┐
│ NAME          │ VALUE           │ TYPE        │
├───────────────┼─────────────────┼─────────────┤
│►myVar         │╔═══════════════╗│ Integer     │ ← value is editable, highlighted
│               │║ 123 [E]       ║│             │
│               │╚═══════════════╝│             │
│ readonly      │ abc             │ String      │
```

Alternative: use color coding instead of borders (editable cell bg = lighter shade).

---

## Type-Based Input Handling

### String (default)
- Multiline textarea for long strings
- Single-line input for short strings (maxLength < 100)
- Validation: regex pattern match, maxLength

### Integer / Number
- Single-line input, numeric keyboard filter
- Validation: must parse as int/float, optional min/max

### Boolean
- Quick toggle: show `[x] true  [ ] false` radio buttons
- Or simple dropdown: `true | false`

### JSON
- Multiline textarea with syntax highlighting (if lipgloss supports it)
- Validation: must parse as valid JSON
- Show formatted JSON in modal

### Date / DateTime
- Single-line input with format hint: `YYYY-MM-DD` or `YYYY-MM-DDTHH:MM:SSZ`
- Validation: must parse as ISO8601 date

### Enum (if configured with allowed values)
```yaml
- name: status
  editable: true
  type: enum
  allowedValues: [pending, approved, rejected]
```
- Show dropdown/radio list with allowed values

---

## Multi-Column Edit Selection

**Scenario**: Row has 3 editable columns (value, assignee, priority)

**Option A: Quick-Pick Menu**
Press `e` → show inline menu in footer:
```
Edit which field?  [1] value  [2] assignee  [3] priority  [Esc] cancel
```
User presses `1`, `2`, or `3` → opens modal for that column.

**Option B: Column Hotkeys**
- `e` + `1` → edit first editable column
- `e` + `2` → edit second editable column
- `e` + `3` → edit third editable column
- `e` + `e` → edit first editable column (double-tap shortcut)

**Option C: Cycle Through**
- First `e` press → highlight first editable column
- Second `e` press → highlight second editable column
- `Enter` → open modal for highlighted column
- `Esc` → cancel selection

**Recommendation**: **Option A** (Quick-Pick) for clear UX and minimal keypresses.

---

## API Integration

### Config Mapping
```yaml
editEndpoint: "/process-instance/{processInstanceId}/variables/{name}"
editMethod: PUT
editParams:
  - rowColumn: name
    apiParam: name
  - contextParam: selectedInstanceID  # from model state
    apiParam: processInstanceId
```

### Code Flow
1. User saves edit → `handleEditSave(rowData, columnName, newValue)`
2. Resolve API endpoint template:
   - Replace `{name}` with `rowData["name"]`
   - Replace `{processInstanceId}` with `m.selectedInstanceID`
3. Build request body (type-dependent):
   - String/Number/Boolean: `{"value": newValue, "type": "String"}`
   - JSON: `{"value": parsedJSON, "type": "Object"}`
4. Call `client.UpdateVariable(instanceID, varName, newValue, varType)`
5. Handle response → update table row or show error

### Client Method Example
```go
// internal/client/client.go
func (c *Client) UpdateVariable(instanceID, varName, value, valueType string) error {
    req := c.operatonAPI.ProcessInstanceAPI.
        SetProcessInstanceVariable(c.authContext, instanceID, varName)
    
    body := operaton.VariableValueDto{
        Value: value,
        Type:  valueType,
    }
    
    _, err := req.Body(body).Execute()
    return err
}
```

---

## Implementation Phases

### Phase 1: Foundation (MVP)
- [ ] Extend `ColumnDef` with `editable`, `type`, `maxLength`
- [ ] Extend `TableDef` with `editEndpoint`, `editMethod`, `editParams`
- [ ] Add `editableColumnIndices()` helper to identify editable columns in a row
- [ ] Visual indicator: append `[E]` to editable cells
- [ ] Key handler: `e` key → detect editable columns

### Phase 2: Modal Input (String Only)
- [ ] Create `EditModal` component (centered overlay with text input)
- [ ] Handle `Ctrl+S` → validate → close
- [ ] Handle `Esc` → cancel
- [ ] Show validation errors inline
- [ ] Integrate modal into `Update()` flow

### Phase 3: API Integration
- [ ] Parse `editEndpoint` template and replace placeholders
- [ ] Add `client.UpdateVariable()` method
- [ ] Call API on save → show spinner → handle response
- [ ] Refresh table row on success

### Phase 4: Multi-Column Selection
- [ ] Implement Quick-Pick menu (Option A)
- [ ] Handle `1`, `2`, `3` key presses in selection mode
- [ ] Update footer to show menu

### Phase 5: Type-Based Inputs
- [ ] Integer: numeric filter + validation
- [ ] Boolean: radio buttons or toggle
- [ ] JSON: multiline textarea + syntax validation
- [ ] Date: format validation

### Phase 6: Polish
- [ ] Highlighted background for editable cells when row selected
- [ ] Color-coded editable columns (config-driven theme color)
- [ ] Better error messages (show API error details)
- [ ] Optimistic updates (update table immediately, revert on error)

---

## Configuration Example (Complete)

```yaml
tables:
  - name: process-variables
    columns:
      - name: name
        visible: true
        width: 30%
        editable: false
      - name: value
        visible: true
        width: 50%
        editable: true
        type: string
        maxLength: 5000
      - name: type
        visible: true
        width: 20%
        editable: false
    editEndpoint: "/process-instance/{processInstanceId}/variables/{name}"
    editMethod: PUT
    editParams:
      - rowColumn: name
        apiParam: name
      - contextParam: selectedInstanceID
        apiParam: processInstanceId

  - name: task
    columns:
      - name: id
        visible: true
        editable: false
      - name: name
        visible: true
        editable: false
      - name: assignee
        visible: true
        width: 30%
        editable: true
        type: string
        validationPattern: "^[a-zA-Z0-9_-]+$"
      - name: priority
        visible: true
        width: 15%
        editable: true
        type: integer
        minValue: 0
        maxValue: 100
    editEndpoint: "/task/{id}"
    editMethod: PATCH
    editParams:
      - rowColumn: id
        apiParam: id
```

---

## Edge Cases & Considerations

### 1. Read-Only Rows
Some rows may be read-only (e.g., completed tasks, archived variables).
- **Solution**: Add row-level `editable` check via API metadata or config rule:
  ```yaml
  editCondition: "status != 'completed'"
  ```

### 2. Concurrent Edits
Another user might edit the same value.
- **Solution**: Show optimistic update, refresh on save. On conflict, show error: "Value was modified by another user. Refresh and try again."

### 3. Large JSON Values
Editing large JSON in a TUI modal is cumbersome.
- **Solution**: Offer "Open in $EDITOR" option (save to temp file, open in vim/nano, read back).

### 4. Undo
User might want to undo an edit.
- **Solution**: Phase N — add undo stack. Store {endpoint, oldValue, newValue}, allow `Ctrl+Z` to revert.

### 5. Bulk Edit
User might want to edit multiple rows at once.
- **Solution**: Phase N — multi-select rows (Spacebar), then `e` → apply same value to all selected.

---

## UX Flow Example

### Editing a Variable Value

1. User navigates to "process-variables" table
2. Sees:
   ```
   NAME       │ VALUE [E]    │ TYPE
   myVar      │ 123 [E]      │ Integer
   flag       │ true [E]     │ Boolean
   ```
3. Selects row "myVar" (cursor on row)
4. Presses `e`
5. Modal opens:
   ```
   ┌─ Edit Variable Value ──────┐
   │ Variable: myVar             │
   │ Type: Integer               │
   │                             │
   │ Current: 123                │
   │ New: █                      │
   │                             │
   │ <ctrl>+s Save  Esc Cancel   │
   └─────────────────────────────┘
   ```
6. Types `456` → presses `Ctrl+S`
7. Modal closes, footer shows: "Saving..."
8. API call succeeds → table updates to:
   ```
   myVar      │ 456 [E]      │ Integer
   ```
9. Footer shows: "✓ Saved" (flash for 2s)

### Editing with Multiple Editable Columns

1. User navigates to "task" table
2. Row has 2 editable columns: `assignee`, `priority`
3. Presses `e`
4. Footer shows quick-pick:
   ```
   Edit which field?  [1] assignee  [2] priority  [Esc] cancel
   ```
5. Presses `1`
6. Modal opens for "assignee" field
7. ... (same flow as above)

---

## Summary

### Key Principles
- **Config-driven**: All editability controlled via YAML config
- **Type-aware**: Different input UIs for different data types
- **Visual clarity**: Clear indicators for editable vs. read-only
- **Minimal keypresses**: Quick-pick menu for multi-column, direct modal for single-column
- **Safe**: Validation before save, error handling, flash feedback

### Next Steps
1. Review & approve this concept
2. Implement Phase 1 (foundation + visual indicators)
3. Implement Phase 2 (modal for string editing)
4. Test with process-variables table
5. Expand to other types & tables

---

## Open Questions for Discussion

1. **Visual marker preference**: `[E]` suffix vs. background color vs. both?
2. **Multi-column quick-pick**: Footer menu vs. inline overlay vs. cycle-through?
3. **Large values**: Offer external editor integration (`$EDITOR`) for JSON/long strings?
4. **Confirmation on save**: Always confirm, or only for destructive changes?
5. **Optimistic updates**: Update table immediately, or wait for API success?

---

**Status**: Concept defined, ready for review & implementation planning.
