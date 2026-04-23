import { test, expect } from '@playwright/test';

test.describe('仪表盘：数据概览测试', () => {
  test('成功进入仪表盘并展示基础统计卡片', async ({ page }) => {
    await page.goto('/');
    // 预期在仪表盘能看到“已规划任务”或“云端转存概览”等文字
    await expect(page.locator('body')).toContainText('云端转存概览');
  });
});
