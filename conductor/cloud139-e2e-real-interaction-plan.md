# Cloud139 E2E Extended Coverage Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand E2E test coverage for 139 accounts (移动云盘) to include various VIP tiers and capacity scenarios by adding dynamic HTTP mocks based on Authorization headers.

**Architecture:**

1. Parameterize `add139Account` in `account.fixture.ts` to accept custom auth tokens and usernames.
2. Update `mock_http.go` to parse the `Authorization` header and dynamically return different nicknames, VIP levels (e.g., 普通用户, 钻石会员), and capacity data (Normal, Over-capacity).
3. Add tests to `cloud139.spec.ts` for different member states and capacity scenarios.

**Tech Stack:** Go (Mock HTTP Interceptor), TypeScript/Playwright (E2E Tests)

---

## Task 1: Parameterize 139 E2E Fixture

**Files:**

- Modify: `e2e/fixtures/account.fixture.ts`

- [ ] **Step 1: Write the minimal implementation**

```typescript
// Replace add139Account in e2e/fixtures/account.fixture.ts
export async function add139Account(page: Page, authStr: string = 'mock_auth', username: string = 'E2E移动云盘用户') {
  await page.goto('/accounts');
  // 如果已经存在，则不再重复添加（简单逻辑处理）
  if (await page.getByText(username).isVisible()) return;

  await page.getByRole('button', { name: /立即绑定账号|添加账号/ }).first().click();
  await page.getByText('移动云盘', { exact: true }).click();
  await page.getByLabel('Authorization').fill(authStr);
  await page.getByRole('button', { name: '确认添加' }).click();
  await expect(page.getByText(username)).toBeVisible({ timeout: 10000 });
}
```

## Task 2: Implement Dynamic HTTP Mocks for 139

**Files:**

- Modify: `internal/core/mock_http.go`

- [ ] **Step 1: Write the minimal implementation**

```go
// Replace the 139 mock section in internal/core/mock_http.go
 // 2. 模拟 139 相关接口
 if strings.Contains(url, "user-njs.yun.139.com/user/getUser") {
  nickname := "E2E移动云盘用户"
  if strings.Contains(req.Header.Get("Authorization"), "mock_normal") {
   nickname = "E2E139普通用户"
  } else if strings.Contains(req.Header.Get("Authorization"), "mock_overcap") {
   nickname = "E2E139超容用户"
  }
  respBody = `{"code": "0000", "success": true, "data": {"auditNickName": "` + nickname + `", "userName": "` + nickname + `", "userDomainId": "mock_domain", "loginName": "13800000000"}}`
 } else if strings.Contains(url, "user-njs.yun.139.com/user/disk/getPersonalDiskInfo") || strings.Contains(url, "user-njs.yun.139.com/user/disk/getFamilyDiskInfo") {
  // 返回 MB 单位
  diskSize := "1048576"   // 1TB (1024 * 1024 MB)
  freeDiskSize := "524288" // 512GB (512 * 1024 MB)
  
  if strings.Contains(req.Header.Get("Authorization"), "mock_normal") {
   diskSize = "20480" // 20GB
   freeDiskSize = "10240" // 10GB
  } else if strings.Contains(req.Header.Get("Authorization"), "mock_overcap") {
   diskSize = "1048576" // 1TB
   freeDiskSize = "-1048576" // -1TB -> Used: 2TB
  }
  respBody = `{"code": "0", "success": true, "data": {"diskSize": "` + diskSize + `", "freeDiskSize": "` + freeDiskSize + `"}}`
 } else if strings.Contains(url, "yun.139.com/orchestration/group-rebuild/member/v1.0/queryUserBenefits") {
  vipName := "黄金会员"
  if strings.Contains(req.Header.Get("Authorization"), "mock_normal") {
   vipName = "普通用户"
  } else if strings.Contains(req.Header.Get("Authorization"), "mock_overcap") {
   vipName = "钻石会员"
  }
  respBody = `{"code": "0", "success": true, "data": {"userSubMemberList": [{"memberLvName": "` + vipName + `"}]}}`
 } else if strings.Contains(url, "share-kd-njs.yun.139.com/yun-share/richlifeApp/devapp/IOutLink/getOutLinkInfoV6") {
```

## Task 3: Add Extended E2E Test Cases for 139

**Files:**

- Modify: `e2e/tests/accounts/cloud139.spec.ts`

- [ ] **Step 1: Write the failing test / implementation**

```typescript
// Replace content in e2e/tests/accounts/cloud139.spec.ts
import { test, expect } from '@playwright/test';
import { add139Account } from '../../fixtures/account.fixture';

test.describe('139 移动云盘账号管理', () => {
  test('成功绑定并展示 139 黄金会员账号', async ({ page }) => {
    await add139Account(page);
    await expect(page.getByText('E2E移动云盘用户')).toBeVisible();
    await expect(page.getByText('黄金会员').last()).toBeVisible();
    await expect(page.getByText('512').last()).toBeVisible(); // 512GB (MB 转换 GB 显示由于页面逻辑)
  });

  test('成功绑定并展示 139 普通用户小容量账号', async ({ page }) => {
    await add139Account(page, 'mock_normal', 'E2E139普通用户');
    await expect(page.getByText('E2E139普通用户')).toBeVisible();
    await expect(page.getByText('普通用户').last()).toBeVisible();
    await expect(page.getByText('10').last()).toBeVisible(); // 10GB used (20 - 10)
  });

  test('成功绑定并展示 139 超容状态账号', async ({ page }) => {
    await add139Account(page, 'mock_overcap', 'E2E139超容用户');
    await expect(page.getByText('E2E139超容用户')).toBeVisible();
    await expect(page.getByText('钻石会员').last()).toBeVisible();
    await expect(page.getByText('2 TB').last()).toBeVisible(); // 2TB used
    await expect(page.getByText('已超额 1 TB')).toBeVisible();
  });
});
```

- [ ] **Step 2: Run test to verify it passes**

Run: `cd e2e && npx playwright test tests/accounts/cloud139.spec.ts`
Expected: PASS
