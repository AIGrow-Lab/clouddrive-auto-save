import { Page, expect } from '@playwright/test';

export async function add139Account(page: Page) {
  await page.goto('/accounts');
  // 如果已经存在，则不再重复添加（简单逻辑处理）
  if (await page.getByText('E2E移动云盘用户').isVisible()) return;

  await page.getByRole('button', { name: /立即绑定账号|添加账号/ }).first().click();
  await page.getByText('移动云盘', { exact: true }).click();
  await page.getByLabel('Authorization').fill('mock_auth');
  await page.getByRole('button', { name: '确认添加' }).click();
  await expect(page.getByText('E2E移动云盘用户')).toBeVisible({ timeout: 10000 });
}

export async function addQuarkAccount(page: Page, cookieStr: string = '__uid=mock; mock_cookie', username: string = 'E2E夸克用户') {
  await page.goto('/accounts');
  if (await page.getByText(username).isVisible()) return;

  await page.getByRole('button', { name: /立即绑定账号|添加账号/ }).first().click();
  await page.getByText('夸克网盘', { exact: true }).click();
  await page.getByLabel('Cookie 全量字符串').fill(cookieStr);
  await page.getByRole('button', { name: '确认添加' }).click();
  await expect(page.getByText(username)).toBeVisible({ timeout: 10000 });
}
