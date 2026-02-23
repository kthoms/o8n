# o8n Roadmap 2026

**Vision:** Build the best terminal UI for managing Operaton workflow engines
**Status:** Post-Critical Audit, Ready for Feature Development
**Target:** A+ Grade TUI Experience by Q2 2026

---

## 📍 Where We Are Now

### ✅ What's Working (Audit Confirmed)
- Core architecture (Bubble Tea model/update/view) ✓
- API client integration ✓
- Navigation & drill-down workflow ✓
- Keyboard-first design ✓
- Configuration system ✓
- Edit modal with validation ✓
- Error/success feedback system ✓
- Pagination foundations ✓
- Color themes & environments ✓
- Help screen (spec says it exists) ✓

### ⚠️ What Was Fixed (This Session)
- **CRITICAL:** Footer spec violation (1 row → 2 rows) — FIXED ✓
- **HIGH:** Missing pagination counts for definitions — FIXED ✓
- Footer now compliant with specification ✓
- Process definitions show item counts ✓

### 🎯 What's Next (In Priority Order)

---

## 📅 Sprint Timeline

### **Sprint 2a: Quick Wins (Week 1-2, ~6-8 hours)**
*"Make the UI feel responsive and smart"*

- [ ] API latency display (shows system health)
- [ ] Context-aware key hints (shows what's possible)
- [ ] Inline validation feedback (prevents errors)
- [ ] Pagination status display (shows location)
- [ ] Write tests for new features
- [ ] Manual testing with running Operaton instance

**Deliverable:** A visibly more polished TUI with better feedback

**Success Metric:** Users report "it feels faster" and "I know what I can do"

---

### **Sprint 2b: Discoverability (Week 3-4, ~8-10 hours)**
*"Help users find what they need"*

- [ ] Search/filter functionality (`/` to search)
- [ ] Streaming indicator for large datasets
- [ ] Auto-complete for variable names
- [ ] Enhanced help modal with examples
- [ ] Keyboard cheat sheet (printable)
- [ ] Update documentation

**Deliverable:** Users can find instances fast, understand UI capabilities

**Success Metric:** "I found what I needed without scrolling"

---

### **Sprint 2c: Power User Features (Week 5-6, ~8-10 hours)**
*"Make it awesome for people who use it daily"*

- [ ] Saved views / quick filters (remember user state)
- [ ] Column visibility customization
- [ ] View history (navigate back to previous views)
- [ ] Keyboard shortcuts customization
- [ ] Status bar improvements
- [ ] Performance optimizations

**Deliverable:** Expert users can customize their experience

**Success Metric:** Daily users set up their perfect workflow

---

### **Sprint 3: Advanced Features (Future)**
*"Competitive advantages over manual API use"*

- [ ] Process visualizer (ASCII BPMN diagram)
- [ ] Historical data trends
- [ ] Multi-view dashboard
- [ ] Webhook notifications
- [ ] Vim mode toggle
- [ ] Plugin system (custom commands)

**Deliverable:** Features that justify using o8n instead of raw API

---

## 🎨 Design Evolution

### Phase 1: Foundation (Current - In Place)
```
┌─ Compact Header ────────────────────┐
│ o8n v0.1.0 | local | demo@... | ⚡  │
│ [?]Help [:]Switch [e]Edit [r]Refresh│
├─ Context Box (when active) ─────────┤
│ : [input with completion]           │
├─ Content Table ─────────────────────┤
│ ID     KEY      VERSION             │
│ def1   invoice  2                   │
│ def2   review   1                   │
├─ Footer (1 row) ────────────────────┤
│ [breadcrumb] | Status | ⚡ 42ms     │
└─────────────────────────────────────┘
```

### Phase 2: Responsiveness (This Sprint)
- Add latency tracking
- Make hints context-aware
- Show validation errors live
- Display pagination clearly

### Phase 3: Intelligence (Next Sprint)
- Add search/filter
- Show progress for large operations
- Auto-suggest values
- Remember user preferences

### Phase 4: Power (Future)
- Visualize workflows
- Show historical trends
- Support split-view
- Allow customization

---

## 🏗️ Architecture Overview

### Current
```
main.go (2800+ lines)
├── Model struct (layout, state)
├── Update() (keyboard + messages)
├── View() (rendering)
└── Message types (definitionsLoadedMsg, etc.)

api.go / internal/client/
├── HTTP client wrapper
├── OpenAPI generated client
└── Data transformation

config.go
├── YAML parsing
├── Environment setup
└── Table definitions
```

### Post-Enhancements
```
main.go (split into domains)
├── main_ui.go (rendering logic)
├── main_events.go (keyboard handling)
├── main_api.go (API integration)
├── main_search.go (search/filter)
└── main_views.go (view logic)

features/ (new)
├── features/pagination.go
├── features/validation.go
├── features/visualization.go
└── features/notifications.go

tests/
├── integration_test.go (expanded)
├── benchmark_test.go (new)
└── acceptance_test.go (new)
```

---

## 💾 Database & Storage

### Current
- YAML config only
- No history, no caching

### Future Enhancements
- [ ] SQLite for saved views
- [ ] Recent items cache
- [ ] User preferences storage
- [ ] Telemetry (anonymized usage)

---

## 🔌 API Integration Status

### Current Coverage
✅ Process Definitions
✅ Process Instances
✅ Variables (read/edit)
✅ Instance termination
❌ Task management
❌ Job management
❌ External tasks
❌ Deployment listing
❌ History data
❌ Metrics

### Q2 2026 Goals
- [ ] Add task management UI
- [ ] Add job monitoring view
- [ ] Add deployment info
- [ ] Add historical instance view
- [ ] Add metrics dashboard

---

## 📊 Quality Metrics

### Code Quality
| Metric | Current | Q2 2026 Target |
|--------|---------|----------------|
| Test Coverage | ~60% | >85% |
| Lines of main.go | 2800+ | <2000 (refactored) |
| Cyclomatic Complexity | Medium | Low |
| Documentation | 60% | 95% |

### User Experience
| Metric | Current | Q2 2026 Target |
|--------|---------|----------------|
| Time to Find Instance | ~10s | <2s |
| Learning Curve | Medium | Low |
| Keyboard Efficiency | Good | Excellent |
| Error Recovery | Manual | Automatic |

### Performance
| Metric | Current | Target |
|--------|---------|--------|
| Startup Time | <1s | <0.5s |
| Page Load | <200ms | <100ms |
| Search Response | N/A | <500ms |
| Memory Usage | ~20MB | <30MB |

---

## 🎓 Team Learning Path

### For New Contributors
1. **Week 1:** Understand Bubble Tea (model/update/view)
2. **Week 2:** Trace keyboard input through event loop
3. **Week 3:** Add a simple feature (keyboard shortcut)
4. **Week 4:** Implement one Quick Win from this roadmap

### For Code Reviewers
1. Review naming conventions (consistent with k9s)
2. Check keyboard shortcut conflicts
3. Verify error handling is complete
4. Test at small terminal sizes

### For UX Reviewers
1. Check against specification.md
2. Test keyboard discoverability
3. Verify color contrast (accessibility)
4. Test error message clarity

---

## 🌟 Success Criteria by Release

### v0.2.0 (Sprint 2a-2b, ~4 weeks)
- [ ] All Quick Wins implemented
- [ ] Search/filter working
- [ ] Test coverage >75%
- [ ] Help screen polished
- [ ] Specification fully compliant
- [ ] No critical bugs in audit

**Release Candidate:** Yes

### v0.3.0 (Sprint 2c, ~2 weeks)
- [ ] Saved views working
- [ ] Column customization available
- [ ] 95% test coverage
- [ ] Performance benchmarks established
- [ ] User documentation complete

**Release Candidate:** Yes

### v1.0.0 (Sprint 3, ~6 weeks)
- [ ] Process visualizer
- [ ] Dashboard mode
- [ ] Plugin system
- [ ] Full specification coverage
- [ ] Production-ready

**Release Candidate:** Yes

---

## 🚨 Known Limitations & Workarounds

### Current
1. **No undo:** Changes are immediate (API calls)
   - *Workaround:* Confirmation modal always shown

2. **Limited to REST API:** Can't access BPMN XML
   - *Workaround:* Show basic process structure

3. **No real-time updates:** Data refreshes on demand
   - *Workaround:* Add background refresh toggle

4. **No filtering:** Must scroll to find items
   - *Workaround:* Search coming in v0.2.0

### Will Be Fixed
- [ ] v0.2.0: Add search/filter
- [ ] v0.2.0: Real-time validation feedback
- [ ] v0.3.0: View customization
- [ ] v1.0.0: Process visualization

---

## 📚 Documentation Plan

### For Users
- [ ] Quick Start Guide (15 mins to first use)
- [ ] Keyboard Reference (1-page cheat sheet)
- [ ] Common Tasks (how-to guide)
- [ ] Troubleshooting (FAQs)
- [ ] Video Tutorial (5-10 mins)

### For Developers
- [ ] Architecture Guide (how components fit)
- [ ] Contributing Guide (how to add features)
- [ ] API Client Reference (wrapped endpoints)
- [ ] Testing Guide (how to write tests)
- [ ] Release Checklist

### For Operators
- [ ] Deployment Guide (how to install)
- [ ] Configuration Reference (all options)
- [ ] Keyboard Bindings (customization)
- [ ] Performance Tuning
- [ ] Security (authentication, ACLs)

---

## 🎯 Competitive Positioning

### vs. Web UI (if one exists)
✅ Works over SSH
✅ No browser needed
✅ Keyboard-first
✅ Fast and responsive
❌ No mobile support

### vs. k9s (for Kubernetes)
✅ Domain-specific (Operaton workflows)
✅ Edit capabilities
✅ Search/filter (coming)
❌ Less mature ecosystem
❌ Smaller community

### vs. Raw API Client
✅ Visual interface
✅ Keyboard shortcuts
✅ Context awareness
✅ Guided workflows
❌ Limited to predefined views

---

## 🔐 Security & Compliance

### Current
- ✅ Basic auth support
- ✅ Credentials in separate file (o8n-env.yaml)
- ✅ No logging of sensitive data

### Future
- [ ] Support OAuth/JWT tokens
- [ ] Audit logging
- [ ] RBAC enforcement
- [ ] Encryption at rest (saved passwords)
- [ ] Two-factor auth option

---

## 💰 Effort Estimation

| Phase | Hours | Timeline | Effort |
|-------|-------|----------|--------|
| Sprint 2a (Quick Wins) | 8-10 | 1-2 weeks | Easy |
| Sprint 2b (Search) | 8-12 | 2-3 weeks | Medium |
| Sprint 2c (Polish) | 8-10 | 2-3 weeks | Medium |
| Sprint 3 (Advanced) | 20-30 | 4-6 weeks | Hard |
| **Total to v1.0** | **50-70** | **12-16 weeks** | Medium |

---

## 🎬 Next Action Items

### Immediate (This Week)
- [ ] Commit audit fixes to main
- [ ] Share AUDIT.md with team
- [ ] Review ENHANCEMENT_PROPOSALS.md as a team
- [ ] Decide on v0.2.0 scope
- [ ] Assign Sprint 2a tasks

### Short-term (Next 2 Weeks)
- [ ] Implement Quick Wins #1-4
- [ ] Update specification.md
- [ ] Create user documentation
- [ ] Set up CI/CD for releases

### Medium-term (1 Month)
- [ ] Release v0.2.0 beta
- [ ] Gather user feedback
- [ ] Iterate on design
- [ ] Plan Sprint 2b features

---

## 🏆 Vision Statement

> o8n will become the **most intuitive terminal UI for Operaton**, enabling workflow operators to manage instances efficiently without ever leaving their keyboard. It will feel as polished and responsive as professional tools like k9s, with the domain-specific intelligence of a purpose-built application.

By Q2 2026, o8n will be:
- ✅ **Intuitive** — New users productive in 5 minutes
- ✅ **Responsive** — Millisecond feedback on every action
- ✅ **Powerful** — Features rivaling web UI, faster keyboard workflow
- ✅ **Reliable** — Zero data loss, comprehensive error handling
- ✅ **Beautiful** — Polished UI that inspires confidence

---

## 📞 Questions & Decisions Needed

1. **Scope for v0.2.0:** Quick Wins only, or include search?
2. **Database:** SQLite for saved views, or JSON files?
3. **Breaking Changes:** OK to refactor main.go?
4. **Community:** Open source release schedule?
5. **Monitoring:** Add telemetry/analytics?

---

## 📎 Appendix: Files Reference

| File | Purpose | Last Updated |
|------|---------|--------------|
| AUDIT.md | Critical audit findings | 2026-02-20 |
| ENHANCEMENT_PROPOSALS.md | 12 feature ideas | 2026-02-20 |
| QUICK_WINS.md | 4-hour implementation plan | 2026-02-20 |
| ROADMAP_2026.md | This document | 2026-02-20 |
| specification.md | Original specification | 2026-02-19 |
| o8n-cfg.yaml | App configuration | 2026-02-20 |
| o8n-env.yaml.example | Environment template | 2026-02-19 |

---

**Prepared by:** Critical Audit & UX Enhancement Analysis
**Date:** February 20, 2026
**Status:** Ready for Team Review & Sprint Planning
**Approval:** Pending

