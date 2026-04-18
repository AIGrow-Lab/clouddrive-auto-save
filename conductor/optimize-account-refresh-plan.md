# Optimize Account Validation Refresh Plan

**Goal:** Avoid full page/list reloads (`fetchList()`) when a user clicks "Check" on a single account. Instead, update only that specific row's data in the table, even if the check fails.

**Target File:** `web/src/views/Accounts.vue`

**Context:** The backend `checkAccount` API already returns the updated account object in the JSON response payload on success (`{ id: 1, ... }`), and includes it in the error response payload on failure (`{ "error": "...", "account": { id: 1, status: 0, last_check: ... } }`).

**Changes:**
1. In `handleCheck(row)`, store the result of `await checkAccount(row.id)`.
2. On success, update the `row` object in-place using `Object.assign(row, updatedAccount)`.
3. On failure (in the `catch` block), check if `err.response.data.account` exists. If it does, update the `row` object in-place using `Object.assign(row, err.response.data.account)`.
4. Remove the `finally` block that calls `fetchList()`.

This will make the UI respond instantly and efficiently without unnecessary network requests for the entire list.
