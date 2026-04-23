import { test, expect } from '@playwright/test';

test.describe('任务管理：状态机与执行测试', () => {
  test('验证任务列表的基础渲染', async ({ page }) => {
    await page.goto('/tasks');
    await expect(page.getByRole('button', { name: /创建新任务/ })).toBeVisible();
  });
});
