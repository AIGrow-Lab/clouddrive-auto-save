# Account Validation Failure Fix Plan

**Goal:** Ensure that the `last_check` timestamp is updated in the database even when account validation fails. Currently, it only sets the `status` to 0 on failure.

**Target File:** `internal/api/router.go`

**Changes:**
1. In `performAccountCheck`, when `err != nil` after calling `driver.GetInfo(ctx)`, modify the `db.DB.Model(account).Update` call to update both `status` and `last_check`.
2. Instead of `Update("status", 0)`, use `Updates` with a map to set `"status": 0` and `"last_check": time.Now()`.
3. Set `account.LastCheck = time.Now()` on the memory object as well so the caller gets the updated timestamp immediately.
