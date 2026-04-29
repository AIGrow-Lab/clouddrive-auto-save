# Quark E2E Extended Coverage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand the E2E test coverage for Quark accounts to include various member types (Normal, Over-capacity) by making the HTTP mock dynamic based on cookie parameters.

**Architecture:**

1. Parameterize `addQuarkAccount` in `account.fixture.ts` to accept a custom cookie string and username prefix.
2. Modify `mock_http.go` to parse the incoming Cookie header (or `kps` for the app route) to return dynamic member and capacity JSON payloads.
3. Add new test cases to `quark.spec.ts` for Normal users and Over-capacity users, asserting the correct display names and sizes.

**Tech Stack:** Go (HTTP Mock), TypeScript / Playwright (E2E Tests)

---

## Task 1: Refactor Quark E2E Fixture

**Files:**

- Modify: `e2e/fixtures/account.fixture.ts`

- [ ] **Step 1: Write the minimal implementation**

```typescript
// Replace addQuarkAccount in e2e/fixtures/account.fixture.ts
export async function addQuarkAccount(page: Page, cookieStr: string = '__uid=mock; mock_cookie', username: string = 'E2E夸克用户') {
  await page.goto('/accounts');
  if (await page.getByText(username).isVisible()) return;

  await page.getByRole('button', { name: /立即绑定账号|添加账号/ }).first().click();
  await page.getByText('夸克网盘', { exact: true }).click();
  await page.getByLabel('Cookie 全量字符串').fill(cookieStr);
  await page.getByRole('button', { name: '确认添加' }).click();
  await expect(page.getByText(username)).toBeVisible({ timeout: 10000 });
}
```

## Task 2: Implement Dynamic HTTP Mocks for Quark

**Files:**

- Modify: `internal/core/mock_http.go`

- [ ] **Step 1: Write the minimal implementation**

```go
// Replace the quark info/member sections in internal/core/mock_http.go
 // 1. 模拟夸克相关接口
 if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/save") {
  respBody = `{"code": 0, "message": "ok", "data": {"task_id": "mock_task_123"}}`
 } else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/task") {
  // 模拟任务成功
  respBody = `{"code": 0, "message": "ok", "data": {"status": 2, "message": "success"}}`
 } else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file/rename") {
  respBody = `{"code": 0, "message": "ok"}`
 } else if strings.Contains(url, "pan.quark.cn/account/info") {
  nickname := "E2E夸克用户"
  if strings.Contains(req.Header.Get("Cookie"), "mock_normal") {
   nickname = "E2E普通用户"
  } else if strings.Contains(req.Header.Get("Cookie"), "mock_overcap") {
   nickname = "E2E超容用户"
  }
  respBody = `{"code": 0, "data": {"nickname": "` + nickname + `"}}`
 } else if strings.Contains(url, "pan.quark.cn/1/clouddrive/member") || strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/capacity") {
  // 根据 Cookie 动态返回会员与容量
  memberType := "SUPER_VIP"
  totalCap := "1099511627776" // 1TB
  usedCap := "549755813888"   // 512GB

  if strings.Contains(req.Header.Get("Cookie"), "mock_normal") {
   memberType = "NORMAL"
   totalCap = "10737418240" // 10GB
   usedCap = "1073741824"   // 1GB
  } else if strings.Contains(req.Header.Get("Cookie"), "mock_overcap") {
   memberType = "SUPER_VIP"
   totalCap = "1099511627776" // 1TB
   usedCap = "2199023255552"  // 2TB (Over-capacity)
  }

  respBody = `{"code": 0, "member_type": "` + memberType + `", "data": {"total_capacity": ` + totalCap + `, "used_capacity": ` + usedCap + `, "use_capacity": ` + usedCap + `, "member_type": "` + memberType + `"}}`
 } else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/detail") {
```

## Task 3: Add Extended E2E Test Cases for Quark

**Files:**

- Modify: `e2e/tests/accounts/quark.spec.ts`

- [ ] **Step 1: Write the failing test / implementation**

```typescript
// Replace e2e/tests/accounts/quark.spec.ts content
import { test, expect } from '@playwright/test';
import { addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('夸克网盘账号管理', () => {
  test('成功绑定并展示夸克SVIP账号', async ({ page }) => {
    await addQuarkAccount(page);
    await expect(page.getByText('512').last()).toBeVisible();
    await expect(page.getByText('SVIP').last()).toBeVisible();
  });

  test('成功绑定并展示夸克普通用户小容量账号', async ({ page }) => {
    await addQuarkAccount(page, '__uid=mock; mock_normal', 'E2E普通用户');
    await expect(page.getByText('E2E普通用户')).toBeVisible();
    await expect(page.getByText('普通用户').last()).toBeVisible();
    await expect(page.getByText('10').last()).toBeVisible(); // 10GB
  });

  test('成功绑定并展示夸克超容状态账号', async ({ page }) => {
    await addQuarkAccount(page, '__uid=mock; mock_overcap', 'E2E超容用户');
    await expect(page.getByText('E2E超容用户')).toBeVisible();
    await expect(page.getByText('2048').last()).toBeVisible(); // 2TB used
  });
});
```

- [ ] **Step 2: Run test to verify it passes**

Run: `cd e2e && npx playwright test tests/accounts/quark.spec.ts`
Expected: PASS
