import { test, expect } from '@playwright/test';
import { addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('夸克网盘账号管理', () => {
  test('成功绑定并展示夸克账号', async ({ page }) => {
    await addQuarkAccount(page);
    await expect(page.getByText('512').last()).toBeVisible();
    await expect(page.getByText('SVIP').last()).toBeVisible();
  });
});
