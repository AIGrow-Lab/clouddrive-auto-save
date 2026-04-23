import { test, expect } from '@playwright/test';

test.describe('系统设置：全局配置测试', () => {
  test('进入设置页面', async ({ page }) => {
    await page.goto('/settings');
    // 目前设置页暂无固定内容，先验证能成功访问
    await expect(page).not.toHaveURL('/404');
  });
});
