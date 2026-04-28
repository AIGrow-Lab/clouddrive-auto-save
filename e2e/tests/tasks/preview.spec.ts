import { test, expect } from '@playwright/test';
import { add139Account, addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('任务管理：重命名预览测试', () => {
  test.beforeEach(async ({ page }) => {
    await add139Account(page);
    await addQuarkAccount(page);
  });

  test('验证 139 移动云盘分享链接解析与重命名预览', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();

    await page.getByLabel('任务名称').fill('139预览测试');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    await page.getByPlaceholder('匹配文件名的正则表达式').fill('.*\\.mp4$');
    await page.getByPlaceholder('支持 {TASKNAME}, {YEAR} 等变量').fill('[{DATE}] {TASKNAME}.{EXT}');
    
    await page.getByRole('button', { name: '全量重命名预览' }).click();

    const previewDialog = page.getByRole('dialog', { name: '重命名预览' });
    await expect(previewDialog).toBeVisible({ timeout: 15000 });
    
    // 验证匹配的文件及其新名字和“匹配”标签
    // 使用 getByText 直接在对话框中查找，这样更稳健
    await expect(previewDialog.getByText('[20240420] 139预览测试.mp4').first()).toBeVisible({ timeout: 10000 });
    await expect(previewDialog.getByText('匹配').first()).toBeVisible();

    // 验证不匹配的文件保持原名
    await expect(previewDialog.getByText('readme.txt').first()).toBeVisible();
    await expect(previewDialog.getByText('未匹配').first()).toBeVisible();
    
    // 按 Esc 关闭预览
    await page.keyboard.press('Escape');
    await expect(previewDialog).not.toBeVisible();
  });

  test('验证夸克网盘分享链接解析与重命名预览', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).first().click();

    await page.getByLabel('任务名称').fill('夸克预览测试');
    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_link_id');
    await page.getByPlaceholder('匹配文件名的正则表达式').fill('.*\\.txt$');
    await page.getByPlaceholder('支持 {TASKNAME}, {YEAR} 等变量').fill('{TASKNAME}_已修改.{EXT}');
    
    await page.getByRole('button', { name: '全量重命名预览' }).click();

    const previewDialog = page.getByRole('dialog', { name: '重命名预览' });
    await expect(previewDialog).toBeVisible({ timeout: 15000 });
    
    // 验证匹配的文件
    await expect(previewDialog.getByText('夸克预览测试_已修改.txt').first()).toBeVisible({ timeout: 10000 });
    await expect(previewDialog.getByText('匹配').first()).toBeVisible();

    // 验证不匹配的文件
    await expect(previewDialog.getByText('[2024.04.20] E2E测试电影.mp4').first()).toBeVisible();
    await expect(previewDialog.getByText('未匹配').first()).toBeVisible();
    
    await page.keyboard.press('Escape');
    await expect(previewDialog).not.toBeVisible();
  });

  test('验证正则捕获组与空正则预览场景', async ({ page }) => {
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();

    await page.getByLabel('任务名称').fill('正则高级测试');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    
    // 测试：高级捕获组 (捕获电影名称，匹配空格后的部分)
    await page.getByPlaceholder('匹配文件名的正则表达式').fill('.* (.*)\\.mp4');
    await page.getByPlaceholder('支持 {TASKNAME}, {YEAR} 等变量').fill('${1}_高清版.mp4');
    
    await page.getByRole('button', { name: '全量重命名预览' }).click();

    const previewDialog = page.getByRole('dialog', { name: '重命名预览' });
    await expect(previewDialog).toBeVisible({ timeout: 15000 });
    
    // 验证捕获组替换正确 (从 [2024.04.20] E2E测试电影.mp4 提取出 E2E测试电影)
    await expect(previewDialog.getByText('E2E测试电影_高清版.mp4').first()).toBeVisible({ timeout: 10000 });
    
    await page.keyboard.press('Escape');
    await expect(previewDialog).not.toBeVisible();

    // 测试：空正则（匹配所有文件）
    await page.getByPlaceholder('匹配文件名的正则表达式').clear();
    await page.getByPlaceholder('支持 {TASKNAME}, {YEAR} 等变量').fill('添加前缀_{OLDNAME}');
    
    await page.getByRole('button', { name: '全量重命名预览' }).click();
    await expect(previewDialog).toBeVisible({ timeout: 15000 });

    // 验证所有文件均被匹配替换
    await expect(previewDialog.getByText('添加前缀_[2024.04.20] E2E测试电影.mp4').first()).toBeVisible({ timeout: 10000 });
    await expect(previewDialog.getByText('添加前缀_readme.txt').first()).toBeVisible();
    
    // 验证此时应该有多个“匹配”标签
    await expect(previewDialog.getByText('匹配').first()).toBeVisible();

    await page.keyboard.press('Escape');
  });
});
