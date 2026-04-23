import { test, expect } from '@playwright/test';
import { add139Account } from '../../fixtures/account.fixture';

test.describe('139 移动云盘账号管理', () => {
  test('成功绑定并展示 139 账号', async ({ page }) => {
    await add139Account(page);
    await expect(page.locator('body')).toContainText('512'); // 验证容量 Mock 数据
  });
});
