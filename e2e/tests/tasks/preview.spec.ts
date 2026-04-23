import { test, expect } from '@playwright/test';
import { add139Account } from '../../fixtures/account.fixture';

test.describe('任务管理：重命名预览测试', () => {
  test.beforeEach(async ({ page }) => {
    await add139Account(page);
  });

  test('验证分享链接解析与正则重命名预览', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: /创建新任务/ }).click();
    
    // 补全：选择账号
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).click();

    await page.getByLabel('任务名称').fill('预览测试');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    await page.getByPlaceholder('匹配文件名的正则表达式').fill('.*\\.mp4$');
    await page.getByPlaceholder('支持 {TASKNAME}, {YEAR} 等变量').fill('[{DATE}] {TASKNAME}.{EXT}');
    
    await page.getByRole('button', { name: '全量重命名预览' }).click();

    const previewDialog = page.getByRole('dialog', { name: '重命名预览' });
    await expect(previewDialog).toBeVisible({ timeout: 15000 });
    await expect(previewDialog.getByText('[20240420] 预览测试.mp4').first()).toBeVisible();
  });
});
