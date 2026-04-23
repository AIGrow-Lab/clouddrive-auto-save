import { test, expect } from '@playwright/test';

test.describe('UCAS 核心功能端到端测试', () => {
  
  test('账号管理：添加与展示 (真实 UI 交互 + HTTP Mock)', async ({ page }) => {
    // 1. 进入账号管理页面
    await page.goto('/accounts');

    // 2. 交互：添加 139 账号
    await page.getByRole('button', { name: '立即绑定账号' }).click();
    await page.getByText('移动云盘', { exact: true }).click();
    await page.getByLabel('Authorization').fill('mock_auth');
    
    // 点击确认并等待请求返回
    const createReq = page.waitForResponse(resp => resp.url().includes('/api/accounts') && resp.status() === 200);
    await page.getByRole('button', { name: '确认添加' }).click();
    await createReq;

    // 3. 验证 139 账号渲染 (增加显式等待)
    await expect(page.getByText('E2E移动云盘用户')).toBeVisible({ timeout: 10000 });
    // 只要页面上出现了 512 这个关键数字，说明容量加载成功
    await expect(page.locator('body')).toContainText('512');

    // 4. 交互：添加 Quark 账号
    await page.getByRole('button', { name: '添加账号' }).click();
    await page.getByText('夸克网盘', { exact: true }).click();
    await page.getByLabel('Cookie 全量字符串').fill('__uid=mock; mock_cookie');

    const createReqQuark = page.waitForResponse(resp => resp.url().includes('/api/accounts') && resp.status() === 200);
    await page.getByRole('button', { name: '确认添加' }).click();
    await createReqQuark;

    // 5. 验证 Quark 账号渲染
    await expect(page.getByText('E2E夸克用户')).toBeVisible({ timeout: 10000 });
    // 此时页面上应该至少有两个地方出现了 512 (两个账号都是这个 Mock 值)
    await expect(page.getByText('512').last()).toBeVisible();
    });


  test('任务管理：创建任务与重命名预览 (AND 逻辑)', async ({ page }) => {
    await page.goto('/tasks');

    // 1. 点击创建任务 (此时列表为空，点击 el-empty 中的按钮)
    await page.getByRole('button', { name: '创建新任务' }).click();

    // 2. 填写任务信息
    // 点击选择账号下拉框
    await page.locator('.el-select').first().click();
    // 等待下拉列表出现并点击对应的测试账号 (名称应与 Test 1 添加的一致)
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).click();
    
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
