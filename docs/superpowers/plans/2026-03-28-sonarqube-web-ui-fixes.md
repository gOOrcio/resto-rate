# SonarQube Web/UI Quality Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all 11 TypeScript/CSS/HTML issues in the `resto-rate-web` SonarQube project, assign the custom quality gate and TypeScript profile, and verify the gate passes.

**Architecture:** Three categories of fix: (1) SonarQube project config — create `sonar-project.properties` + assign gate/profile via provision.sh; (2) CSS false positives — inline `/* NOSONAR */` on valid Tailwind v4 at-rules; (3) TypeScript re-export pattern — convert `button/index.ts` from import+export to `export...from`. The `Web:S5254` false positive (Paraglide `lang` placeholder) is suppressed via `sonar.issue.ignore.multicriteria` in `sonar-project.properties`.

**Tech Stack:** SonarQube CE 26.2.0, TypeScript, Tailwind CSS v4, SvelteKit + Paraglide, sonarqube-sandbox provisioning scripts.

---

## File Map

| Action | File | Purpose |
|--------|------|---------|
| Create | `apps/web/sonar-project.properties` | Configure web scan: sources, exclusions, issue suppressions |
| Modify | `~/Projects/sonarqube-sandbox/sonar-config/provision.sh` | Add gate + profile assignment for `resto-rate-web` |
| Modify | `apps/web/src/app.css` | Add `/* NOSONAR */` to Tailwind v4 `@plugin` + `@custom-variant` lines |
| Modify | `apps/web/src/lib/components/ui/button/index.ts` | Convert named re-exports to `export { ... } from` (S7763) |

---

### Task 1: Assign gate + profile to `resto-rate-web` and create sonar-project.properties

**Files:**
- Modify: `~/Projects/sonarqube-sandbox/sonar-config/provision.sh`
- Create: `apps/web/sonar-project.properties`

**Context:** `resto-rate-web` currently uses the "Sonar way" gate (default). `provision.sh` already has `assign_gate_to_project` and `assign_profile_to_project` functions (added for the API). We only need to add two calls for the web project. The `sonar-project.properties` file is picked up automatically by `sonar-scanner` when run from `apps/web/` (matching the CI `projectBaseDir: apps/web`).

- [ ] **Step 1: Add web project assignment calls to provision.sh**

In `~/Projects/sonarqube-sandbox/sonar-config/provision.sh`, find the `=== Project Assignments ===` section (near the bottom) and add two lines after the existing `resto-rate-api` assignments:

```bash
echo "=== Project Assignments ==="
assign_gate_to_project    "resto-rate-api" "Mati-Lab Default"
assign_profile_to_project "resto-rate-api" "Mati-Lab Go" "go"
assign_gate_to_project    "resto-rate-web" "Mati-Lab Default"
assign_profile_to_project "resto-rate-web" "Mati-Lab TypeScript" "ts"
echo ""
```

- [ ] **Step 2: Run provision.sh**

```bash
export SONAR_TOKEN=<YOUR_SONARQUBE_TOKEN>
export SONAR_HOST_URL=https://sonarqube.mati-lab.online
bash ~/Projects/sonarqube-sandbox/sonar-config/provision.sh
```

Expected output includes:
```
=== Project Assignments ===
  Assigning gate 'Mati-Lab Default' to project 'resto-rate-api'...  Done
  Assigning profile 'Mati-Lab Go' (go) to project 'resto-rate-api'...  Done
  Assigning gate 'Mati-Lab Default' to project 'resto-rate-web'...  Done
  Assigning profile 'Mati-Lab TypeScript' (ts) to project 'resto-rate-web'...  Done
```

- [ ] **Step 3: Verify gate assigned**

```bash
curl -sf -u "$SONAR_TOKEN:" \
  "https://sonarqube.mati-lab.online/api/qualitygates/get_by_project?project=resto-rate-web" \
  | python3 -c "import sys,json; print('Gate:', json.load(sys.stdin)['qualityGate']['name'])"
```

Expected: `Gate: Mati-Lab Default`

- [ ] **Step 4: Create `apps/web/sonar-project.properties`**

```properties
sonar.projectKey=resto-rate-web
sonar.projectName=resto-rate (Web)
sonar.sources=src
sonar.exclusions=**/node_modules/**,**/*.test.ts,**/*.spec.ts
sonar.javascript.lcov.reportPaths=coverage/lcov.info

# Web:S5254 — lang="%paraglide.lang%" is the canonical Paraglide SvelteKit placeholder.
# SonarQube sees it as invalid BCP 47 and flags descendants; this is a false positive.
sonar.issue.ignore.multicriteria=e1
sonar.issue.ignore.multicriteria.e1.ruleKey=Web:S5254
sonar.issue.ignore.multicriteria.e1.resourceKey=**/app.html
```

- [ ] **Step 5: Commit sonarqube-sandbox changes**

```bash
cd ~/Projects/sonarqube-sandbox
git add sonar-config/provision.sh
git commit -m "feat: assign Mati-Lab Default gate + TypeScript profile to resto-rate-web"
```

- [ ] **Step 6: Commit sonar-project.properties**

```bash
cd /home/gooral/Projects/resto-rate
git add apps/web/sonar-project.properties
git commit -m "feat: add sonar-project.properties for web project (gate config + S5254 suppression)"
```

---

### Task 2: Fix css:S4662 — suppress valid Tailwind v4 at-rules in app.css

**Files:**
- Modify: `apps/web/src/app.css:5-8`

**Context:** SonarQube S4662 flags unknown CSS at-rules. Lines 5–6 use `@plugin` (Tailwind v4 plugin directive) and line 8 uses `@custom-variant` (Tailwind v4 custom variant directive). Both are valid Tailwind v4 CSS-first configuration syntax — SonarQube's CSS parser doesn't know about them. Fix: add `/* NOSONAR */` at the end of each flagged line.

Current `app.css` lines 1–8:
```css
@import "tailwindcss";

@import "tw-animate-css";

@plugin '@tailwindcss/forms';
@plugin '@tailwindcss/typography';

@custom-variant dark (&:is(.dark *));
```

- [ ] **Step 1: Add NOSONAR suppression comments**

The file after the fix (lines 1–8):

```css
@import "tailwindcss";

@import "tw-animate-css";

@plugin '@tailwindcss/forms'; /* NOSONAR */
@plugin '@tailwindcss/typography'; /* NOSONAR */

@custom-variant dark (&:is(.dark *)); /* NOSONAR */
```

- [ ] **Step 2: Build check**

```bash
cd /home/gooral/Projects/resto-rate
bunx nx run web:build 2>&1 | tail -5
```

Expected: exit 0, no errors.

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/app.css
git commit -m "fix: suppress SonarQube S4662 on valid Tailwind v4 at-rules (NOSONAR)"
```

---

### Task 3: Fix typescript:S7763 — convert button/index.ts to export...from

**Files:**
- Modify: `apps/web/src/lib/components/ui/button/index.ts`

**Context:** SonarQube S7763 flags 5 re-exports in `button/index.ts` that use the `import X from './file'; export { X }` pattern instead of the direct `export { X } from './file'` form. The file exports `Root` (default Svelte component) as both `Root` and `Button` — default exports can't be re-exported under two names in a single `export...from` statement, so `Root` must stay as an import. Everything else (`buttonVariants`, `ButtonProps`, `ButtonProps as Props`, `ButtonSize`, `ButtonVariant`) can move to `export...from`.

Current `button/index.ts`:
```typescript
import Root, {
    type ButtonProps,
    type ButtonSize,
    type ButtonVariant,
    buttonVariants,
} from "./button.svelte";

export {
    Root,
    type ButtonProps as Props,
    //
    Root as Button,
    buttonVariants,
    type ButtonProps,
    type ButtonSize,
    type ButtonVariant,
};
```

- [ ] **Step 1: Rewrite button/index.ts**

```typescript
import Root from "./button.svelte";

export { Root, Root as Button };
export {
    buttonVariants,
    type ButtonProps,
    type ButtonProps as Props,
    type ButtonSize,
    type ButtonVariant,
} from "./button.svelte";
```

Explanation:
- `import Root` is kept because `Root` must be aliased as both `Root` and `Button` — you cannot express `default as Root` and `default as Button` in a single `export...from` (same binding twice in one clause).
- All named exports (`buttonVariants`, types) move to `export { ... } from './button.svelte'`, removing the S7763 violations.

- [ ] **Step 2: svelte-check**

```bash
cd /home/gooral/Projects/resto-rate
bunx nx run web:check 2>&1 | tail -5
```

Expected: 0 errors, 0 warnings (or any pre-existing warnings unrelated to button).

- [ ] **Step 3: Commit**

```bash
git add apps/web/src/lib/components/ui/button/index.ts
git commit -m "fix: use export...from for named re-exports in button/index.ts (S7763)"
```

---

### Task 4: Run local SonarQube scan + verify gate passes + open PR

**Context:** The `sonar-scanner` is installed locally. The web `sonar-project.properties` is at `apps/web/` which is `projectBaseDir`. Run from `apps/web/` so the properties file is auto-discovered.

- [ ] **Step 1: Run vitest with coverage to produce lcov report**

```bash
cd /home/gooral/Projects/resto-rate/apps/web
bunx vitest run --coverage --coverage.reporter=lcov --passWithNoTests 2>&1 | tail -5
```

Expected: exit 0, creates `coverage/lcov.info` (may be empty if no tests — that's OK).

- [ ] **Step 2: Run sonar-scanner**

```bash
cd /home/gooral/Projects/resto-rate/apps/web
sonar-scanner \
  -Dsonar.host.url=https://sonarqube.mati-lab.online \
  -Dsonar.token=$SONAR_TOKEN \
  -Dsonar.branch.name=feat/sonarqube-web-fixes \
  2>&1 | tail -10
```

Expected: `EXECUTION SUCCESS`

- [ ] **Step 3: Check quality gate**

```bash
sleep 10
curl -sf -u "$SONAR_TOKEN:" \
  "https://sonarqube.mati-lab.online/api/qualitygates/project_status?projectKey=resto-rate-web&branch=feat%2Fsonarqube-web-fixes" \
  | python3 -c "
import sys, json
d = json.load(sys.stdin)['projectStatus']
print('Status:', d['status'])
for c in d.get('conditions', []):
    flag = '✓' if c['status'] == 'OK' else '✗'
    print(f\"  {flag} {c['metricKey']}: {c.get('actualValue','?')} (threshold {c.get('errorThreshold','?')})\")
"
```

Expected: `Status: OK`

- [ ] **Step 4: Push branch and open PR**

```bash
cd /home/gooral/Projects/resto-rate
git push -u origin feat/sonarqube-web-fixes
gh pr create \
  --title "fix: SonarQube web/UI quality fixes (S7763, S4662, S5254)" \
  --body "$(cat <<'EOF'
## Summary

- Assigns \`Mati-Lab Default\` quality gate + \`Mati-Lab TypeScript\` profile to \`resto-rate-web\`
- Creates \`apps/web/sonar-project.properties\` with scan config and \`Web:S5254\` suppression for Paraglide lang placeholder
- **S4662 (3×):** Adds \`/* NOSONAR */\` to valid Tailwind v4 \`@plugin\` and \`@custom-variant\` directives in \`app.css\`
- **S7763 (5×):** Converts \`button/index.ts\` named re-exports to \`export { ... } from\` pattern
- **S5254 (3×):** Suppressed via \`sonar.issue.ignore.multicriteria\` — \`%paraglide.lang%\` is the canonical Paraglide SvelteKit placeholder, not a real BCP 47 tag

## Test plan

- [ ] \`bunx nx run web:build\` passes
- [ ] \`bunx nx run web:check\` passes
- [ ] SonarQube CI scan shows Quality Gate: OK

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```
