import { test, expect } from '@playwright/test';
import { add139Account, addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('任务管理：状态机与执行测试', () => {
  test.beforeEach(async ({ page }) => {
    await add139Account(page);
    await addQuarkAccount(page);
  });

  test('手动执行 139 转存任务并验证成功状态及文件入库', async ({ page }) => {
    const taskName = `E2E_139_成功_${Date.now()}`;
    const shareUrl = 'https://yun.139.com/w/#/share/link/mock_success';
    
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill(shareUrl);
    await page.getByLabel('保存路径').fill('/139_exec');
    await page.getByRole('button', { name: '确认并保存' }).click();
    
    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();

    // 点击执行按钮
    await taskRow.getByRole('button', { name: '运行' }).click();

    // 验证状态变更为“SUCCESS”
    const updatedTaskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(updatedTaskRow.locator('.el-tag').filter({ hasText: 'SUCCESS' })).toBeVisible({ timeout: 60000 });

    // 校验：文件是否真的被转存成功了 (通过起始文件选择器查看是否显示“已在网盘”)
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('分享链接').fill(shareUrl);
    await page.getByLabel('保存路径').fill('/139_exec');
    
    // 等待解析接口完成（会有一次加载效果或列表更新）
    // 我们可以通过等待“选择文件”按钮可用或直接点击它
    await page.getByRole('button', { name: '选择文件' }).click();
    const startFileDialog = page.getByRole('dialog', { name: '选择起始转存文件' });
    await expect(startFileDialog).toBeVisible();
    
    // 验证 mock 数据中的文件现在标记为“已在网盘”
    // 由于后端 mock 返回的是 JSON，前端应该渲染成功。增加一个更长的超时。
    const existedTag = startFileDialog.getByText('已在网盘').first();
    await expect(existedTag).toBeVisible({ timeout: 20000 });
  });

  test('手动执行夸克网盘转存任务并验证成功状态及文件入库', async ({ page }) => {
    const taskName = `E2E_Quark_成功_${Date.now()}`;
    const shareUrl = 'https://pan.quark.cn/s/mock_success';
    
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill(shareUrl);
    await page.getByLabel('保存路径').fill('/quark_exec');
    await page.getByRole('button', { name: '确认并保存' }).click();
    
    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(taskRow).toBeVisible();

    // 点击执行按钮
    await taskRow.getByRole('button', { name: '运行' }).click();

    // 验证状态变更为“SUCCESS”
    const updatedTaskRow = page.locator('tr').filter({ hasText: taskName });
    await expect(updatedTaskRow.locator('.el-tag').filter({ hasText: 'SUCCESS' })).toBeVisible({ timeout: 60000 });

    // 校验：文件是否真的被转存成功了 (通过起始文件选择器查看是否显示“已在网盘”)
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).first().click();
    await page.getByLabel('分享链接').fill(shareUrl);
    await page.getByLabel('保存路径').fill('/quark_exec');
    
    // 等待解析接口完成
    await page.getByRole('button', { name: '选择文件' }).click();
    const startFileDialog = page.getByRole('dialog', { name: '选择起始转存文件' });
    await expect(startFileDialog).toBeVisible();
    
    // 验证 mock 数据中的文件现在标记为“已在网盘”
    const existedTag = startFileDialog.getByText('已在网盘').first();
    await expect(existedTag).toBeVisible({ timeout: 20000 });
  });

  test('139 移动云盘：验证分享链接失效场景', async ({ page }) => {
    const taskName = `139_失效_${Date.now()}`;
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_invalid');
    await page.getByLabel('保存路径').fill('/139_error');
    await page.getByRole('button', { name: '确认并保存' }).click();
    
    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await taskRow.getByRole('button', { name: '运行' }).click();

    await page.waitForTimeout(5000);
    await page.reload();

    const updatedTaskRow = page.locator('tr').filter({ hasText: taskName });
    const errorTag = updatedTaskRow.locator('.el-tag--danger').filter({ hasText: 'LINK ERROR' });
    await expect(errorTag).toBeVisible({ timeout: 15000 });
    
    await errorTag.hover();
    await expect(page.getByText('分享链接不存在或已被取消。')).toBeVisible();
  });

  test('夸克网盘：验证分享内容违规被屏蔽场景', async ({ page }) => {
    const taskName = `Quark_违规_${Date.now()}`;
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();
    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).first().click();
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_violation');
    await page.getByRole('button', { name: '确认并保存' }).click();
    
    const taskRow = page.locator('tr').filter({ hasText: taskName });
    await taskRow.getByRole('button', { name: '运行' }).click();

    await page.waitForTimeout(5000);
    await page.reload();

    const updatedTaskRow = page.locator('tr').filter({ hasText: taskName });
    const errorTag = updatedTaskRow.locator('.el-tag--danger').filter({ hasText: 'LINK ERROR' });
    await expect(errorTag).toBeVisible({ timeout: 15000 });
    
    await errorTag.hover();
    await expect(page.getByText('该分享文件涉及违规内容，已被官方屏蔽。')).toBeVisible();
  });
});
