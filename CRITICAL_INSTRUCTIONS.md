# ⚠️ CRITICAL INSTRUCTIONS - DO NOT IGNORE

## 🚨 CRITICAL CONFIG FILES - NEVER DELETE OR EMPTY

These files MUST be preserved at all times:

### 1. **o8n-cfg.yaml** (760 lines)
   - Contains all table definitions, column specs, drilldown rules
   - **DO NOT DELETE**
   - **DO NOT EMPTY**
   - **RESTORE from git if corrupted:**
     ```bash
     git show HEAD~2:o8n-cfg.yaml > o8n-cfg.yaml
     ```

### 2. **o8n-env.yaml** (NOT in git - local credentials)
   - Contains environment URLs, usernames, passwords
   - **DO NOT DELETE**
   - **PRESERVE across commits**
   - **IF ACCIDENTALLY DELETED, restore from:**
     ```yaml
     environments:
       local:
         url: "http://localhost:8080/engine-rest"
         username: "demo"
         password: "demo"
         ui_color: "#00A8E1"
     active: local
     ```

### 3. **o8n-env.yaml.example** (reference only)
   - Safe to view but don't use for actual config

---

## ✅ BEFORE EACH COMMIT - CHECKLIST

- [ ] Verify o8n-cfg.yaml exists and has 760+ lines
- [ ] Verify o8n-env.yaml exists and has ~10 lines
- [ ] Run: `wc -l o8n-*.yaml`
- [ ] Build and test: `make build && go test ./...`
- [ ] Never use `git add .` blindly - always review changes

---

## 🔒 How These Files Get Deleted

**DANGER POINTS:**
1. ❌ Test cleanup code that deletes all YAML files
2. ❌ Overly aggressive `.gitignore` patterns
3. ❌ Build scripts that reset config
4. ❌ Using `git clean -fd` (forcefully deletes untracked files)

---

## 📝 ADDED TO CLAUDE.MD

The following instruction has been added to CLAUDE.md:

```
⚠️ CRITICAL: Never delete or empty o8n-cfg.yaml or o8n-env.yaml
- These are application configuration files required for runtime
- o8n-cfg.yaml: Table definitions (760 lines) - Always preserve
- o8n-env.yaml: Local credentials (gitignored) - Always restore if lost
- Before each commit, verify: wc -l o8n-*.yaml
- If deleted: git show HEAD~2:o8n-cfg.yaml > o8n-cfg.yaml
```

---

## 📞 ACTION ITEMS FOR FUTURE SESSIONS

1. **Start of session:** Check file status
   ```bash
   wc -l o8n-*.yaml
   # Expected: ~760 for cfg, ~11 for env
   ```

2. **Before commit:** Verify files intact
   ```bash
   git diff o8n-*.yaml
   # Should show NO deletions
   ```

3. **Never run:**
   - ❌ `git clean -fd`
   - ❌ Build commands that reset config
   - ❌ Test cleanup that deletes YAML

---

**Status:** ✅ Config files restored (760 + 11 lines)
**Last Restored:** February 20, 2026
