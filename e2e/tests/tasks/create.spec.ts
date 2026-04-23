import { test, expect } from '@playwright/test';
import { add139Account } from '../../fixtures/account.fixture';

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
