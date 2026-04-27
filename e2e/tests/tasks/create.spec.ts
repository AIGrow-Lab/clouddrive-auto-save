import { test, expect } from '@playwright/test';
import { add139Account, addQuarkAccount } from '../../fixtures/account.fixture';

test.describe('任务管理：创建功能测试', () => {
  test.beforeEach(async ({ page }) => {
    await add139Account(page);
    await addQuarkAccount(page);
  });

  test('创建 139 移动云盘转存任务（包含高级选项）', async ({ page }) => {
    const taskName = `E2E_139_转存_${Date.now()}`;
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E移动云盘用户' }).first().click();
    
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://yun.139.com/w/#/share/link/mock_link_id');
    
    // 测试：起始文件选择
    await page.getByRole('button', { name: '选择文件' }).click();
    const startFileDialog = page.getByRole('dialog', { name: '选择起始转存文件' });
    await expect(startFileDialog).toBeVisible();
    await startFileDialog.getByText('readme.txt').first().click();
    await startFileDialog.getByRole('button', { name: '确认选择' }).click();
    await expect(page.getByPlaceholder('从该文件开始向前转存 (为空则转存全量)')).toHaveValue('readme.txt');

    // 测试：选择保存目录与新建文件夹 (基于已有目录)
    await page.getByRole('button', { name: '选择目录' }).click();
    const folderDialog = page.getByRole('dialog', { name: '选择保存目录' });
    await expect(folderDialog).toBeVisible();
    
    // 等待已有目录渲染并点击选中
    await expect(folderDialog.getByText('139已有目录')).toBeVisible();
    await folderDialog.getByText('139已有目录').click();

    // 在选中的已有目录下新建文件夹
    await folderDialog.getByPlaceholder('新文件夹名称').fill('139_new_folder');
    await folderDialog.getByRole('button', { name: '新建' }).click();
    
    // 验证新文件夹出现并选中
    await expect(folderDialog.getByText('139_new_folder')).toBeVisible();
    await folderDialog.getByText('139_new_folder').click();
    await folderDialog.getByRole('button', { name: '确认路径' }).click();
    
    // 验证最终生成的路径包含父目录
    await expect(page.getByPlaceholder('可手动输入或点击右侧选择')).toHaveValue('/139已有目录/139_new_folder');
    
    await page.getByRole('button', { name: '确认并保存' }).click();

    // 验证回到任务列表并展示该任务
    await expect(page.getByText(taskName)).toBeVisible({ timeout: 10000 });
  });

  test('创建夸克网盘转存任务（包含高级选项）', async ({ page }) => {
    const taskName = `E2E_Quark_转存_${Date.now()}`;
    await page.goto('/tasks');
    await page.getByRole('button', { name: '创建任务' }).last().click();

    await page.locator('.el-select').first().click();
    await page.getByRole('option', { name: 'E2E夸克用户' }).first().click();
    
    await page.getByLabel('任务名称').fill(taskName);
    await page.getByLabel('分享链接').fill('https://pan.quark.cn/s/mock_link_id');
    
    // 测试：起始文件选择
    await page.getByRole('button', { name: '选择文件' }).click();
    const startFileDialog = page.getByRole('dialog', { name: '选择起始转存文件' });
    await expect(startFileDialog).toBeVisible();
    await startFileDialog.getByText('readme.txt').first().click();
    await startFileDialog.getByRole('button', { name: '确认选择' }).click();
    await expect(page.getByPlaceholder('从该文件开始向前转存 (为空则转存全量)')).toHaveValue('readme.txt');

    // 测试：选择保存目录与新建文件夹 (基于已有目录)
    await page.getByRole('button', { name: '选择目录' }).click();
    const folderDialog = page.getByRole('dialog', { name: '选择保存目录' });
    await expect(folderDialog).toBeVisible();
    
    // 等待已有目录渲染并点击选中
    await expect(folderDialog.getByText('夸克已有目录')).toBeVisible();
    await folderDialog.getByText('夸克已有目录').click();

    // 在选中的已有目录下新建文件夹
    await folderDialog.getByPlaceholder('新文件夹名称').fill('quark_new_folder');
    await folderDialog.getByRole('button', { name: '新建' }).click();
    
    // 验证新文件夹出现并选中
    await expect(folderDialog.getByText('quark_new_folder')).toBeVisible();
    await folderDialog.getByText('quark_new_folder').click();
    await folderDialog.getByRole('button', { name: '确认路径' }).click();
    
    // 验证最终生成的路径包含父目录
    await expect(page.getByPlaceholder('可手动输入或点击右侧选择')).toHaveValue('/夸克已有目录/quark_new_folder');

    await page.getByRole('button', { name: '确认并保存' }).click();

    // 验证回到任务列表并展示该任务
    await expect(page.getByText(taskName)).toBeVisible({ timeout: 10000 });
  });
});
