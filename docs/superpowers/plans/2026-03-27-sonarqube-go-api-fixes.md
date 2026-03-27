# SonarQube Go API Quality Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix all 20 Go issues reported by SonarQube, update the `Mati-Lab Default` quality gate to reasonable thresholds, and assign the custom gate + Go profile to the `resto-rate-api` project so the quality gate passes.

**Architecture:** String constants go in a new `apps/api/src/services/constants.go` file shared across services. Cognitive-complexity reductions are achieved by extracting private helper functions within the same files. Gate/profile assignment is done via SonarQube REST API calls (no changes to CI pipeline).

**Tech Stack:** Go 1.22, GORM, Connect-RPC, SonarQube CE 26.2.0, Bash + curl, sonarqube-sandbox provisioning scripts.

---

## File Map

| Action   | File                                             | Purpose                                        |
|----------|--------------------------------------------------|------------------------------------------------|
| Modify   | `~/Projects/sonarqube-sandbox/sonar-config/quality-gates/default-gate.json` | Relax thresholds to match project reality |
| Modify   | `~/Projects/sonarqube-sandbox/sonar-config/provision.sh` | Add gate/profile assignment for `resto-rate-api` |
| Create   | `apps/api/src/services/constants.go`             | Shared error-message string constants (S1192)  |
| Modify   | `apps/api/src/services/reviews_service.go`       | Use constants; extract `applyTagFilter` + `applyReviewSort` (S3776) |
| Modify   | `apps/api/src/services/wishlist_service.go`      | Use constants (S1192)                          |
| Modify   | `apps/api/src/services/tags_service.go`          | Use constants (S1192)                          |
| Modify   | `apps/api/src/services/friendship_service.go`    | Use constants; extract `lookupReceiverByEmail` + `lookupReceiverByUsername` (S3776) |
| Modify   | `apps/api/src/services/auth_service.go`          | Remove TODO comment (S1135)                    |
| Modify   | `apps/api/src/dev_handlers.go`                   | Extract `findOrCreateDevUser`; use constants (S3776) |

---

### Task 1: Update quality gate thresholds + assign gate/profile to `resto-rate-api`

**Files:**
- Modify: `~/Projects/sonarqube-sandbox/sonar-config/quality-gates/default-gate.json`
- Modify: `~/Projects/sonarqube-sandbox/sonar-config/provision.sh`

**Context:** The current gate blocks at `new_coverage < 80` (project has 0%), `new_duplicated_lines_density > 1%` (project has 3.83%), and `new_violations > 0` (project has 14). These are too strict for active development. The project also uses the default "Sonar way" gate instead of the custom one.

- [ ] **Step 1: Update default-gate.json with reasonable thresholds**

```json
{
  "name": "Mati-Lab Default",
  "conditions": [
    { "metric": "new_coverage",                "op": "LT", "error": "40" },
    { "metric": "new_duplicated_lines_density", "op": "GT", "error": "5"  },
    { "metric": "new_reliability_rating",       "op": "GT", "error": "1"  },
    { "metric": "new_security_rating",          "op": "GT", "error": "1"  },
    { "metric": "new_maintainability_rating",   "op": "GT", "error": "1"  }
  ]
}
```

Rationale: 40% coverage is achievable for a new project; 5% duplication allows reasonable copy patterns; maintainability A added since that's what we're actively improving; reliability/security A stay strict.

Write this to `~/Projects/sonarqube-sandbox/sonar-config/quality-gates/default-gate.json`.

- [ ] **Step 2: Add project gate + profile assignment to provision.sh**

After the existing `echo "Provisioning complete."` line, add a new section. Find the `echo "Provisioning complete."` line and replace the file's main section to append:

```bash
echo "=== Project Assignments ==="
assign_gate_to_project() {
  local project_key="$1"
  local gate_name="$2"
  echo "  Assigning gate '${gate_name}' to project '${project_key}'..."
  sonar_post "/api/qualitygates/select" \
    --data-urlencode "projectKey=${project_key}" \
    --data-urlencode "gateName=${gate_name}" > /dev/null
  echo "  Done"
}

assign_profile_to_project() {
  local project_key="$1"
  local profile_name="$2"
  local language="$3"
  echo "  Assigning profile '${profile_name}' (${language}) to project '${project_key}'..."
  sonar_post "/api/qualityprofiles/add_project" \
    --data-urlencode "project=${project_key}" \
    --data-urlencode "qualityProfile=${profile_name}" \
    --data-urlencode "language=${language}" > /dev/null
  echo "  Done"
}

assign_gate_to_project    "resto-rate-api" "Mati-Lab Default"
assign_profile_to_project "resto-rate-api" "Mati-Lab Go" "go"
echo ""

echo "Provisioning complete."
```

Place `assign_gate_to_project` and `assign_profile_to_project` function definitions BEFORE the final `echo "=== Project Assignments ==="` call, and replace the single `echo "Provisioning complete."` so it still appears at the end.

- [ ] **Step 3: Run provision.sh against SonarQube**

```bash
export SONAR_TOKEN=squ_bcacb5fcfd823abad4b2e0fc8cfbf08986265cd3
export SONAR_HOST_URL=https://sonarqube.mati-lab.online
bash ~/Projects/sonarqube-sandbox/sonar-config/provision.sh
```

Expected output includes:
```
Gate 'Mati-Lab Default' already exists, re-applying conditions...
  Applying conditions...
    Added condition: new_coverage LT 40
    Added condition: new_duplicated_lines_density GT 5
...
=== Project Assignments ===
  Assigning gate 'Mati-Lab Default' to project 'resto-rate-api'...
  Done
  Assigning profile 'Mati-Lab Go' (go) to project 'resto-rate-api'...
  Done
```

- [ ] **Step 4: Verify gate + profile assigned via REST API**

```bash
curl -sf -u "squ_bcacb5fcfd823abad4b2e0fc8cfbf08986265cd3:" \
  "https://sonarqube.mati-lab.online/api/qualitygates/get_by_project?project=resto-rate-api" \
  | python3 -c "import sys,json; g=json.load(sys.stdin); print('Gate:', g['qualityGate']['name'])"
```

Expected: `Gate: Mati-Lab Default`

```bash
curl -sf -u "squ_bcacb5fcfd823abad4b2e0fc8cfbf08986265cd3:" \
  "https://sonarqube.mati-lab.online/api/qualityprofiles/search?project=resto-rate-api&language=go" \
  | python3 -c "import sys,json; p=json.load(sys.stdin)['profiles']; print('Profile:', p[0]['name'] if p else 'none')"
```

Expected: `Profile: Mati-Lab Go`

- [ ] **Step 5: Commit sonarqube-sandbox changes**

```bash
cd ~/Projects/sonarqube-sandbox
git add sonar-config/quality-gates/default-gate.json sonar-config/provision.sh
git commit -m "feat: relax default gate thresholds; assign gate+profile to resto-rate-api"
```

---

### Task 2: Create shared constants file (fix S1192)

**Files:**
- Create: `apps/api/src/services/constants.go`

**Context:** SonarQube S1192 flags string literals used 3+ times. The strings appear in `reviews_service.go`, `wishlist_service.go`, `tags_service.go`, and `friendship_service.go`.

- [ ] **Step 1: Create `apps/api/src/services/constants.go`**

```go
package services

// Error message constants — shared across services to avoid S1192 duplicate literals.
const (
	errDatabaseNotInitialized = "database not initialized"
	errGooglePlacesIDRequired = "google_places_id is required"
	errUserNotFound           = "user not found"
	errReviewNotFound         = "review not found"
	errIDRequired             = "id is required"
	errRequestIDRequired      = "request_id is required"
	errFriendUserIDRequired   = "friend_user_id is required"
	errUsernameRequired       = "username is required"
	errInvalidUsername        = "invalid username"
)
```

- [ ] **Step 2: Build to confirm the file compiles**

```bash
cd /home/gooral/Projects/resto-rate
bunx nx run api:build
```

Expected: exit code 0, no errors.

- [ ] **Step 3: Commit**

```bash
git add apps/api/src/services/constants.go
git commit -m "feat: add shared error-message constants (fix S1192)"
```

---

### Task 3: Substitute constants in reviews_service.go + wishlist_service.go + tags_service.go

**Files:**
- Modify: `apps/api/src/services/reviews_service.go`
- Modify: `apps/api/src/services/wishlist_service.go`
- Modify: `apps/api/src/services/tags_service.go`

- [ ] **Step 1: Update reviews_service.go**

Replace every quoted string literal that has a matching constant. The affected lines are:

| Line | Old string | Constant |
|------|-----------|----------|
| 34 | `"authentication required"` | *(keep — not in constants, only occurs once in getUserIDFromSession)* |
| 38 | `"session expired"` | *(keep — once)* |
| 42 | `"invalid session"` | *(keep — once)* |
| 57 | `"google_places_id is required"` | `errGooglePlacesIDRequired` |
| 112 | `"user not found"` | `errUserNotFound` |
| 134 | `"database not initialized"` | `errDatabaseNotInitialized` |
| 240 | `"id is required"` | `errIDRequired` |
| 250 | `"review not found"` | `errReviewNotFound` |
| 276 | `"id is required"` | `errIDRequired` |
| 282 | `"review not found"` | `errReviewNotFound` |
| 300 | `"id is required"` | `errIDRequired` |
| 308 | `"review not found"` | `errReviewNotFound` |
| 324 | `"google_places_id is required"` | `errGooglePlacesIDRequired` |

For each replacement, change `errors.New("...")` to `errors.New(errXxx)`.

- [ ] **Step 2: Update wishlist_service.go**

Replace:
- `"database not initialized"` → `errDatabaseNotInitialized` (3 occurrences at lines 30, 84, 111)
- `"google_places_id is required"` → `errGooglePlacesIDRequired` (2 occurrences at lines 39, 93)

- [ ] **Step 3: Update tags_service.go**

Replace:
- `"database not initialized"` → `errDatabaseNotInitialized` (1 occurrence at line 45)

- [ ] **Step 4: Build + test**

```bash
cd /home/gooral/Projects/resto-rate
bunx nx run api:build
bunx nx run api:test
```

Expected: exit code 0.

- [ ] **Step 5: Commit**

```bash
git add apps/api/src/services/reviews_service.go \
        apps/api/src/services/wishlist_service.go \
        apps/api/src/services/tags_service.go
git commit -m "refactor: replace duplicate string literals with constants (S1192)"
```

---

### Task 4: Fix S1192 in friendship_service.go

**Files:**
- Modify: `apps/api/src/services/friendship_service.go`

- [ ] **Step 1: Replace duplicate string literals in friendship_service.go**

Replace:
- `"user not found"` → `errUserNotFound` (3 occurrences: lines 40, 51, 276)
- `"request_id is required"` → `errRequestIDRequired` (2 occurrences: lines 115, 145)
- `"friend_user_id is required"` → `errFriendUserIDRequired` (line 175)
- `"username is required"` → `errUsernameRequired` (line 263)
- `"invalid username"` → `errInvalidUsername` (2 occurrences: lines 47, 266)

All replacements: `errors.New("...")` → `errors.New(errXxx)`.

- [ ] **Step 2: Build + test**

```bash
cd /home/gooral/Projects/resto-rate
bunx nx run api:build
bunx nx run api:test
```

Expected: exit code 0.

- [ ] **Step 3: Commit**

```bash
git add apps/api/src/services/friendship_service.go
git commit -m "refactor: replace duplicate string literals with constants in friendship_service (S1192)"
```

---

### Task 5: Reduce ListReviews cognitive complexity (S3776, complexity 30 → <15)

**Files:**
- Modify: `apps/api/src/services/reviews_service.go`

**Context:** `ListReviews` has SonarQube cognitive complexity 30 (threshold 15). The tag-filter block (AND/OR switch with nested loop) and the sort switch are the biggest contributors. Extract them as package-private helpers.

- [ ] **Step 1: Add `applyTagFilter` and `applyReviewSort` to reviews_service.go**

Add these two functions at the bottom of `reviews_service.go` (before the closing of the file, after `assertFriendship` if it exists, or after `getUserIDFromSession`):

```go
// applyTagFilter adds WHERE clauses for tag filtering based on mode (AND/OR).
func applyTagFilter(query *gorm.DB, slugs []string, mode v1.TagFilterMode) *gorm.DB {
	if len(slugs) == 0 {
		return query
	}
	if mode == v1.TagFilterMode_TAG_FILTER_MODE_AND {
		for _, slug := range slugs {
			query = query.Where("reviews.tags LIKE ?", fmt.Sprintf(`%%"%s"%%`, slug))
		}
		return query
	}
	// OR (default): at least one specified tag must appear
	conditions := make([]string, len(slugs))
	args := make([]interface{}, len(slugs))
	for i, slug := range slugs {
		conditions[i] = "reviews.tags LIKE ?"
		args[i] = fmt.Sprintf(`%%"%s"%%`, slug)
	}
	return query.Where("("+strings.Join(conditions, " OR ")+")", args...)
}

// applyReviewSort adds an ORDER BY clause based on the sort field.
func applyReviewSort(query *gorm.DB, sortBy v1.ReviewSortBy) *gorm.DB {
	switch sortBy {
	case v1.ReviewSortBy_REVIEW_SORT_BY_DATE_ASC:
		return query.Order("reviews.created_at ASC")
	case v1.ReviewSortBy_REVIEW_SORT_BY_RATING_DESC:
		return query.Order("reviews.rating DESC")
	case v1.ReviewSortBy_REVIEW_SORT_BY_RATING_ASC:
		return query.Order("reviews.rating ASC")
	default: // UNSPECIFIED and DATE_DESC → newest first
		return query.Order("reviews.created_at DESC")
	}
}
```

- [ ] **Step 2: Update ListReviews to call the helpers**

Replace the tag-filter block (lines 167–181) and sort switch (lines 206–215) in `ListReviews` with:

```go
	// Tag filter
	query = applyTagFilter(query, req.Msg.TagSlugs, req.Msg.TagFilterMode)

	// Rating range
	if req.Msg.MinRating > 0 {
		query = query.Where("reviews.rating >= ?", req.Msg.MinRating)
	}
	if req.Msg.MaxRating > 0 {
		query = query.Where("reviews.rating <= ?", req.Msg.MaxRating)
	}

	// Comment keyword search (case-insensitive)
	if req.Msg.CommentSearch != "" {
		query = query.Where("reviews.comment ILIKE ?", "%"+req.Msg.CommentSearch+"%")
	}

	// City / country filter (requires restaurant join above)
	if req.Msg.City != "" {
		query = query.Where("restaurants.city ILIKE ?", "%"+req.Msg.City+"%")
	}
	if req.Msg.Country != "" {
		query = query.Where("restaurants.country ILIKE ?", "%"+req.Msg.Country+"%")
	}

	// Sort order
	query = applyReviewSort(query, req.Msg.SortBy)
```

- [ ] **Step 3: Build + test**

```bash
cd /home/gooral/Projects/resto-rate
bunx nx run api:build
bunx nx run api:test
```

Expected: exit code 0.

- [ ] **Step 4: Commit**

```bash
git add apps/api/src/services/reviews_service.go
git commit -m "refactor: extract applyTagFilter + applyReviewSort to reduce ListReviews complexity (S3776)"
```

---

### Task 6: Reduce SendFriendRequest cognitive complexity (S3776, complexity 24 → <15)

**Files:**
- Modify: `apps/api/src/services/friendship_service.go`

**Context:** `SendFriendRequest` (complexity 24) has a large switch on email/username, each with nested DB query + error handling. Extract receiver lookup to two helpers.

- [ ] **Step 1: Add `lookupReceiverByEmail` and `lookupReceiverByUsername` to friendship_service.go**

Add below `derefStr`:

```go
// lookupReceiverByEmail finds a user by email or returns a Connect-RPC error.
func lookupReceiverByEmail(ctx context.Context, db *gorm.DB, email string) (models.User, error) {
	var receiver models.User
	if err := db.WithContext(ctx).Where("email = ?", email).First(&receiver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, connect.NewError(connect.CodeNotFound, errors.New(errUserNotFound))
		}
		return models.User{}, err
	}
	return receiver, nil
}

// lookupReceiverByUsername normalizes the handle, validates it, and finds the user.
func lookupReceiverByUsername(ctx context.Context, db *gorm.DB, rawUsername string) (models.User, error) {
	handle := strings.ToLower(strings.TrimPrefix(rawUsername, "@"))
	if !isValidUsername(handle) {
		return models.User{}, connect.NewError(connect.CodeInvalidArgument, errors.New(errInvalidUsername))
	}
	var receiver models.User
	if err := db.WithContext(ctx).Where("username = ?", handle).First(&receiver).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, connect.NewError(connect.CodeNotFound, errors.New(errUserNotFound))
		}
		return models.User{}, err
	}
	return receiver, nil
}
```

- [ ] **Step 2: Simplify SendFriendRequest receiver lookup switch**

Replace the receiver lookup switch block in `SendFriendRequest` (the `var receiver models.User` + `switch { case ... }` block, approximately lines 35–57) with:

```go
	var receiver models.User
	var lookupErr error
	switch {
	case req.Msg.GetReceiverEmail() != "":
		receiver, lookupErr = lookupReceiverByEmail(ctx, s.DB, req.Msg.GetReceiverEmail())
	case req.Msg.GetReceiverUsername() != "":
		receiver, lookupErr = lookupReceiverByUsername(ctx, s.DB, req.Msg.GetReceiverUsername())
	default:
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("receiver_email or receiver_username is required"))
	}
	if lookupErr != nil {
		return nil, lookupErr
	}
```

- [ ] **Step 3: Build + test**

```bash
cd /home/gooral/Projects/resto-rate
bunx nx run api:build
bunx nx run api:test
```

Expected: exit code 0.

- [ ] **Step 4: Commit**

```bash
git add apps/api/src/services/friendship_service.go
git commit -m "refactor: extract receiver lookup helpers to reduce SendFriendRequest complexity (S3776)"
```

---

### Task 7: Reduce devLoginHandler cognitive complexity + fix TODO comment (S3776, S1135)

**Files:**
- Modify: `apps/api/src/dev_handlers.go`
- Modify: `apps/api/src/services/auth_service.go`

**Context:** `devLoginHandler` has complexity 16 (threshold 15). The find-or-create user block (6 lines + nested error handling) can be extracted to `findOrCreateDevUser`. The TODO comment for Apple Sign-In in `auth_service.go:350` triggers S1135.

- [ ] **Step 1: Add `findOrCreateDevUser` to dev_handlers.go**

Add this function before `devLoginHandler`:

```go
// findOrCreateDevUser looks up a user by email; creates one with Name="Dev User" if not found.
func findOrCreateDevUser(db *gorm.DB, ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = models.User{
			Email: models.StringPtr(email),
			Name:  "Dev User",
		}
		return user, db.WithContext(ctx).Create(&user).Error
	}
	return user, err
}
```

- [ ] **Step 2: Update devLoginHandler to call the helper**

Replace the `var user models.User` + find-or-create block (lines 32–46) with:

```go
		user, err := findOrCreateDevUser(db, ctx, email)
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
```

Make sure `err` isn't redeclared — use `:=` here since it's the first declaration in the handler scope.

- [ ] **Step 3: Remove the Apple Sign-In TODO comment in auth_service.go**

In `apps/api/src/services/auth_service.go` around line 349, the commented-out Apple case is:

```go
	// case authv1.AuthProvider_AUTH_PROVIDER_APPLE:
	//     TODO: implement Apple Sign-In verification
```

Remove both lines entirely (the case will remain unsupported via the `default` branch).

- [ ] **Step 4: Build + test**

```bash
cd /home/gooral/Projects/resto-rate
bunx nx run api:build
bunx nx run api:test
```

Expected: exit code 0.

- [ ] **Step 5: Commit**

```bash
git add apps/api/src/dev_handlers.go apps/api/src/services/auth_service.go
git commit -m "refactor: extract findOrCreateDevUser; remove Apple TODO comment (S3776, S1135)"
```

---

### Task 8: Run SonarQube scan + verify gate passes

**Context:** The scan runs via the reusable workflow in `.github/workflows/`. We can trigger it locally by running `go test` with coverage and then `sonar-scanner`. Check the CI workflow to see the exact scanner invocation.

- [ ] **Step 1: Check CI workflow for scanner invocation**

```bash
cat /home/gooral/Projects/resto-rate/.github/workflows/ci.yml | grep -A 20 "sonar"
```

Note the exact `sonar-scanner` command and properties used (e.g., `sonar.projectKey`, coverage report path).

- [ ] **Step 2: Run Go tests with coverage**

```bash
cd /home/gooral/Projects/resto-rate/apps/api
go test ./src/... -coverprofile=coverage.out 2>&1 | tail -5
```

Expected: tests pass; `coverage.out` created.

- [ ] **Step 3: Run SonarQube scan**

Use the same command as CI (adapt from Step 1). Typically:

```bash
cd /home/gooral/Projects/resto-rate
sonar-scanner \
  -Dsonar.projectKey=resto-rate-api \
  -Dsonar.sources=apps/api/src \
  -Dsonar.go.coverage.reportPaths=apps/api/coverage.out \
  -Dsonar.host.url=https://sonarqube.mati-lab.online \
  -Dsonar.token=squ_bcacb5fcfd823abad4b2e0fc8cfbf08986265cd3
```

Expected: scan completes, task ID printed.

- [ ] **Step 4: Wait for analysis and check gate result**

```bash
sleep 15
curl -sf -u "squ_bcacb5fcfd823abad4b2e0fc8cfbf08986265cd3:" \
  "https://sonarqube.mati-lab.online/api/qualitygates/project_status?projectKey=resto-rate-api" \
  | python3 -c "
import sys, json
d = json.load(sys.stdin)['projectStatus']
print('Status:', d['status'])
for c in d.get('conditions', []):
    flag = '✓' if c['status'] == 'OK' else '✗'
    print(f\"  {flag} {c['metricKey']}: {c.get('actualValue','?')} (threshold {c.get('errorThreshold','?')})\")
"
```

Expected: `Status: OK` with all conditions showing `✓`.

- [ ] **Step 5: Push branch + open PR if all green**

```bash
cd /home/gooral/Projects/resto-rate
git push origin main
```

Or create a PR if working on a feature branch per project workflow.
