# Fix Quark VIP Info Mock Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correct the mock response for Quark account VIP information to match the expected format so that the UI can properly display the member status in E2E tests, and add an E2E test assertion for it.

**Architecture:** Update the HTTP mock interceptor to return the backend's expected raw value (`SUPER_VIP` instead of `SVIP`) and extend the existing playwright test to verify the UI correctly displays the mapped value (`SVIP`).

**Tech Stack:** Go (for mock HTTP interceptor), TypeScript / Playwright (for E2E tests).

---

## Task 1: Fix Member Type in HTTP Mock

**Files:**

- Modify: `internal/core/mock_http.go:27-29`

- [ ] **Step 1: Write the minimal implementation**

```go
// Replace lines in internal/core/mock_http.go
 } else if strings.Contains(url, "pan.quark.cn/1/clouddrive/member") || strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/capacity") {
  respBody = `{"code": 0, "data": {"total_capacity": 1099511627776, "use_capacity": 549755813888, "member_type": "SUPER_VIP"}}`
 } else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/detail") {
```

## Task 2: Add VIP Assertion to Quark E2E Test

**Files:**

- Modify: `e2e/tests/accounts/quark.spec.ts:6-8`

- [ ] **Step 1: Write the failing test / implementation**

```typescript
// Replace lines in e2e/tests/accounts/quark.spec.ts
  test('成功绑定并展示夸克账号', async ({ page }) => {
    await addQuarkAccount(page);
    await expect(page.getByText('512').last()).toBeVisible();
    await expect(page.getByText('SVIP').last()).toBeVisible();
  });
```

- [ ] **Step 2: Run test to verify it passes**

Run: `cd e2e && npx playwright test tests/accounts/quark.spec.ts`
Expected: PASS
