import { test, expect } from '@playwright/test';

test.describe('UCAS 核心功能端到端测试', () => {
  
  test('账号管理：展示与容量检查 (HTTP Mock)', async ({ page }) => {
    // 1. 进入账号管理页面
    await page.goto('/accounts');

    // 2. 验证初始状态 (应显示 E2E 模式注入的测试账号)
    await expect(page.getByText('E2E测试账号(移动云盘)')).toBeVisible();
    await expect(page.getByText('E2E测试账号(夸克网盘)')).toBeVisible();
    
    // 3. 验证容量显示 (Mock 返回的是 512GB / 1TB)
    await expect(page.getByText('512 GB / 1 TB').first()).toBeVisible();
    await expect(page.getByText('剩 512 GB').first()).toBeVisible();
  });

  test('任务管理：创建任务与重命名预览 (AND 逻辑)', async ({ page }) => {
    await page.goto('/tasks');

    // 1. 点击创建任务 (此时列表为空，点击 el-empty 中的按钮)
    await page.getByRole('button', { name: '创建新任务' }).click();

    // 2. 填写任务信息
    // 点击选择账号下拉框
    await page.locator('.el-select').first().click();
    // 等待下拉列表出现并点击对应的测试账号
    await page.getByRole('option', { name: 'E2E测试账号(移动云盘)' }).click();
    
    await page.getByLabel('任务名称').fill('E2E测试电影');
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    await page.getByPlaceholder('匹配文件名的正则表达式').fill('.*\\.mp4$');
    await page.getByPlaceholder('支持 {TASKNAME}, {YEAR} 等变量').fill('[{DATE}] {TASKNAME}.{EXT}');
    
    // 3. 执行预览
    await page.getByRole('button', { name: '全量重命名预览' }).click();

    // 4. 验证预览结果契约 (等待预览对话框弹出)
    const previewDialog = page.getByRole('dialog', { name: '重命名预览' });
    await expect(previewDialog).toBeVisible();
    
    // 验证预览表格内容
    await expect(previewDialog.getByText('匹配', { exact: true })).toBeVisible();
    await expect(previewDialog.getByText('[20240420] E2E测试电影.mp4').first()).toBeVisible();
    await expect(previewDialog.getByText('未匹配', { exact: true })).toBeVisible();
    await expect(previewDialog.getByText('readme.txt').first()).toBeVisible();
  });

});
