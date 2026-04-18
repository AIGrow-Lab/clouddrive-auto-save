# Human-Friendly Error Mapping Plan

**Goal:** Improve backend API error messages for account validation, specifically replacing generic codes like "API error: 01000004" with human-readable descriptions.

**Target File:** `internal/core/cloud139/client.go`

**Changes:**
1. In `GetInfo` (where account info is fetched and validated), extract the `msg`, `message`, or `desc` fields from the raw API response.
2. Add specific mapping for common error codes (e.g., `01000004` to "登录凭证无效或已过期 (AuthToken / Cookie 错误)").
3. Include the original error code and extracted message in the fallback error text for easier debugging.
