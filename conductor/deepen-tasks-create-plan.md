# Deepen Tasks Create E2E Tests Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enhance the task creation E2E tests (`create.spec.ts`) by simulating user interactions for "start file selection" and "new folder creation in save path" using realistic mock API responses.

**Architecture:**

- In `create.spec.ts`, for both 139 and Quark task creation tests, add steps to:
  1. Click "选择文件" (Choose Start File).
  2. Wait for the dialog to load the parsed share files (mocked via `parseShareLink`).
  3. Select a specific file (e.g., `readme.txt`) and confirm.
  4. Assert the selected file name is populated in the input.
  5. Click "选择目录" (Choose Save Path).
  6. Wait for the folder tree dialog to open.
  7. Fill in a new folder name in the input and click "新建" (Create).
  8. Assert the newly created folder appears in the tree, select it, and confirm.
  9. Assert the selected folder path is populated in the save path input.

**Tech Stack:** TypeScript, Playwright

---

## Task 1: Update Create Task E2E Tests

**Files:**

- Modify: `e2e/tests/tasks/create.spec.ts`

- [ ] **Step 1: Write the updated implementation**

```typescript
// Replace content in e2e/tests/tasks/create.spec.ts
import { test, expect } from '@playwright/test';
import { add139Account, addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('任务管理：创建功能测试', () => {
  test.beforeEach(async ({ page }) => {
    await add139Account(page);
    await addQuarkAccount(page);
  });

  test('创建 139 移动云盘转存任务（包含高级选项）', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    
    await page.getByLabel('任务名称').fill('E2E_139_转存任务');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    
    // 测试：起始文件选择
    await page.getByRole('button', { name: '选择文件' }).click();
    const startFileDialog = page.getByRole('dialog', { name: '选择起始转存文件' });
    await expect(startFileDialog).toBeVisible();
    await startFileDialog.getByText('readme.txt').click();
    await startFileDialog.getByRole('button', { name: '确认选择' }).click();
    await expect(page.getByPlaceholder('从该文件开始向前转存 (为空则转存全量)')).toHaveValue('readme.txt');

    // 测试：选择保存目录与新建文件夹
    await page.getByRole('button', { name: '选择目录' }).click();
    const folderDialog = page.getByRole('dialog', { name: '选择保存目录' });
    await expect(folderDialog).toBeVisible();
    await folderDialog.getByPlaceholder('新文件夹名称').fill('139_new_folder');
    await folderDialog.getByRole('button', { name: '新建' }).click();
    
    await expect(folderDialog.getByText('139_new_folder')).toBeVisible();
    await folderDialog.getByText('139_new_folder').click();
    await folderDialog.getByRole('button', { name: '确认路径' }).click();
    
    await expect(page.getByPlaceholder('可手动输入或点击右侧选择')).toHaveValue('/139_new_folder');
    
    await page.getByRole('button', { name: '确认并保存' }).click();

    // 验证回到任务列表并展示该任务
    await expect(page.getByText('E2E_139_转存任务')).toBeVisible({ timeout: 10000 });
  });

  test('创建夸克网盘转存任务（包含高级选项）', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).first().click();
    
    await page.getByLabel('任务名称').fill('E2E_Quark_转存任务');
    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_link_id');
    
    // 测试：起始文件选择
    await page.getByRole('button', { name: '选择文件' }).click();
    const startFileDialog = page.getByRole('dialog', { name: '选择起始转存文件' });
    await expect(startFileDialog).toBeVisible();
    await startFileDialog.getByText('readme.txt').click();
    await startFileDialog.getByRole('button', { name: '确认选择' }).click();
    await expect(page.getByPlaceholder('从该文件开始向前转存 (为空则转存全量)')).toHaveValue('readme.txt');

    // 测试：选择保存目录与新建文件夹
    await page.getByRole('button', { name: '选择目录' }).click();
    const folderDialog = page.getByRole('dialog', { name: '选择保存目录' });
    await expect(folderDialog).toBeVisible();
    await folderDialog.getByPlaceholder('新文件夹名称').fill('quark_new_folder');
    await folderDialog.getByRole('button', { name: '新建' }).click();
    
    await expect(folderDialog.getByText('quark_new_folder')).toBeVisible();
    await folderDialog.getByText('quark_new_folder').click();
    await folderDialog.getByRole('button', { name: '确认路径' }).click();
    
    await expect(page.getByPlaceholder('可手动输入或点击右侧选择')).toHaveValue('/quark_new_folder');

    await page.getByRole('button', { name: '确认并保存' }).click();

    // 验证回到任务列表并展示该任务
    await expect(page.getByText('E2E_Quark_转存任务')).toBeVisible({ timeout: 10000 });
  });
});
```

- [ ] **Step 2: Run test to verify it passes**

Run: `cd e2e && npx playwright test tests/tasks/create.spec.ts`
Expected: PASS
