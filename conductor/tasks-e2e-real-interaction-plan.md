# Tasks Module E2E Test Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Provide comprehensive E2E test coverage for the Tasks module, structured by CloudDrive type (Quark and 139) and functional areas (Preview, Create, Execute), interacting with actual backend business logic and HTTP mocks.

**Architecture:** 
1. **Preview (`preview.spec.ts`)**: Test the "Preview" button logic. Ensure that parsing share links for both Quark and 139 correctly retrieves file lists from the mock backend and applies renaming rules correctly.
2. **Create (`create.spec.ts`)**: Test the task creation flow. Submit the form for both Quark and 139, ensuring the tasks are successfully created and displayed in the task list.
3. **Execute (`execute.spec.ts`)**: Test the task execution lifecycle. Manually trigger a task execution from the list and assert that the status updates from "Running" to "Success", reflecting the backend mock responses (e.g., Quark's async task polling and 139's batch task creation).

**Tech Stack:** TypeScript, Playwright

---

### Task 1: Implement Preview E2E Tests

**Files:**
- Modify: `e2e/tests/tasks/preview.spec.ts`

- [ ] **Step 1: Write the failing test / implementation**

```typescript
// Replace content in e2e/tests/tasks/preview.spec.ts
import { test, expect } from '@playwright/test';
import { add139Account, addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('任务管理：重命名预览测试', () => {
  test.beforeEach(async ({ page }) => {
    await add139Account(page);
    await addQuarkAccount(page);
  });

  test('验证 139 移动云盘分享链接解析与重命名预览', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: /创建新任务/ }).click();
    
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).click();

    await page.getByLabel('任务名称').fill('139预览测试');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    await page.getByPlaceholder('匹配文件名的正则表达式').fill('.*\\.mp4$');
    await page.getByPlaceholder('支持 {TASKNAME}, {YEAR} 等变量').fill('[{DATE}] {TASKNAME}.{EXT}');
    
    await page.getByRole('button', { name: '全量重命名预览' }).click();

    const previewDialog = page.getByRole('dialog', { name: '重命名预览' });
    await expect(previewDialog).toBeVisible({ timeout: 15000 });
    // 验证能够抓取到文件，并且正则和变量替换正常工作
    await expect(previewDialog.getByText('[20240420] 139预览测试.mp4').first()).toBeVisible();
    await page.getByRole('button', { name: '关闭' }).click();
  });

  test('验证夸克网盘分享链接解析与重命名预览', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: /创建新任务/ }).click();
    
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).click();

    await page.getByLabel('任务名称').fill('夸克预览测试');
    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_link_id');
    await page.getByPlaceholder('匹配文件名的正则表达式').fill('.*\\.txt$');
    await page.getByPlaceholder('支持 {TASKNAME}, {YEAR} 等变量').fill('{TASKNAME}_已修改.{EXT}');
    
    await page.getByRole('button', { name: '全量重命名预览' }).click();

    const previewDialog = page.getByRole('dialog', { name: '重命名预览' });
    await expect(previewDialog).toBeVisible({ timeout: 15000 });
    // 验证能够抓取到文件，并且正则和变量替换正常工作
    await expect(previewDialog.getByText('夸克预览测试_已修改.txt').first()).toBeVisible();
    await page.getByRole('button', { name: '关闭' }).click();
  });
});
```

### Task 2: Implement Create Task E2E Tests

**Files:**
- Modify: `e2e/tests/tasks/create.spec.ts`

- [ ] **Step 1: Write the failing test / implementation**

```typescript
// Replace content in e2e/tests/tasks/create.spec.ts
import { test, expect } from '@playwright/test';
import { add139Account, addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('任务管理：创建功能测试', () => {
  test.beforeEach(async ({ page }) => {
    await add139Account(page);
    await addQuarkAccount(page);
  });

  test('创建 139 移动云盘转存任务', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: /创建新任务/ }).click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).click();
    
    await page.getByLabel('任务名称').fill('E2E_139_转存任务');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    await page.getByLabel('保存路径').fill('/139_sync_folder');
    
    await page.getByRole('button', { name: '保存任务' }).click();

    // 验证回到任务列表并展示该任务
    await expect(page.getByText('E2E_139_转存任务')).toBeVisible({ timeout: 10000 });
  });

  test('创建夸克网盘转存任务', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: /创建新任务/ }).click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).click();
    
    await page.getByLabel('任务名称').fill('E2E_Quark_转存任务');
    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_link_id');
    await page.getByLabel('保存路径').fill('/quark_sync_folder');
    
    await page.getByRole('button', { name: '保存任务' }).click();

    // 验证回到任务列表并展示该任务
    await expect(page.getByText('E2E_Quark_转存任务')).toBeVisible({ timeout: 10000 });
  });
});
```

### Task 3: Implement Execute Task E2E Tests

**Files:**
- Modify: `e2e/tests/tasks/execute.spec.ts`

- [ ] **Step 1: Write the failing test / implementation**

```typescript
// Replace content in e2e/tests/tasks/execute.spec.ts
import { test, expect } from '@playwright/test';
import { add139Account, addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('任务管理：状态机与执行测试', () => {
  test.beforeEach(async ({ page }) => {
    await add139Account(page);
    await addQuarkAccount(page);
  });

  test('手动执行 139 转存任务并验证成功状态', async ({ page }) => {
    // 1. 创建任务
    await page.goto('/tasks');
    await page.getByRole('button', { name: /创建新任务/ }).click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).click();
    await page.getByLabel('任务名称').fill('E2E_139_执行测试');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    await page.getByLabel('保存路径').fill('/139_exec');
    await page.getByRole('button', { name: '保存任务' }).click();
    await expect(page.getByText('E2E_139_执行测试')).toBeVisible();

    // 2. 点击执行按钮 (使用 tooltip 或 aria-label 定位执行按钮)
    const taskRow = page.locator('tr').filter({ hasText: 'E2E_139_执行测试' });
    await taskRow.getByRole('button').filter({ has: page.locator('.lucide-play') }).click();

    // 3. 验证状态变更为“成功” (由于 Mock 极快，可能直接跳到成功)
    await expect(taskRow.locator('.el-tag--success').filter({ hasText: '成功' })).toBeVisible({ timeout: 15000 });
  });

  test('手动执行夸克转存任务并验证成功状态', async ({ page }) => {
    // 1. 创建任务
    await page.goto('/tasks');
    await page.getByRole('button', { name: /创建新任务/ }).click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).click();
    await page.getByLabel('任务名称').fill('E2E_Quark_执行测试');
    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_link_id');
    await page.getByLabel('保存路径').fill('/quark_exec');
    await page.getByRole('button', { name: '保存任务' }).click();
    await expect(page.getByText('E2E_Quark_执行测试')).toBeVisible();

    // 2. 点击执行按钮
    const taskRow = page.locator('tr').filter({ hasText: 'E2E_Quark_执行测试' });
    await taskRow.getByRole('button').filter({ has: page.locator('.lucide-play') }).click();

    // 3. 验证状态变更为“成功”
    await expect(taskRow.locator('.el-tag--success').filter({ hasText: '成功' })).toBeVisible({ timeout: 15000 });
  });
});
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `cd e2e && npx playwright test tests/tasks/preview.spec.ts tests/tasks/create.spec.ts tests/tasks/execute.spec.ts`
Expected: PASS
