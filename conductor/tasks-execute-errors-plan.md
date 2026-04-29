# Deepen Tasks Execute E2E Tests Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enhance the task execution E2E tests (`execute.spec.ts`) by introducing various error scenarios (e.g., invalid share links, violation/banned files) based on real cloud drive API error codes, and verify the task logs/messages accurately reflect these errors.

**Architecture:**

1. **Mock HTTP Enhancement**:
   - Update Quark (`sharepage/token` or `detail`) to return specific error JSON based on the share URL ID (e.g., `mock_invalid` returns code `24001`, `mock_violation` returns code `41010`).
   - Update 139 (`getOutLinkInfoV6`) to return error code `200000727` if the `linkID` is `mock_invalid`.
   - Implement a simple way to verify "files are saved" by ensuring the task execution UI shows "SUCCESS" and maybe we can check the error messages via UI tooltips or error tags for failures.
2. **E2E Test Updates**:
   - Add a test for 139 with an invalid link, expecting a Fatal error (`[Fatal] 分享链接不存在或已被取消。`).
   - Add a test for Quark with a banned file link, expecting a Fatal error (`[Fatal] 该分享文件涉及违规内容，已被官方屏蔽。`).
   - Add a test for Quark with an invalid link (`[Fatal] 该分享已失效，可能已被取消或删除。`).

**Tech Stack:** Go (Mock HTTP Interceptor), TypeScript / Playwright (E2E Tests)

---

## Task 1: Update HTTP Mocks for Error Scenarios

**Files:**

- Modify: `internal/core/mock_http.go`

- [ ] **Step 1: Update the Quark and 139 mocks**

```go
// Replace the sharepage/token and getOutLinkInfoV6 handlers in internal/core/mock_http.go to support error simulation based on URL/Body.
```

## Task 2: Add Error Scenario E2E Tests

**Files:**

- Modify: `e2e/tests/tasks/execute.spec.ts`

- [ ] **Step 1: Add the failing tests**

```typescript
// Add new tests to e2e/tests/tasks/execute.spec.ts for 139 invalid link, Quark invalid link, and Quark violation link.
```

- [ ] **Step 2: Run test to verify it passes**

Run: `cd e2e && npx playwright test tests/tasks/execute.spec.ts`
Expected: PASS
