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
    await expect(page.getByText('2 TB').last()).toBeVisible(); // 2TB used
    await expect(page.getByText('已超额 1 TB')).toBeVisible();
  });

  test('成功绑定并展示夸克SVIP+账号', async ({ page }) => {
    await addQuarkAccount(page, '__uid=mock; mock_svipplus', 'E2E超超级会员');
    await expect(page.getByText('SVIP+').last()).toBeVisible();
    await expect(page.getByText('10 TB').last()).toBeVisible();
  });
});
