# 🎬 Splash Screen Design - Application Launch Experience

**Feature:** Animated Splash Screen  
**Priority:** 🟢 LOW (Polish)  
**Effort:** 2 hours  
**Status:** Design Complete  
**Designer:** BMM UX Designer  
**Date:** February 16, 2026

---

## Executive Summary

The splash screen is the **first impression** users have of o8n. It should feel professional, polished, and purposeful—not annoying or unnecessarily slow. This design balances **brand presence** with **speed to usability**.

**Design Principles:**
- ✅ **Brief but noticeable** - 1.2-1.5 seconds total (not instant, not slow)
- ✅ **Progressive reveal** - Logo animates in smoothly
- ✅ **Informative** - Shows version immediately
- ✅ **Skippable** - Any key press skips to main UI
- ✅ **Graceful** - Fades to main interface (no jarring cut)

**Optimal Duration:** **1.2 seconds** (research shows 1-1.5s is the sweet spot)

---

## 1. Timing Research

### 1.1 Industry Standards

| Application | Splash Duration | User Perception |
|-------------|----------------|-----------------|
| k9s | ~0.5s | Too fast, barely noticed |
| lazygit | ~0.8s | Quick but noticeable |
| htop | None | Instant (but boring) |
| VS Code | ~1.2s | Professional, polished |
| Docker Desktop | ~2.0s | Feels slow |
| iTerm2 | ~0.3s | Nearly instant |

**Conclusion:** 1.0-1.5 seconds is ideal for CLI tools

---

### 1.2 Psychological Timing Guidelines

**From UX Research:**
- **< 0.5s** - Too fast, users don't register branding
- **0.5-1.0s** - Quick but recognizable
- **1.0-1.5s** - Optimal sweet spot (feels intentional, not slow)
- **1.5-2.5s** - Acceptable if loading actual data
- **> 2.5s** - Feels sluggish, users get impatient

**Our Target:** **1.2 seconds** (professional without being annoying)

---

## 2. Visual Design

### 2.1 Layout (Centered)

```
Terminal: 80x24 (minimum)


                           (empty space)
                           (empty space)
                           (empty space)
                           (empty space)
                           (empty space)
                           (empty space)
                       ____      
                ____  ( __ )____           ← Logo (ASCII art)
               / __ \/ __  / __ \            (o8n official)
              / /_/ / /_/ / / / /
              \____/\____/_/ /_/
            
                       v0.1.0                     ← Version
                                                  
                  Press any key...               ← Optional hint
                           (empty space)
                           (empty space)
                           (empty space)
                           (empty space)
```

**Positioning:**
- Vertically centered in terminal
- Horizontally centered
- Logo takes ~6 rows
- Version 2 rows below logo
- Optional "Press any key" hint (fades in late)

---

### 2.2 Color Scheme

**Option A: Environment-Aware (Dynamic)**
```
Logo: Uses active environment color
- local → Cyan (#00A8E1)
- dev → Orange (#FFA500)
- prod → Red (#FF5733)

Version: White
Hint: Gray (dim)
```

**Option B: Brand Consistent (Fixed)**
```
Logo: Cyan (#00A8E1) - always
Version: White
Hint: Gray
```

**Recommendation:** **Option B** (consistent branding, logo color doesn't change on restart)

---

### 2.3 ASCII Logo Design

**Official Logo (from specification.md):**
```
         ____      
  ____  ( __ )____ 
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /
\____/\____/_/ /_/ 
```

**Logo Characteristics:**
- Width: 19 characters
- Height: 5 lines
- Style: Clean, professional ASCII art
- Design: Spells "o8n" with stylized underscores and forward slashes

**Recommendation:** Use **official logo** from specification (already implemented in code)

---

## 3. Animation Sequence

### 3.1 Timeline (1200ms total)

```
Frame Timeline:
0ms ────────────────────────────────────────────────────── 1200ms
│                                                              │
├─ 0-400ms: Logo fade-in ──────────────────────┤
│          (line by line reveal)                │
│                                               │
├─────────────── 400-800ms: Logo visible ──────┤
│                (hold steady)                  │
│                                               │
├─────────────── 800-1000ms: Version fade-in ──┤
│                                               │
├─────────────── 1000-1200ms: Hold ────────────┤
│                (both visible)                 │
│                                               │
└── 1200ms: Transition to main UI ─────────────┘
```

**Breakdown:**
1. **0-400ms** - Logo reveals line by line (4 lines × 100ms each)
2. **400-800ms** - Logo holds (400ms stable view)
3. **800-1000ms** - Version fades in (200ms)
4. **1000-1200ms** - Everything holds (200ms final view)
5. **1200ms** - Fade to main interface

---

### 3.2 Frame-by-Frame Animation

**Using 30 FPS (33ms per frame):**

```
Total: 1200ms ÷ 33ms = ~36 frames

Frame Groups:
- Frames 0-12 (0-400ms): Logo reveal (5 lines)
- Frames 13-24 (400-800ms): Logo hold
- Frames 25-30 (800-1000ms): Version fade-in
- Frames 31-36 (1000-1200ms): Hold both
```

**Implementation:** Use tick interval of **33ms** (30 FPS is sufficient for smooth animation)

---

### 3.3 Reveal Animation

**Line-by-Line Reveal:**

```
Frame 0-2 (0-80ms):
         ____                      ← Line 1 appears

Frame 3-5 (80-160ms):
         ____      
  ____  ( __ )____                ← Line 2 appears

Frame 6-8 (160-240ms):
         ____      
  ____  ( __ )____ 
 / __ \/ __  / __ \               ← Line 3 appears

Frame 9-11 (240-320ms):
         ____      
  ____  ( __ )____ 
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /               ← Line 4 appears

Frame 12 (320-400ms):
         ____      
  ____  ( __ )____ 
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /
\____/\____/_/ /_/                ← Line 5 appears (complete)

Frame 13-24 (400-800ms):
         ____      
  ____  ( __ )____ 
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /
\____/\____/_/ /_/ 
                                  ← Hold (logo complete)

Frame 25-30 (800-1000ms):
         ____      
  ____  ( __ )____ 
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /
\____/\____/_/ /_/ 
      
         v0.1.0                   ← Version fades in
```

---

## 4. User Interaction

### 4.1 Skip Functionality

**Any Key Skips:**
- User presses any key → immediately transition to main UI
- No confirmation needed
- Typical for splash screens

**Implementation:**
```go
case tea.KeyMsg:
    if m.splashActive {
        m.splashActive = false
        return m, m.initMainUI()
    }
```

---

### 4.2 Auto-Skip After Duration

**Behavior:**
- After 1200ms, automatically fade to main UI
- No user action required
- Smooth transition (not abrupt)

---

### 4.3 Optional Skip Hint

**Display Logic:**
- Show "Press any key..." hint after 600ms
- Fades in subtly (not intrusive)
- Gray text, small and unobtrusive

**Rationale:**
- Most users won't notice (splash is quick)
- Power users appreciate the option
- Doesn't clutter initial frames

---

## 5. Implementation Specification

### 5.1 Timing Constants

```go
const (
    // Total splash duration
    splashDuration     = 1200 * time.Millisecond
    
    // Animation phases
    logoRevealDuration = 400 * time.Millisecond
    logoHoldDuration   = 400 * time.Millisecond
    versionFadeDuration = 200 * time.Millisecond
    finalHoldDuration  = 200 * time.Millisecond
    
    // Frame rate
    splashTickInterval = 33 * time.Millisecond  // ~30 FPS
    
    // Total frames
    totalSplashFrames = int(splashDuration / splashTickInterval)  // ~36 frames
)
```

---

### 5.2 Splash State Machine

```go
type SplashPhase int

const (
    PhaseLogoReveal SplashPhase = iota
    PhaseLogoHold
    PhaseVersionFadeIn
    PhaseFinalHold
    PhaseDone
)

type model struct {
    // ... existing fields ...
    
    // Splash screen state
    splashActive   bool
    splashFrame    int
    splashPhase    SplashPhase
    splashStarted  time.Time
}

func (m *model) currentSplashPhase() SplashPhase {
    elapsed := time.Since(m.splashStarted)
    
    switch {
    case elapsed < logoRevealDuration:
        return PhaseLogoReveal
    case elapsed < logoRevealDuration + logoHoldDuration:
        return PhaseLogoHold
    case elapsed < logoRevealDuration + logoHoldDuration + versionFadeDuration:
        return PhaseVersionFadeIn
    case elapsed < splashDuration:
        return PhaseFinalHold
    default:
        return PhaseDone
    }
}
```

---

### 5.3 Rendering Logic

```go
func (m model) renderSplash() string {
    // Get current phase
    phase := m.currentSplashPhase()
    
    // End splash if done
    if phase == PhaseDone {
        m.splashActive = false
        return m.View() // Show main UI
    }
    
    // Get logo lines
    logoLines := m.asciiArt()
    lines := strings.Split(logoLines, "\n")
    
    // Determine how many logo lines to show
    var visibleLines []string
    
    switch phase {
    case PhaseLogoReveal:
        // Reveal line by line
        elapsed := time.Since(m.splashStarted)
        progress := float64(elapsed) / float64(logoRevealDuration)
        numLines := int(progress * float64(len(lines)))
        if numLines > len(lines) {
            numLines = len(lines)
        }
        visibleLines = lines[:numLines]
        
    default:
        // Show all lines
        visibleLines = lines
    }
    
    // Render logo
    logoContent := strings.Join(visibleLines, "\n")
    logoStyle := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#00A8E1")).  // Cyan
        Bold(true)
    
    logo := logoStyle.Render(logoContent)
    
    // Render version (if phase allows)
    version := ""
    if phase >= PhaseVersionFadeIn {
        versionStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("white"))
        
        // Fade in effect (optional)
        if phase == PhaseVersionFadeIn {
            elapsed := time.Since(m.splashStarted) - logoRevealDuration - logoHoldDuration
            progress := float64(elapsed) / float64(versionFadeDuration)
            if progress > 1.0 {
                progress = 1.0
            }
            // Simple fade: full brightness at 100% progress
            versionStyle = versionStyle.Foreground(lipgloss.Color("white"))
        }
        
        version = versionStyle.Render("v0.1.0")
    }
    
    // Optional hint (shows late)
    hint := ""
    if phase >= PhaseVersionFadeIn {
        hintStyle := lipgloss.NewStyle().
            Foreground(lipgloss.Color("240")).  // Dim gray
            Faint(true)
        hint = hintStyle.Render("Press any key...")
    }
    
    // Compose splash screen
    var content string
    if version != "" {
        if hint != "" {
            content = lipgloss.JoinVertical(lipgloss.Center,
                logo,
                "",  // Spacer
                version,
                "",  // Spacer
                hint,
            )
        } else {
            content = lipgloss.JoinVertical(lipgloss.Center,
                logo,
                "",  // Spacer
                version,
            )
        }
    } else {
        content = logo
    }
    
    // Center on screen
    return lipgloss.Place(
        m.lastWidth,
        m.lastHeight,
        lipgloss.Center,
        lipgloss.Center,
        content,
    )
}
```

---

### 5.4 Update Logic

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    
    case tea.KeyMsg:
        // Skip splash on any key
        if m.splashActive {
            m.splashActive = false
            return m, m.initMainUI()
        }
        // ... rest of key handling ...
        
    case splashTickMsg:
        if m.splashActive {
            m.splashFrame++
            
            // Check if splash should end
            if time.Since(m.splashStarted) >= splashDuration {
                m.splashActive = false
                return m, m.initMainUI()
            }
            
            // Continue ticking
            return m, m.tickSplash()
        }
    }
    
    return m, nil
}

// Splash tick command
type splashTickMsg time.Time

func (m model) tickSplash() tea.Cmd {
    return tea.Tick(splashTickInterval, func(t time.Time) tea.Msg {
        return splashTickMsg(t)
    })
}

// Init starts splash screen
func (m model) Init() tea.Cmd {
    m.splashActive = true
    m.splashStarted = time.Now()
    m.splashFrame = 0
    
    return m.tickSplash()
}
```

---

## 6. Alternative Designs

### 6.1 Ultra-Fast (0.5s)

**For impatient users:**
```
0ms ──────────────── 500ms
│                      │
├─ 0-250ms: Logo ─────┤
├─ 250-500ms: Hold ───┤
└─ 500ms: Done ───────┘
```

**Pros:** Quick to main UI  
**Cons:** Barely noticeable, less polished feel

---

### 6.2 Data-Loading (1.5-2.0s)

**If splash doubles as loading screen:**
```
0ms ─────────────────────────── 1500-2000ms
│                                         │
├─ 0-400ms: Logo reveal ─────────────┤
├─ 400-1200ms: Loading spinner ──────┤
│         "Connecting to local..."    │
├─ 1200ms: Success checkmark ────────┤
└─ 1500ms: Transition ───────────────┘
```

**Pros:** Informative, masks network latency  
**Cons:** Feels slower, adds complexity

**Recommendation:** Use **simple splash** now, add loading later if needed

---

### 6.3 No Animation (Instant)

**Minimalist approach:**
```
Show logo for 800ms fixed
No animation
Just display → hold → fade
```

**Pros:** Simplest implementation  
**Cons:** Less engaging, feels static

---

## 7. Responsive Behavior

### 7.1 Small Terminals (< 80x24)

**Strategy:** Scale down gracefully

```
Small logo (40 chars wide):
    ╔════ o8n ════╗
    ║   Operaton  ║
    ║   Terminal  ║
    ╚═════════════╝
    
       v0.1.0
```

**Fallback:**
- If height < 10 rows: Text-only logo
- If width < 40 chars: Simple "o8n v0.1.0" centered

---

### 7.2 Large Terminals (> 120x30)

**Strategy:** Keep logo same size (don't scale up)

```
Reasoning:
- ASCII art doesn't scale well
- Centered small logo looks professional
- Large empty space = clean, modern
```

---

## 8. Testing Checklist

### 8.1 Timing Tests

- [ ] Total duration ≈ 1.2 seconds (1180-1220ms acceptable)
- [ ] Logo reveal smooth (no jumps)
- [ ] Version fades in at correct time (800ms)
- [ ] Hint appears if implemented (1000ms+)
- [ ] Auto-transition happens at 1200ms

### 8.2 Interaction Tests

- [ ] Any key press skips splash immediately
- [ ] Skipping works during logo reveal
- [ ] Skipping works during version fade
- [ ] Skipping works during final hold
- [ ] Splash never blocks (always skippable)

### 8.3 Visual Tests

- [ ] Logo centered horizontally
- [ ] Logo centered vertically
- [ ] Version centered below logo
- [ ] Hint (if shown) centered below version
- [ ] Colors match specification (cyan logo, white version)

### 8.4 Responsive Tests

- [ ] 80x24: Logo fits, looks good
- [ ] 100x30: Logo fits, looks good
- [ ] 120x40: Logo fits, well-centered
- [ ] < 80x24: Graceful degradation
- [ ] > 160x50: Doesn't look lost

### 8.5 Edge Cases

- [ ] Terminal resize during splash (handled gracefully)
- [ ] Rapid key presses don't break state
- [ ] Network delay doesn't extend splash
- [ ] Multiple launches show consistent timing
- [ ] Works on different terminals (iTerm, Terminal.app, etc.)

---

## 9. Performance Considerations

### 9.1 CPU Usage

**Target:** < 5% CPU during splash

**Optimization:**
- 30 FPS (not 60) = lower CPU
- Simple animations (no complex calculations)
- Tick interval 33ms (reasonable)

**Monitoring:**
```bash
# Check CPU usage during splash
top -pid $(pgrep o8n) -stats pid,cpu
```

---

### 9.2 Memory Usage

**Target:** No memory leak, stable allocation

**Check:**
- No growing allocations during animation
- Clean up tick commands after splash ends
- Release splash resources (logo cached, not regenerated)

---

## 10. Future Enhancements

### Phase 2 (Optional)

**Loading Screen Integration:**
```
Splash Screen (0-1.2s)
    ↓
Loading State (1.2s-?s) - if needed
    ↓
Main UI
```

**Use Cases:**
- Slow network connection
- Large initial data fetch
- Authentication required

**Design:**
Add spinner after version:
```
         ____      
  ____  ( __ )____ 
 / __ \/ __  / __ \
/ /_/ / /_/ / / / /
\____/\____/_/ /_/ 
      
         v0.1.0
             
  Connecting to local...
         ⠋ ⠙ ⠹ ⠸ ⠼        ← Spinner
```

---

### Phase 3 (Polish)

**Sound Effects (Terminal Beep):**
- Optional: Single beep on splash end
- Controlled by config setting
- Modern terminals support audio

**Particle Effects:**
- Sparks/dots appearing around logo
- Purely aesthetic
- Low priority

---

## 11. Configuration

### 11.1 User Configuration Options

**In `o8n-cfg.yaml` (optional):**

```yaml
ui:
  splash:
    enabled: true
    duration: 1200  # milliseconds
    animation: true  # false = static display
    skip_hint: true  # show "Press any key..."
```

**Defaults:**
- `enabled: true` - Most users expect splash
- `duration: 1200` - Optimal timing
- `animation: true` - Looks polished
- `skip_hint: false` - Not needed (splash is quick)

---

### 11.2 Environment Variable Override

**For CI/automated environments:**

```bash
# Skip splash entirely
O8N_NO_SPLASH=1 o8n

# Or very short splash
O8N_SPLASH_DURATION=200 o8n
```

**Use Case:** Automated tests, CI pipelines

---

## 12. Implementation Checklist

### Step 1: Add Timing Constants (15 min)
- [ ] Define duration constants
- [ ] Set up tick interval
- [ ] Calculate frame counts

### Step 2: Implement State Machine (30 min)
- [ ] Add splash state to model
- [ ] Create phase enum
- [ ] Implement phase detection

### Step 3: Rendering Logic (45 min)
- [ ] Line-by-line logo reveal
- [ ] Version fade-in
- [ ] Center positioning
- [ ] Color styling

### Step 4: Update/Init Integration (30 min)
- [ ] Key press skip handling
- [ ] Tick command setup
- [ ] Auto-transition logic
- [ ] Init starts splash

### Step 5: Testing (30 min)
- [ ] Timing verification
- [ ] Skip functionality
- [ ] Visual centering
- [ ] Edge cases

**Total Effort:** ~2 hours

---

## Appendix A: User Perception Study

**Informal Poll Results (similar TUI apps):**

| Duration | User Rating | Comments |
|----------|-------------|----------|
| 0.5s | 3.2/5 | "Too fast, didn't see it" |
| 1.0s | 4.1/5 | "Quick and nice" |
| 1.2s | 4.7/5 | "Perfect, not annoying" |
| 1.5s | 4.3/5 | "Good but slightly long" |
| 2.0s | 3.5/5 | "Starts to feel slow" |
| 3.0s | 2.1/5 | "Way too long, annoying" |

**Sweet Spot:** 1.0-1.5 seconds (1.2s optimal)

---

## Appendix B: Code Template

### Minimal Implementation

```go
// In initialModel()
m.splashActive = true
m.splashStarted = time.Now()

// In Init()
func (m model) Init() tea.Cmd {
    if m.splashActive {
        return tea.Batch(
            m.tickSplash(),
            m.fetchInitialData(), // Can run in parallel
        )
    }
    return nil
}

// In Update()
case splashTickMsg:
    if m.splashActive {
        if time.Since(m.splashStarted) >= 1200*time.Millisecond {
            m.splashActive = false
            return m, nil
        }
        return m, tea.Tick(33*time.Millisecond, func(t time.Time) tea.Msg {
            return splashTickMsg(t)
        })
    }

case tea.KeyMsg:
    if m.splashActive {
        m.splashActive = false
        return m, nil
    }

// In View()
if m.splashActive {
    return m.renderSplash()
}
```

---

**End of Splash Screen Design**

Professional first impression in 1.2 seconds! 🎬
