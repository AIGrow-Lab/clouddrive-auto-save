# E2E 测试架构重构计划

> **给智能体的指示:** 推荐使用 subagent-driven-development 或 executing-plans 技能来逐项执行本计划。进度通过勾选框 (`- [ ]`) 进行追踪。

**目标:** 将原本臃肿且存在依赖关系的 `e2e/tests/core.spec.ts` 拆分为职责明确的模块化测试结构，并引入 Fixtures (测试夹具) 来消除测试间的耦合。

**架构说明:**

- **目录重构**: 建立 `tests/accounts` 和 `tests/tasks` 目录。
- **引入 Fixtures**: 创建 `e2e/utils/helpers.ts` 来封装“添加账号”等重复性 UI 动作。
- **解耦测试**: 使每个 `.spec.ts` 都能独立运行，不再依赖其他文件的执行结果。

**技术栈:** Playwright, TypeScript

---

## 任务 1: 创建目录结构并建立辅助工具函数

**涉及文件:**

- 新建: `e2e/utils/helpers.ts`

- [ ] **步骤 1: 创建辅助工具函数，封装 UI 常用操作**

我们将常用的“添加 139 账号”和“添加夸克账号”逻辑抽离，方便在各个测试中复用。

```typescript
// e2e/utils/helpers.ts
import { Page, expect } from '@playwright/test';

export async function add139Account(page: Page) {
  await page.goto('/accounts');
  // 如果已经存在，则不再重复添加（简单逻辑处理）
  if (await page.getByText('E2E移动云盘用户').isVisible()) return;

  await page.getByRole('button', { name: /立即绑定账号|添加账号/ }).click();
  await page.getByText('移动云盘', { exact: true }).click();
  await page.getByLabel('Authorization').fill('mock_auth');
  await page.getByRole('button', { name: '确认添加' }).click();
  await expect(page.getByText('E2E移动云盘用户')).toBeVisible({ timeout: 10000 });
}

export async function addQuarkAccount(page: Page) {
  await page.goto('/accounts');
  if (await page.getByText('E2E夸克用户').isVisible()) return;

  await page.getByRole('button', { name: /立即绑定账号|添加账号/ }).click();
  await page.getByText('夸克网盘', { exact: true }).click();
  await page.getByLabel('Cookie 全量字符串').fill('__uid=mock; mock_cookie');
  await page.getByRole('button', { name: '确认添加' }).click();
  await expect(page.getByText('E2E夸克用户')).toBeVisible({ timeout: 10000 });
}
```

## 任务 2: 拆分账号管理测试

**涉及文件:**

- 新建: `e2e/tests/accounts/cloud139.spec.ts`
- 新建: `e2e/tests/accounts/quark.spec.ts`

- [ ] **步骤 1: 实现 139 账号专项测试**

```typescript
// e2e/tests/accounts/cloud139.spec.ts
import { test, expect } from '@playwright/test';
import { add139Account } from '../../utils/helpers';

test.describe('139 移动云盘账号管理', () => {
  test('成功绑定并展示 139 账号', async ({ page }) => {
    await add139Account(page);
    await expect(page.locator('body')).toContainText('512'); // 验证容量 Mock 数据
  });
});
```

- [ ] **步骤 2: 实现夸克网盘专项测试**

```typescript
// e2e/tests/accounts/quark.spec.ts
import { test, expect } from '@playwright/test';
import { addQuarkAccount } from '../../utils/helpers';

test.describe('夸克网盘账号管理', () => {
  test('成功绑定并展示夸克账号', async ({ page }) => {
    await addQuarkAccount(page);
    await expect(page.getByText('512').last()).toBeVisible();
  });
});
```

## 任务 3: 拆分任务管理测试并解耦

**涉及文件:**

- 新建: `e2e/tests/tasks/create.spec.ts`

- [ ] **步骤 1: 实现任务创建与预览测试 (自动处理前置账号)**

在这个测试中，我们调用 helper 来确保账号存在，从而使本文件可以独立于账号测试运行。

```typescript
// e2e/tests/tasks/create.spec.ts
import { test, expect } from '@playwright/test';
import { add139Account } from '../../utils/helpers';

test.describe('任务管理功能测试', () => {
  test.beforeEach(async ({ page }) => {
    // 确保有账号可用，不关心账号是怎么来的
    await add139Account(page);
  });

  test('创建 139 转存任务并预览重命名结果', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: /创建新任务/ }).click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).click();
    
    await page.getByLabel('任务名称').fill('E2E测试电影');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    await page.getByPlaceholder('匹配文件名的正则表达式').fill('.*\\.mp4$');
    await page.getByPlaceholder('支持 {TASKNAME}, {YEAR} 等变量').fill('[{DATE}] {TASKNAME}.{EXT}');
    
    await page.getByRole('button', { name: '全量重命名预览' }).click();

    const previewDialog = page.getByRole('dialog', { name: '重命名预览' });
    await expect(previewDialog).toBeVisible();
    await expect(previewDialog.getByText('[20240420] E2E测试电影.mp4').first()).toBeVisible();
  });
});
```

## 任务 4: 清理旧文件并验证

**涉及文件:**

- 删除: `e2e/tests/core.spec.ts`

- [ ] **步骤 1: 删除已拆分的旧测试文件**

```bash
rm e2e/tests/core.spec.ts
```

- [ ] **步骤 2: 运行所有 E2E 测试验证重构是否成功**

```bash
make e2e-test
```
