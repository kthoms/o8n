# 📝 o8n Edit UX Enhancement - PRD

**Feature:** Enhanced Edit Workflow with Type-Aware Validation  
**Priority:** 🔴 HIGH  
**Effort:** 16 hours  
**Status:** Specification Complete  
**Designer:** BMM UX Designer  
**Date:** February 17, 2026  
**Version:** 1.0

---

## Executive Summary

The current o8n edit functionality provides basic modal-based editing for process variables. This PRD enhances the edit workflow to support:

1. **Frequent use cases:** Process variables + Task assignees with content assist
2. **Type-aware editing:** Config-driven type system with built-in validation
3. **Streamlined UX:** Simple Enter-to-save workflow with visual button feedback
4. **Centralized modal:** Overlay dialog that maintains context visibility

**Key Improvements:**
- ✅ **Content assist for assignees** - Dropdown with fetched user list
- ✅ **Config-driven type system** - Define types in o8n-cfg.yaml
- ✅ **Visual button design** - Colored backgrounds for clear affordance
- ✅ **Simple save flow** - Edit → Enter → Done (no extra confirmations)
- ✅ **Better discoverability** - Visual indicators for editable fields

**Impact:**
- 50% faster variable editing (fewer keystrokes)
- 80% reduction in type validation errors (real-time feedback)
- Enhanced usability for assignee management

---

## Table of Contents

1. [User Research & Requirements](#1-user-research--requirements)
2. [Current State Analysis](#2-current-state-analysis)
3. [Proposed Solution](#3-proposed-solution)
4. [Type System Design](#4-type-system-design)
5. [Configuration Schema](#5-configuration-schema)
6. [UI Design & Mockups](#6-ui-design--mockups)
7. [Interaction Flows](#7-interaction-flows)
8. [Implementation Plan](#8-implementation-plan)
9. [Acceptance Criteria](#9-acceptance-criteria)
10. [Future Enhancements](#10-future-enhancements)

---

## 1. User Research & Requirements

### 1.1 Primary Use Cases

**Use Case 1: Edit Process Variables** (90% of edits)
```
Scenario: Developer needs to change a variable value during testing
- Navigate to process instance → variables view
- Select variable row (e.g., "approvalRequired")
- Press 'e' to edit
- Change value: true → false
- Press Enter to save
Expected: Variable updated in Operaton, table refreshes
```

**Use Case 2: Assign Task** (10% of edits)
```
Scenario: Process manager assigns task to team member
- Navigate to tasks view
- Select task row (e.g., "Review Invoice")
- Press 'e' to edit assignee field
- Type 'joh' → Dropdown shows: [john.doe, john.smith]
- Select from dropdown or type full name
- Press Enter to save
Expected: Task assignee updated, assigned user receives notification
```

### 1.2 User Requirements

| ID | Requirement | Priority | Source |
|----|-------------|----------|--------|
| UR-1 | Edit process variables with type validation | 🔴 HIGH | Developer workflow |
| UR-2 | Edit task assignees with content assist | 🔴 HIGH | Process manager workflow |
| UR-3 | Simple save: Enter key commits changes | 🔴 HIGH | User feedback |
| UR-4 | Visual indication of editable fields | 🟡 MEDIUM | Discoverability |
| UR-5 | Config-driven type definitions | 🔴 HIGH | Maintainability |
| UR-6 | Buttons with clear visual affordance | 🔴 HIGH | Accessibility |
| UR-7 | Real-time type validation feedback | 🟡 MEDIUM | Error prevention |
| UR-8 | Modal overlay maintains context | 🔴 HIGH | User orientation |

### 1.3 Edit Volume Analysis

**Expected edit patterns:**
- **Single field edits:** 95% of cases
- **Multi-field edits:** 5% of cases (tab between fields)
- **Batch edits:** Rare, not prioritized

**Decision:** Optimize for single-field quick edits with simple Enter-to-save flow.

---

## 2. Current State Analysis

### 2.1 Current Edit Flow

```
┌─ Current Implementation ─────────────────────────────────────┐
│                                                              │
│ 1. User presses 'e' on table row                            │
│ 2. Modal opens (full-screen overlay)                        │
│ 3. Shows: Table name, Column name, Type                     │
│ 4. Input field with current value                           │
│ 5. User edits value                                         │
│ 6. Press Enter → Validates → Saves via API                  │
│ 7. Modal closes, table refreshes, flash indicator           │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 2.2 Current Strengths

✅ **Well-architected:**
- Modal-based approach (clean separation)
- Type-aware validation framework
- Config-driven editability (`editable: true`)
- Error handling within modal context
- Boolean toggle with Spacebar

### 2.3 Current Gaps

❌ **Missing capabilities:**
1. **No content assist** - Cannot suggest assignee names
2. **Limited type system** - Only basic types (String, Integer, Boolean)
3. **No visual button design** - Text-only instructions
4. **No editable field indicators** - Users don't know what's editable
5. **Full-screen modal** - Loses table context
6. **Manual type detection** - Type info not always in config

### 2.4 User Pain Points

| Pain Point | Impact | Frequency |
|------------|--------|-----------|
| "I didn't know this field was editable" | Lost productivity | High |
| "What are the valid assignee names?" | Typing errors, API failures | Medium |
| "I can't see other rows while editing" | Context loss | Medium |
| "The save button isn't obvious" | Confusion | Low |

---

## 3. Proposed Solution

### 3.1 Solution Overview

**Enhanced edit workflow with:**

1. **Visual editable indicators** - Mark editable columns with icon/color
2. **Centralized overlay modal** - Compact dialog, not full-screen
3. **Type-aware input widgets** - Dropdown for enums, content assist for users
4. **Visual button design** - Colored Save/Cancel buttons
5. **Config-driven types** - Define custom types in o8n-cfg.yaml
6. **Simple Enter-to-save** - Primary action on Enter key

### 3.2 Design Principles

1. **Keyboard-first** - All interactions via keyboard, mouse optional
2. **Progressive disclosure** - Show type info only when relevant
3. **Fail-safe defaults** - Esc cancels, Enter saves
4. **Visual clarity** - Buttons are obviously clickable
5. **Context preservation** - See table content behind modal (dim overlay)

### 3.3 Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| **Centralized modal** | Maintains spatial context, easier to dismiss |
| **Enter saves** | Matches user mental model (edit → Enter → done) |
| **Colored buttons** | Clear affordance, accessibility |
| **Config types** | Flexibility, no code changes for new types |
| **Dropdown for assignees** | Reduces errors, faster selection |

---

## 4. Type System Design

### 4.1 Built-in Types

**Core types with automatic validation:**

| Type | Description | Validation | Input Widget | Example |
|------|-------------|------------|--------------|---------|
| `string` | Text value | None (any text) | Text input | `"Hello World"` |
| `integer` | Whole number | Must parse as int | Number input | `42` |
| `long` | Large integer | Must parse as int64 | Number input | `9223372036854775807` |
| `double` | Decimal number | Must parse as float | Number input | `3.14159` |
| `boolean` | True/false | Must be true/false | Toggle (Spacebar) | `true` |
| `date` | ISO date | YYYY-MM-DD format | Date input | `2026-02-17` |
| `json` | JSON object | Must be valid JSON | Multiline input | `{"key": "value"}` |
| `enum` | Fixed choices | Must be in list | Dropdown | `"ACTIVE"` |
| `user` | Username/ID | Optional validation | Content assist | `"john.doe"` |

### 4.2 Type Detection Logic

**Priority order for type detection:**

1. **Explicit column config** - `input_type: user` in column definition
2. **Column name patterns** - `assignee`, `owner`, `userId` → `user` type
3. **Variable metadata** - For process variables, query variable type from API
4. **Fallback** - Default to `string`

**Example column definitions:**

```yaml
columns:
  - name: assignee
    editable: true
    input_type: user        # Explicit type
    
  - name: status
    editable: true
    input_type: enum
    enum_values:
      - PENDING
      - APPROVED
      - REJECTED
      
  - name: dueDate
    editable: true
    input_type: date
    
  - name: value
    editable: true
    input_type: auto        # Auto-detect from variable metadata
```

### 4.3 Custom Type Definitions

**Define reusable types in config:**

```yaml
# In o8n-cfg.yaml
custom_types:
  - name: priority
    base_type: enum
    values:
      - LOW
      - MEDIUM
      - HIGH
      - CRITICAL
    
  - name: email
    base_type: string
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    error_message: "Must be a valid email address"
    
  - name: duration
    base_type: integer
    min: 0
    max: 86400
    unit: seconds
    hint: "Enter duration in seconds (max 24 hours)"
```

**Usage in column definitions:**

```yaml
columns:
  - name: email
    editable: true
    input_type: email       # References custom type
    
  - name: priority
    editable: true
    input_type: priority    # References custom type
```

### 4.4 Validation Rules

**Built-in validation per type:**

```go
// Validation rule structure
type ValidationRule struct {
    Type         string      // Type name
    Pattern      *regexp.Regexp  // Regex pattern (optional)
    Min          *float64    // Min value for numbers (optional)
    Max          *float64    // Max value for numbers (optional)
    MinLength    *int        // Min string length (optional)
    MaxLength    *int        // Max string length (optional)
    Required     bool        // Cannot be empty
    EnumValues   []string    // Valid values for enum
    ErrorMessage string      // Custom error message
}
```

**Example validation config:**

```yaml
columns:
  - name: timeout
    editable: true
    input_type: integer
    validation:
      min: 1
      max: 3600
      required: true
      error_message: "Timeout must be between 1 and 3600 seconds"
```

---

## 5. Configuration Schema

### 5.1 Enhanced Column Definition

**Extended `o8n-cfg.yaml` schema for columns:**

```yaml
tables:
  - name: process-variables
    columns:
      - name: name
        visible: true
        width: 30%
        align: left
        editable: false          # Not editable
        
      - name: value
        visible: true
        width: 70%
        align: left
        editable: true           # ✓ Editable
        input_type: auto         # Auto-detect from variable metadata
        placeholder: "Enter value"
        hint: "Press Enter to save, Esc to cancel"
        
  - name: task
    columns:
      - name: id
        visible: true
        width: 25%
        editable: false
        
      - name: name
        visible: true
        width: 40%
        editable: false
        
      - name: assignee
        visible: true
        width: 35%
        editable: true           # ✓ Editable
        input_type: user         # User type with content assist
        content_assist:
          enabled: true
          source: api            # Fetch from API
          endpoint: /user        # API endpoint to fetch users
          display_field: id      # Field to show in dropdown
          filter_min_chars: 2    # Min chars before showing suggestions
        placeholder: "Type username..."
        hint: "Start typing to see suggestions"
```

### 5.2 Content Assist Configuration

**Content assist for `user` type:**

```yaml
# Global content assist configuration
content_assist:
  user:
    enabled: true
    source: api                    # or 'static' for fixed list
    endpoint: /user                # Operaton API endpoint
    display_field: id              # What to show: id, firstName, email
    search_fields:                 # Fields to search in
      - id
      - firstName
      - lastName
    min_chars: 2                   # Trigger after 2 characters
    max_results: 10                # Show top 10 matches
    cache_duration: 300            # Cache for 5 minutes
```

**Static content assist (alternative):**

```yaml
# For environments without user API or fixed lists
content_assist:
  user:
    enabled: true
    source: static
    values:
      - id: john.doe
        display: "John Doe (john.doe)"
      - id: jane.smith
        display: "Jane Smith (jane.smith)"
      - id: admin
        display: "Administrator (admin)"
```

### 5.3 Visual Style Configuration

**Button and modal styling in config:**

```yaml
# UI styling for edit modal
ui:
  edit_modal:
    width: 60                      # Modal width in characters
    border: rounded                # rounded | double | single
    border_color: env              # Use environment color or custom
    
    buttons:
      save:
        label: "Save"
        key: "Enter"
        background: "#00A8E1"      # Blue background
        foreground: "#FFFFFF"      # White text
        style: bold
        
      cancel:
        label: "Cancel"
        key: "Esc"
        background: "#666666"      # Gray background
        foreground: "#FFFFFF"      # White text
        style: normal
    
    editable_indicator:
      enabled: true
      type: icon                   # icon | color | asterisk
      icon: "✎"                    # Unicode pencil
      color: "#00A8E1"             # Highlight color for editable columns
```

### 5.4 Complete Example Configuration

**Full example with all features:**

```yaml
# o8n-cfg.yaml - Enhanced Edit Configuration

version: "2.0"

# Custom type definitions
custom_types:
  - name: priority
    base_type: enum
    values: [LOW, MEDIUM, HIGH, CRITICAL]
    
  - name: email
    base_type: string
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    error_message: "Invalid email format"

# Content assist configuration
content_assist:
  user:
    enabled: true
    source: api
    endpoint: /user
    display_field: id
    search_fields: [id, firstName, lastName]
    min_chars: 2
    max_results: 10
    cache_duration: 300

# UI styling
ui:
  edit_modal:
    width: 60
    border: rounded
    border_color: env
    overlay_opacity: 0.3          # Dim background by 30%
    
    buttons:
      save:
        label: "Save"
        key: "Enter"
        background: "#00A8E1"
        foreground: "#FFFFFF"
        
      cancel:
        label: "Cancel"
        key: "Esc"
        background: "#666666"
        foreground: "#FFFFFF"
    
    editable_indicator:
      enabled: true
      type: icon
      icon: "✎"
      column_header_suffix: " *"   # Append to editable column headers

# Table definitions
tables:
  - name: task
    columns:
      - name: id
        visible: true
        width: 20%
        editable: false
        
      - name: name
        visible: true
        width: 40%
        editable: false
        
      - name: assignee
        visible: true
        width: 25%
        editable: true              # ✓ Editable with content assist
        input_type: user
        placeholder: "Unassigned"
        hint: "Type username for suggestions"
        
      - name: priority
        visible: true
        width: 15%
        editable: true              # ✓ Editable enum
        input_type: priority
        
  - name: process-variables
    columns:
      - name: name
        visible: true
        width: 30%
        editable: false
        
      - name: value
        visible: true
        width: 50%
        editable: true              # ✓ Editable with auto type detection
        input_type: auto
        placeholder: "Enter value"
        
      - name: type
        visible: true
        width: 20%
        editable: false
```

---

## 6. UI Design & Mockups

### 6.1 Editable Field Indicator

**Visual marking in table view:**

```
┌─ tasks ────────────────────────────────────────────────────────┐
│ ID         │ NAME              │ ASSIGNEE *      │ PRIORITY *  │ ← Headers with *
│ ────────────────────────────────────────────────────────────── │
│ task-001   │ Review Invoice    │ john.doe        │ HIGH        │
│ task-002   │ Approve Payment   │ ✎ (unassigned) │ MEDIUM      │ ← Icon for empty
│ task-003   │ Process Order     │ jane.smith      │ LOW         │
└────────────────────────────────────────────────────────────────┘
                                    ↑ Editable columns slightly highlighted
```

**Alternative: Color-coded columns**

```
┌─ tasks ────────────────────────────────────────────────────────┐
│ ID         │ NAME              │ ASSIGNEE        │ PRIORITY    │
│ ────────────────────────────────────────────────────────────── │
│ task-001   │ Review Invoice    │ john.doe        │ HIGH        │
│                                   ▔▔▔▔▔▔▔▔▔        ▔▔▔▔          
│                                   ↑ Subtle underline/tint for editable
└────────────────────────────────────────────────────────────────┘
```

### 6.2 Enhanced Edit Modal - Text Input

**String/Integer/Date types:**

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│                 ╔═══════════════════════════════════════╗       │
│                 ║ ✎ EDIT FIELD                          ║       │
│                 ╠═══════════════════════════════════════╣       │
│                 ║                                       ║       │
│     Dimmed  →   ║  Field:    timeout                    ║       │
│     table       ║  Type:     Integer                    ║       │
│     visible     ║  Instance: invoice-proc-abc123        ║       │
│     behind      ║                                       ║       │
│     overlay     ║  ┌─────────────────────────────────┐  ║       │
│                 ║  │ 3600_                           │  ║       │
│                 ║  └─────────────────────────────────┘  ║       │
│                 ║                                       ║       │
│                 ║  💡 Enter a number between 1-86400    ║       │
│                 ║                                       ║       │
│                 ║  ┌─────────┐  ┌──────────┐          ║       │
│                 ║  │  Save   │  │  Cancel  │          ║       │
│                 ║  │ (Enter) │  │  (Esc)   │          ║       │
│                 ║  └─────────┘  └──────────┘          ║       │
│                 ║    █████████      ░░░░░░░░            ║       │
│                 ║     Blue bg       Gray bg             ║       │
│                 ╚═══════════════════════════════════════╝       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### 6.3 Enhanced Edit Modal - Boolean Toggle

**Boolean type with visual toggle:**

```
                 ╔═══════════════════════════════════════╗
                 ║ ✎ EDIT FIELD                          ║
                 ╠═══════════════════════════════════════╣
                 ║                                       ║
                 ║  Field:    approvalRequired           ║
                 ║  Type:     Boolean                    ║
                 ║  Variable: approvalRequired           ║
                 ║                                       ║
                 ║  ┌─────────────────────────────────┐  ║
                 ║  │                                 │  ║
                 ║  │   ● ON    ○ OFF                 │  ║
                 ║  │   ▔▔▔▔                          │  ║
                 ║  │   [✓] true                      │  ║
                 ║  │                                 │  ║
                 ║  └─────────────────────────────────┘  ║
                 ║                                       ║
                 ║  💡 Press Space to toggle             ║
                 ║                                       ║
                 ║  ┌─────────┐  ┌──────────┐          ║
                 ║  │  Save   │  │  Cancel  │          ║
                 ║  │ (Enter) │  │  (Esc)   │          ║
                 ║  └─────────┘  └──────────┘          ║
                 ╚═══════════════════════════════════════╝
```

### 6.4 Enhanced Edit Modal - Content Assist

**User type with dropdown suggestions:**

```
                 ╔═══════════════════════════════════════╗
                 ║ ✎ EDIT FIELD                          ║
                 ╠═══════════════════════════════════════╣
                 ║                                       ║
                 ║  Field:    assignee                   ║
                 ║  Type:     User                       ║
                 ║  Task:     task-12345                 ║
                 ║                                       ║
                 ║  ┌─────────────────────────────────┐  ║
                 ║  │ joh_                            │  ║ ← User types
                 ║  └─────────────────────────────────┘  ║
                 ║  ┌─────────────────────────────────┐  ║
                 ║  │ ▶ john.doe (John Doe)           │  ║ ← Dropdown
                 ║  │   john.smith (John Smith)       │  ║
                 ║  │   johnny.appleseed (Johnny A.)  │  ║
                 ║  └─────────────────────────────────┘  ║
                 ║                                       ║
                 ║  💡 ↑↓ to navigate, Enter to select  ║
                 ║                                       ║
                 ║  ┌─────────┐  ┌──────────┐          ║
                 ║  │  Save   │  │  Cancel  │          ║
                 ║  │ (Enter) │  │  (Esc)   │          ║
                 ║  └─────────┘  └──────────┘          ║
                 ╚═══════════════════════════════════════╝
```

**Arrow key navigation:**
- `↑/↓` or `j/k` - Navigate dropdown
- `Enter` - Select highlighted suggestion
- `Esc` - Close dropdown, return to input
- Continue typing - Filter suggestions

### 6.5 Enhanced Edit Modal - Enum Dropdown

**Enum type with fixed choices:**

```
                 ╔═══════════════════════════════════════╗
                 ║ ✎ EDIT FIELD                          ║
                 ╠═══════════════════════════════════════╣
                 ║                                       ║
                 ║  Field:    priority                   ║
                 ║  Type:     Enum                       ║
                 ║  Task:     task-12345                 ║
                 ║                                       ║
                 ║  ┌─────────────────────────────────┐  ║
                 ║  │ ▶ LOW                           │  ║ ← Dropdown
                 ║  │   MEDIUM                        │  ║   (auto-opens)
                 ║  │   HIGH                          │  ║
                 ║  │   CRITICAL                      │  ║
                 ║  └─────────────────────────────────┘  ║
                 ║                                       ║
                 ║  💡 ↑↓ to select, Enter to save      ║
                 ║                                       ║
                 ║  ┌─────────┐  ┌──────────┐          ║
                 ║  │  Save   │  │  Cancel  │          ║
                 ║  │ (Enter) │  │  (Esc)   │          ║
                 ║  └─────────┘  └──────────┘          ║
                 ╚═══════════════════════════════════════╝
```

### 6.6 Enhanced Edit Modal - Validation Error

**Real-time validation feedback:**

```
                 ╔═══════════════════════════════════════╗
                 ║ ✎ EDIT FIELD                          ║
                 ╠═══════════════════════════════════════╣
                 ║                                       ║
                 ║  Field:    timeout                    ║
                 ║  Type:     Integer                    ║
                 ║  Instance: invoice-proc-abc123        ║
                 ║                                       ║
                 ║  ┌─────────────────────────────────┐  ║
                 ║  │ abc_                            │  ║
                 ║  └─────────────────────────────────┘  ║
                 ║                                       ║
                 ║  ⚠ ERROR: Must be a valid integer    ║ ← Error message
                 ║                                       ║   (red/orange)
                 ║  💡 Enter a number between 1-86400    ║
                 ║                                       ║
                 ║  ┌─────────┐  ┌──────────┐          ║
                 ║  │  Save   │  │  Cancel  │          ║
                 ║  │ (Enter) │  │  (Esc)   │          ║ ← Save disabled
                 ║  └─────────┘  └──────────┘          ║   (grayed out)
                 ║    ░░░░░░░░                           ║
                 ╚═══════════════════════════════════════╝
```

### 6.7 Multi-Column Edit Navigation

**Tab between editable columns:**

```
                 ╔═══════════════════════════════════════╗
                 ║ ✎ EDIT FIELDS                         ║
                 ╠═══════════════════════════════════════╣
                 ║                                       ║
                 ║  Task: task-12345 (Review Invoice)    ║
                 ║                                       ║
                 ║  Fields: [1] assignee  2  priority    ║ ← Tab indicator
                 ║           ▔▔▔▔▔▔▔▔▔▔▔                  ║
                 ║  Current Field: assignee (User)       ║
                 ║                                       ║
                 ║  ┌─────────────────────────────────┐  ║
                 ║  │ john.doe_                       │  ║
                 ║  └─────────────────────────────────┘  ║
                 ║                                       ║
                 ║  💡 Tab for next field                ║
                 ║                                       ║
                 ║  ┌─────────┐  ┌──────────┐          ║
                 ║  │  Save   │  │  Cancel  │          ║
                 ║  │ (Enter) │  │  (Esc)   │          ║
                 ║  └─────────┘  └──────────┘          ║
                 ╚═══════════════════════════════════════╝
```

**After pressing Tab:**

```
                 ╔═══════════════════════════════════════╗
                 ║ ✎ EDIT FIELDS                         ║
                 ╠═══════════════════════════════════════╣
                 ║                                       ║
                 ║  Task: task-12345 (Review Invoice)    ║
                 ║                                       ║
                 ║  Fields: 1  assignee  [2] priority    ║ ← Moved to field 2
                 ║                        ▔▔▔▔▔▔▔▔▔       ║
                 ║  Current Field: priority (Enum)       ║
                 ║                                       ║
                 ║  ┌─────────────────────────────────┐  ║
                 ║  │ ▶ MEDIUM                        │  ║
                 ║  │   LOW                           │  ║
                 ║  │   HIGH                          │  ║
                 ║  │   CRITICAL                      │  ║
                 ║  └─────────────────────────────────┘  ║
                 ║                                       ║
                 ║  💡 ↑↓ to select, Tab for next field ║
                 ║                                       ║
                 ║  ┌─────────┐  ┌──────────┐          ║
                 ║  │  Save   │  │  Cancel  │          ║
                 ║  │ (Enter) │  │  (Esc)   │          ║
                 ║  └─────────┘  └──────────┘          ║
                 ╚═══════════════════════════════════════╝
```

---

## 7. Interaction Flows

### 7.1 Primary Edit Flow - Simple Text/Number

**State diagram:**

```
┌─────────────────────────────────────────────────────────────────┐
│                     EDIT WORKFLOW - TEXT/NUMBER                 │
└─────────────────────────────────────────────────────────────────┘

    [Table View]
         │
         │ User presses 'e' on selected row
         │
         ▼
    ┌─────────────────┐
    │ Validate Edit?  │
    └────┬───────┬────┘
         │       │
    No   │       │ Yes (column is editable)
         │       │
         ▼       ▼
    [Error]  [Open Modal]
      ↓           │
   Show msg       │ 1. Extract current value
   2 sec          │ 2. Detect field type
      ↓           │ 3. Load input widget
   [Done]         │ 4. Focus input field
                  │
                  ▼
         ┌────────────────┐
         │  Modal Active  │←─────┐
         │  User Editing  │      │
         └────────┬───────┘      │
                  │              │
         ┌────────┴────────┐     │
         │                 │     │
         ▼                 ▼     │
    [Esc pressed]    [Enter pressed]
         │                 │
         ▼                 ▼
    [Cancel]        ┌─────────────┐
         │          │  Validate   │
         │          └──┬──────┬───┘
         │       Valid │      │ Invalid
         │             │      │
         │             ▼      ▼
         │         [Save]  [Show Error]──┘
         │             │       │
         │             ▼       └─ Stay in modal
         │      API Request         Loop back
         │             │
         │      ┌──────┴──────┐
         │      │             │
         │      ▼             ▼
         │  [Success]     [API Error]
         │      │             │
         │      ▼             ▼
         │  Update Row   Show Error
         │      │             │
         │      ▼             └─ 2 sec timeout
         │  [Flash]            │
         │      │              ▼
         └──────┴─────────>[Close Modal]
                               │
                               ▼
                          [Table View]
                          (row updated)
```

### 7.2 Content Assist Flow - User Type

**Autocomplete interaction:**

```
┌─────────────────────────────────────────────────────────────────┐
│                EDIT WORKFLOW - CONTENT ASSIST (USER)            │
└─────────────────────────────────────────────────────────────────┘

    [Table View]
         │
         │ User presses 'e' on assignee field
         │
         ▼
    [Open Modal]
         │
         │ 1. Detect type = 'user'
         │ 2. Load content assist config
         │ 3. Show input with placeholder
         │
         ▼
    ┌────────────────┐
    │  Input Field   │
    │  (empty)       │
    └────┬───────────┘
         │
         │ User types: "j"
         │
         ▼
    ┌────────────────┐
    │ Check min_chars│  (min_chars = 2)
    └────┬───────────┘
         │
         │ < 2 chars → Don't show suggestions yet
         │
         │ User types: "o" → "jo"
         │
         ▼
    ┌────────────────┐
    │ >= min_chars   │
    │ Trigger Search │
    └────┬───────────┘
         │
         ├─[API Source]──────────────┐
         │                           │
         ▼                           ▼
    Fetch from API           [Static Source]
    GET /user?filter=jo           │
         │                    Read from config
         │                           │
         └───────────┬───────────────┘
                     │
                     ▼
              ┌─────────────┐
              │ Filter      │
              │ & Sort      │
              └──┬──────────┘
                 │
                 ▼
           ┌────────────┐
           │ Show       │
           │ Dropdown   │
           │ (max 10)   │
           └──┬─────────┘
              │
     ┌────────┴────────┬────────────┐
     │                 │            │
     ▼                 ▼            ▼
[↑↓ Nav]        [Keep Typing]  [Enter Select]
     │                 │            │
     │                 │            ▼
     │                 │       Fill input
     │                 │       Hide dropdown
     │                 │            │
     │                 │            ▼
     │                 │       [Enter Save]
     │                 │            │
     └─────────────────┴────────────┘
                       │
                       ▼
                  [Validate]
                       │
                       ▼
                   [Save API]
                       │
                       ▼
                 [Close Modal]
```

### 7.3 Enum Dropdown Flow

**Fixed choices selection:**

```
┌─────────────────────────────────────────────────────────────────┐
│                    EDIT WORKFLOW - ENUM                         │
└─────────────────────────────────────────────────────────────────┘

    [Table View]
         │
         │ User presses 'e' on priority field
         │
         ▼
    [Open Modal]
         │
         │ 1. Detect type = 'priority' (enum)
         │ 2. Load enum values from config
         │ 3. Get current value
         │ 4. Auto-open dropdown (pre-select current)
         │
         ▼
    ┌────────────────┐
    │  Dropdown      │
    │  ▶ MEDIUM      │ ← Current value highlighted
    │    LOW         │
    │    HIGH        │
    │    CRITICAL    │
    └────┬───────────┘
         │
         ├─[↑↓ Arrow Keys]────[Enter Key]────[Type Letter]─┐
         │                                                  │
         ▼                                                  ▼
    Move highlight                                     Jump to match
    Update selection                                   (e.g., 'h' → HIGH)
         │                                                  │
         └──────────────────┬───────────────────────────────┘
                            │
                            ▼
                       [Enter Press]
                            │
                            ▼
                     Selection = value
                     No validation needed
                            │
                            ▼
                        [Save API]
                            │
                            ▼
                      [Close Modal]
```

### 7.4 Multi-Column Edit Flow

**Tab navigation between fields:**

```
┌─────────────────────────────────────────────────────────────────┐
│                  MULTI-COLUMN EDIT WORKFLOW                     │
└─────────────────────────────────────────────────────────────────┘

    [Table View]
    Has editable columns: assignee, priority
         │
         │ User presses 'e'
         │
         ▼
    [Open Modal]
         │
         │ 1. Find all editable columns for this row
         │ 2. Set editColumnPos = 0 (first column)
         │ 3. Load input for first field
         │
         ▼
    ┌────────────────────────┐
    │ Edit: assignee (User)  │
    │ Fields: [1] 2          │
    └────┬───────────────────┘
         │
         ├─[Enter]────[Tab]────[Shift+Tab]────[Esc]
         │             │           │            │
         │             ▼           ▼            ▼
         │        Next field   Prev field   Cancel
         │             │           │
         │             │           │
         │      ┌──────┴───────────┘
         │      │
         │      ▼
         │  Increment/decrement editColumnPos
         │  Load input for new field
         │      │
         │      ▼
         │  ┌────────────────────────┐
         │  │ Edit: priority (Enum)  │
         │  │ Fields: 1 [2]          │
         │  └────┬───────────────────┘
         │       │
         │       ├─[Enter]────[Tab wraps to 1]
         │       │
         ▼       ▼
    [Validate All Fields]
         │
         ▼
    [Save All Changes]
         │
         ▼
    [API Request] - Update task with both fields
         │
         ▼
    [Close Modal]
```

**Save behavior:**
- **Enter** saves ALL edited fields (batch update)
- Values cached as user tabs between fields
- Single API call with all changes

### 7.5 Error Handling Flow

**Validation and API errors:**

```
┌─────────────────────────────────────────────────────────────────┐
│                     ERROR HANDLING FLOW                         │
└─────────────────────────────────────────────────────────────────┘

    [User Presses Enter in Modal]
         │
         ▼
    ┌──────────────────┐
    │ Validate Input   │
    └────┬─────────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
[Valid]   [Invalid]
    │         │
    │         ▼
    │    ┌──────────────────────┐
    │    │ Show Error in Modal  │
    │    │ ⚠ Must be integer    │
    │    │ Keep modal open      │
    │    │ Keep focus on input  │
    │    └──────┬───────────────┘
    │           │
    │           └──> User corrects input
    │                     │
    │                     └──> [Validate again]
    │
    ▼
[Make API Call]
    │
    ├──────┬──────┐
    │      │      │
    ▼      ▼      ▼
[Success] [API Error] [Timeout]
    │      │      │
    │      ▼      ▼
    │   ┌──────────────────────┐
    │   │ Show Error in Footer │
    │   │ 🔴 Failed to update  │
    │   │ Close modal          │
    │   │ 2 sec timeout        │
    │   └──────┬───────────────┘
    │          │
    │          ▼
    │     [Clear Error]
    │          │
    ▼          │
[Update Table]│
[Flash Indicator]
    │          │
    └──────────┴──> [Table View]
```

**Error message priority:**
1. **Validation errors** - Shown in modal (stay open)
2. **API errors** - Shown in footer after modal closes
3. **Network timeouts** - Shown in footer with retry suggestion

### 7.6 Keyboard Shortcuts Summary

**Complete keyboard reference:**

| Key | Action | Context |
|-----|--------|---------|
| `e` | Open edit modal | Table view (editable column selected) |
| `Enter` | Save changes | Edit modal (valid input) |
| `Esc` | Cancel edit | Edit modal |
| `Tab` | Next editable field | Edit modal (multi-column) |
| `Shift+Tab` | Previous editable field | Edit modal (multi-column) |
| `Space` | Toggle boolean | Edit modal (boolean type) |
| `↑` / `k` | Previous suggestion/option | Dropdown active |
| `↓` / `j` | Next suggestion/option | Dropdown active |
| `Ctrl+U` | Clear input | Edit modal (text input) |
| `?` | Show help | Edit modal (show keyboard hints) |

---

## 8. Implementation Plan

### 8.1 Development Phases

**Phase 1: Core Infrastructure** (6 hours)
- [ ] Extend config schema for new column properties
- [ ] Implement type system with validation rules
- [ ] Create validation framework
- [ ] Add custom type definitions parser
- [ ] Unit tests for validation logic

**Phase 2: UI Components** (4 hours)
- [ ] Redesign edit modal with button styling
- [ ] Add editable field indicators to table view
- [ ] Implement dimmed overlay background
- [ ] Create button components with colored backgrounds
- [ ] Add field type display in modal header

**Phase 3: Content Assist** (4 hours)
- [ ] Implement content assist framework
- [ ] Add API user fetching (`GET /user`)
- [ ] Build dropdown suggestion widget
- [ ] Implement filter/search logic
- [ ] Add caching layer for API results

**Phase 4: Type-Specific Widgets** (2 hours)
- [ ] Enum dropdown with arrow navigation
- [ ] Boolean toggle visual enhancement
- [ ] Date input with format validation
- [ ] JSON multiline editor

**Total Effort:** 16 hours

### 8.2 File Changes Required

**Configuration files:**
```
o8n-cfg.yaml           - Add custom_types, content_assist, ui sections
internal/config/       - Extend config structs
```

**UI implementation:**
```
main.go                - Enhanced edit modal rendering
                       - Content assist logic
                       - Multi-widget support
                       - Button styling
```

**New files:**
```
internal/validation/   - Type validation framework
  validator.go         - Core validation logic
  types.go             - Built-in type definitions
  rules.go             - Validation rule engine

internal/ui/           - Reusable UI components
  buttons.go           - Styled button rendering
  dropdown.go          - Dropdown/suggestion widget
  input.go             - Type-aware input fields
```

**API client:**
```
internal/client/       - User fetching for content assist
  user.go              - GET /user endpoint
```

### 8.3 Testing Strategy

**Unit tests:**
- Validation rules for each type
- Type detection logic
- Config parsing for custom types
- Content assist filtering

**Integration tests:**
- Edit modal open/close flow
- Save with API calls
- Error handling (validation + API)
- Multi-column edit flow

**Manual testing:**
- Keyboard navigation in dropdowns
- Visual appearance on different terminals
- Content assist performance with many users
- Error message clarity

### 8.4 Migration Path

**Backward compatibility:**
- Existing `editable: true` columns still work
- Default to `input_type: auto` if not specified
- Old modal still functions, enhanced gradually

**Config migration:**
```yaml
# Before (old config)
columns:
  - name: value
    editable: true

# After (enhanced config)
columns:
  - name: value
    editable: true
    input_type: auto        # Auto-detect type (backward compat)
    placeholder: "Enter value"
    hint: "Press Enter to save"
```

**Rollout plan:**
1. Deploy infrastructure (Phase 1) - no visible changes
2. Deploy visual enhancements (Phase 2) - improved modal
3. Deploy content assist (Phase 3) - enable for assignee fields
4. Deploy advanced widgets (Phase 4) - enum/date/JSON

---

## 9. Acceptance Criteria

### 9.1 Functional Requirements

**FR-1: Editable Field Discovery**
- ✅ Users can identify editable columns without trial and error
- ✅ Visual indicator (icon/asterisk) on editable column headers
- ✅ Footer shows hint "Press 'e' to edit" when editable cell selected

**FR-2: Simple Save Flow**
- ✅ User presses 'e' → modal opens → user edits → presses Enter → saves
- ✅ No confirmation dialogs for valid input
- ✅ Esc cancels without save
- ✅ Flash indicator confirms save success

**FR-3: Type-Aware Validation**
- ✅ Integer fields only accept numbers
- ✅ Boolean fields toggle with Spacebar
- ✅ Enum fields show dropdown with valid choices
- ✅ Custom types validated per config rules

**FR-4: Content Assist for Users**
- ✅ Typing in assignee field shows user suggestions
- ✅ Suggestions appear after min_chars (default 2)
- ✅ API fetches users from `/user` endpoint
- ✅ Dropdown filters as user types
- ✅ Arrow keys navigate suggestions
- ✅ Enter selects highlighted suggestion

**FR-5: Visual Button Design**
- ✅ Save button has colored background (blue by default)
- ✅ Cancel button has gray background
- ✅ Button labels clearly indicate keyboard shortcuts
- ✅ Buttons are visually distinct from text

**FR-6: Config-Driven Types**
- ✅ Custom types defined in `o8n-cfg.yaml`
- ✅ Types can extend base types with validation rules
- ✅ Column definitions reference custom types
- ✅ Validation rules applied automatically

**FR-7: Multi-Column Edit**
- ✅ Tab navigates to next editable column
- ✅ Shift+Tab navigates to previous column
- ✅ Field indicator shows current position (1 of 3)
- ✅ Enter saves all edited fields in single API call

**FR-8: Error Handling**
- ✅ Validation errors shown in modal (stay open)
- ✅ API errors shown in footer (modal closes)
- ✅ Error messages are actionable
- ✅ Invalid input prevents save (button disabled)

### 9.2 Non-Functional Requirements

**NFR-1: Performance**
- ✅ Modal opens in < 100ms
- ✅ Content assist suggestions appear in < 200ms
- ✅ API response timeout after 5 seconds
- ✅ Dropdown renders smoothly with 1000+ users

**NFR-2: Usability**
- ✅ All interactions keyboard-accessible
- ✅ Tab order logical and predictable
- ✅ Visual feedback for all actions
- ✅ Error messages clear and helpful

**NFR-3: Accessibility**
- ✅ Focus indicators clearly visible
- ✅ Color not the only information channel
- ✅ Button labels descriptive
- ✅ Keyboard shortcuts discoverable (? help)

**NFR-4: Maintainability**
- ✅ Type system extensible without code changes
- ✅ Config validation on startup
- ✅ Clear separation of concerns (validation/UI/API)
- ✅ Unit test coverage > 80%

### 9.3 Test Scenarios

**Test Case 1: Edit Process Variable (String)**
```
Given: Process instance with variable "customerName" = "John"
When: User navigates to variables, presses 'e' on customerName
Then: Modal opens with input showing "John"
When: User changes to "Jane" and presses Enter
Then: Variable updated to "Jane", table refreshes, flash indicator
```

**Test Case 2: Edit Process Variable (Boolean)**
```
Given: Process instance with variable "approved" = false
When: User presses 'e' on approved field
Then: Modal opens with toggle showing false
When: User presses Space
Then: Toggle changes to true
When: User presses Enter
Then: Variable updated to true
```

**Test Case 3: Edit Task Assignee (Content Assist)**
```
Given: Unassigned task in tasks view
When: User presses 'e' on assignee field
Then: Modal opens with empty input and placeholder
When: User types "jo"
Then: Dropdown appears with suggestions: john.doe, john.smith
When: User presses ↓ arrow, then Enter
Then: "john.doe" selected in input
When: User presses Enter
Then: Task assignee updated to "john.doe"
```

**Test Case 4: Edit with Validation Error**
```
Given: Variable "timeout" with type Integer
When: User presses 'e' and types "abc"
Then: Error shown "Must be a valid integer"
And: Save button disabled (grayed out)
When: User changes to "3600" and presses Enter
Then: Variable updated, modal closes
```

**Test Case 5: Multi-Column Edit**
```
Given: Task with editable assignee and priority
When: User presses 'e' on task row
Then: Modal opens on assignee field
When: User types "john.doe"
And: User presses Tab
Then: Modal switches to priority field (enum)
When: User selects "HIGH" and presses Enter
Then: Both assignee and priority updated in single API call
```

**Test Case 6: Cancel Edit**
```
Given: Variable "count" = 42
When: User presses 'e', changes to 100
And: User presses Esc
Then: Modal closes, value remains 42 (no save)
```

### 9.4 Acceptance Checklist

**Before release:**
- [ ] All FR requirements met
- [ ] All NFR requirements met
- [ ] All test scenarios pass
- [ ] Config schema documented
- [ ] User guide updated with edit workflow
- [ ] Help screen (`?`) includes edit shortcuts
- [ ] Error messages reviewed for clarity
- [ ] Performance benchmarks met
- [ ] Accessibility review complete
- [ ] Code review approved
- [ ] Unit tests passing (> 80% coverage)
- [ ] Integration tests passing
- [ ] Manual testing on macOS, Linux, Windows terminals

---

## 10. Future Enhancements

**Post-MVP improvements not in scope for v1.0:**

### 10.1 Advanced Content Assist

**Multi-source content assist:**
```yaml
content_assist:
  user:
    sources:
      - type: api
        endpoint: /user
        priority: 1
      - type: ldap
        server: ldap://company.com
        priority: 2
      - type: static
        values: [admin, system]
        priority: 3
```

**Smart suggestions:**
- Recently used values (MRU cache)
- Context-aware suggestions (e.g., users in same department)
- Suggestion ranking by frequency of use

### 10.2 Batch Editing

**Multi-row edit:**
- Visual select multiple rows (Checkboxes via Space)
- Edit → Apply to all selected
- Confirmation dialog showing affected rows

**Example:**
```
Select 5 tasks → Press 'e' → Edit assignee
→ "Assign 5 tasks to john.doe?" → Confirm → Batch update
```

### 10.3 Edit History & Audit

**Track changes:**
- Who edited what when
- View edit history for a field
- Undo/redo functionality
- Audit log export

**UI:**
```
Press 'h' on edited field → Show history:
  2026-02-17 14:30  admin     42 → 100
  2026-02-17 09:15  john.doe  10 → 42
```

### 10.4 Advanced Validation

**Complex rules:**
```yaml
validation:
  depends_on: otherField
  conditional:
    - if: priority == "CRITICAL"
      then: timeout >= 3600
      error: "Critical tasks need timeout >= 1 hour"
```

**Cross-field validation:**
- Start date < End date
- Count <= MaxCount
- Conditional required fields

### 10.5 Rich Editors

**JSON editor:**
- Syntax highlighting
- Bracket matching
- Auto-formatting
- Validation with JSON Schema

**Date picker:**
- Calendar widget
- Keyboard navigation
- Relative dates (today, +7d, etc.)

**File upload:**
- Drag & drop files
- Base64 encoding for API
- File size validation

### 10.6 Undo/Redo

**Local undo:**
- Ctrl+Z in edit modal → restore original value
- Undo stack for multiple changes

**Server undo:**
- Ctrl+Z in table view → revert last save
- "Undo last change" action

### 10.7 Inline Editing (Alternative Mode)

**Quick edit without modal:**
- Double-click or F2 → edit in-place
- For simple text fields only
- Enter saves, Esc cancels

**When to use:**
- Quick edits (single character/word)
- Spreadsheet-like workflows
- Power users who prefer speed over safety

### 10.8 Mobile/Small Terminal Optimization

**Compact modal for 80x24:**
- Simplified layout
- Essential info only
- Smart wrapping

### 10.9 Keyboard Macro Support

**Record/replay edits:**
- Record: Ctrl+R → [series of edits] → Ctrl+R
- Replay: Ctrl+P
- Save macros to config

**Use case:** Repetitive bulk edits

### 10.10 API Rate Limiting & Caching

**Smart caching:**
- Cache content assist results (5 min TTL)
- Batch API calls when possible
- Debounce rapid searches

**Rate limiting:**
- Max N requests/sec to API
- Queue requests if limit exceeded
- Show "Loading..." indicator

---

## 11. Appendix

### 11.1 API Endpoints Used

**Variable editing:**
```
PUT /process-instance/{id}/variables/{varName}
Body: {"value": <typed-value>, "type": "<type>"}
```

**Task assignment:**
```
POST /task/{id}/assignee
Body: {"userId": "john.doe"}
```

**User fetching (content assist):**
```
GET /user?firstResult=0&maxResults=50
Optional filters: ?id=john* or ?firstNameLike=John%25
```

### 11.2 Configuration Examples

**Minimal config (backward compatible):**
```yaml
tables:
  - name: process-variables
    columns:
      - name: value
        editable: true        # Minimal - uses defaults
```

**Full config (all features):**
```yaml
custom_types:
  - name: email
    base_type: string
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    error_message: "Invalid email"

content_assist:
  user:
    enabled: true
    source: api
    endpoint: /user
    min_chars: 2

ui:
  edit_modal:
    buttons:
      save:
        background: "#00A8E1"
      cancel:
        background: "#666666"

tables:
  - name: task
    columns:
      - name: assignee
        editable: true
        input_type: user
        placeholder: "Unassigned"
```

### 11.3 Type Reference

**Complete type catalog:**

| Type | Base | Widget | Validation | Example Config |
|------|------|--------|------------|----------------|
| string | - | Text input | Optional pattern | `input_type: string` |
| integer | - | Number input | Integer parse | `input_type: integer` |
| long | integer | Number input | Int64 parse | `input_type: long` |
| double | - | Number input | Float parse | `input_type: double` |
| boolean | - | Toggle | true/false only | `input_type: boolean` |
| date | string | Date input | YYYY-MM-DD | `input_type: date` |
| json | string | Multiline | Valid JSON | `input_type: json` |
| enum | string | Dropdown | In enum list | `input_type: priority` + custom type |
| user | string | Content assist | Optional | `input_type: user` |
| email | string | Text input | Regex | Custom type with pattern |

### 11.4 Keyboard Reference Card

**Quick reference for users:**

```
╔══════════════════════════════════════════════════════════════╗
║                    o8n EDIT KEYBOARD SHORTCUTS               ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  TRIGGER EDIT                                                ║
║  ─────────────────────────────────────────────────────────   ║
║  e              Open edit modal for selected cell            ║
║                                                              ║
║  MODAL ACTIONS                                               ║
║  ─────────────────────────────────────────────────────────   ║
║  Enter          Save changes and close                       ║
║  Esc            Cancel edit (no save)                        ║
║  Tab            Next editable field                          ║
║  Shift+Tab      Previous editable field                      ║
║  ?              Show help                                    ║
║                                                              ║
║  TYPE-SPECIFIC                                               ║
║  ─────────────────────────────────────────────────────────   ║
║  Space          Toggle boolean (true ↔ false)                ║
║  ↑/↓ or j/k     Navigate dropdown suggestions                ║
║  Ctrl+U         Clear input field                            ║
║                                                              ║
║  TIPS                                                        ║
║  ─────────────────────────────────────────────────────────   ║
║  • Editable columns marked with * in header                  ║
║  • Type 2+ characters to see assignee suggestions            ║
║  • Validation errors prevent save (fix before Enter)         ║
║  • Press Tab to edit multiple fields before saving           ║
║                                                              ║
╚══════════════════════════════════════════════════════════════╝
```

### 11.5 Glossary

**Content Assist:** Auto-complete dropdown showing suggestions as user types.

**Enum Type:** Fixed set of valid values (dropdown selection).

**Custom Type:** User-defined type with validation rules in config.

**Editable Indicator:** Visual marker (icon/asterisk) showing a field can be edited.

**Overlay Modal:** Dialog centered on screen with dimmed background.

**Type Detection:** Automatic inference of data type from metadata or column name.

**Validation Rule:** Constraint that input must satisfy (e.g., min/max, pattern).

---

## Document Control

**Version History:**

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-02-17 | BMM UX Designer | Initial PRD created |

**Stakeholders:**
- Product Owner: Karsten Thoms
- UX Designer: BMM UX Designer
- Developer: TBD
- QA: TBD

**Review & Approval:**
- [ ] Product Owner approval
- [ ] Technical review
- [ ] UX review
- [ ] Security review (for API user fetching)

**Related Documents:**
- [Layout Design Optimized](layout-design-optimized.md)
- [o8n README](../../README.md)
- [o8n Configuration Schema](../../o8n-cfg.yaml)

---

**END OF PRD**
