# Refine Tasks Create E2E with Existing Folders Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enhance the task creation E2E tests to interact with existing folders on the cloud drive. We will mock the file list APIs to return pre-existing directories, and have the E2E tests expand these directories and create new sub-folders inside them.

**Architecture:**

1. **Mock HTTP Enhancement**:
   - For Quark (`file/sort`), check the `pdir_fid` query parameter. If it's the root (`0`), return a list containing a folder named "夸克已有目录". Otherwise, return an empty list.
   - For 139 (`hcy/file/list`), read the request body to check `parentFileId`. If it's `root` or `/`, return a folder named "139已有目录". Otherwise, return an empty list.
2. **E2E Test Updates**:
   - Modify `e2e/tests/tasks/create.spec.ts` so that when the "选择保存目录" dialog opens, the test clicks on the existing folder ("139已有目录" or "夸克已有目录") to select/expand it.
   - Then, input the new folder name and click "新建".
   - Verify that the resulting path is correctly nested, e.g., `/139已有目录/139_new_folder` and `/夸克已有目录/quark_new_folder`.

**Tech Stack:** Go (Mock HTTP Interceptor), TypeScript / Playwright (E2E Tests)

---

## Task 1: Update HTTP Mock to Return Existing Folders

**Files:**

- Modify: `internal/core/mock_http.go`

- [ ] **Step 1: Update the Quark file list mock**

```go
// In internal/core/mock_http.go, replace the Quark file/sort mock logic:
 } else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file/sort") {
  // 目标目录文件列表 (预检)
  if strings.Contains(url, "pdir_fid=0") {
   respBody = `{"code": 0, "data": {"list": [{"fid": "quark_exist_dir", "file_name": "夸克已有目录", "dir": true, "size": 0, "updated_at": 1612345678000}]}}`
  } else {
   respBody = `{"code": 0, "data": {"list": []}}`
  }
 } else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/file") && req.Method == "POST" {
```

- [ ] **Step 2: Update the 139 file list mock**

```go
// In internal/core/mock_http.go, replace the 139 hcy/file/list mock logic:
 } else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/list") {
  bodyBytes, _ := io.ReadAll(req.Body)
  req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
  bodyStr := string(bodyBytes)
  if strings.Contains(bodyStr, `"parentFileId":"root"`) || strings.Contains(bodyStr, `"parentFileId":"/"`) {
   respBody = `{"code": "0", "success": true, "data": {"items": [{"fileId": "139_exist_dir", "name": "139已有目录", "type": "folder", "category": "folder", "size": 0, "updatedAt": "2024-04-20 12:00:00"}]}}`
  } else {
   respBody = `{"code": "0", "success": true, "data": {"items": []}}`
  }
 } else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/update") {
```

## Task 2: Update E2E Create Tests for Nested Folder Creation

**Files:**

- Modify: `e2e/tests/tasks/create.spec.ts`

- [ ] **Step 1: Update the 139 Create Task E2E Test**

```typescript
// In e2e/tests/tasks/create.spec.ts, replace the "选择保存目录与新建文件夹" section for 139:
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
```

- [ ] **Step 2: Update the Quark Create Task E2E Test**

```typescript
// In e2e/tests/tasks/create.spec.ts, replace the "选择保存目录与新建文件夹" section for Quark:
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
```

- [ ] **Step 3: Run test to verify it passes**

Run: `cd e2e && npx playwright test tests/tasks/create.spec.ts`
Expected: PASS
