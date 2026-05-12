# 移动云盘分享链接子目录支持设计

## 背景

移动云盘（139）的分享链接 URL 不支持通过参数区分目录（不像夸克网盘可以通过 URL 中的 pdirFID 区分）。当用户选择子文件夹作为新的分享链接后，URL 不变，导致再次打开"选择起始转存文件"时，系统还是解析原来的根目录内容。

## 设计方案

### 1. 后端：Task 模型新增字段

**文件**: `internal/db/db.go`

在 Task 结构体中新增 `ShareParentID` 字段：

```go
type Task struct {
    // ... existing fields ...
    ShareParentID string `json:"share_parent_id"` // 139 分享链接的目录 ID
}
```

### 2. 前端：表单新增字段

**文件**: `web/src/views/Tasks.vue`

在表单初始化时添加 `share_parent_id` 字段：

```javascript
form.value = {
    // ... existing fields ...
    share_parent_id: ''
}
```

### 3. 前端：选择子文件夹时存储

在 `confirmSelectShareUrl` 函数中，对于 139 平台，存储 `share_parent_id`：

```javascript
if (account.platform === '139') {
    // 139 通过 pCaID 区分目录，URL 格式不变
    // 存储当前目录 ID 为 share_parent_id
    form.value.share_parent_id = currentDirId || ''
}
```

### 4. 前端：调用 parseShareLink 时传递

在 `loadShareFiles` 函数中，传递 `share_parent_id` 作为 `parent_id`：

```javascript
const data = await parseShareLink({
    account_id: form.value.account_id,
    share_url: form.value.share_url,
    extract_code: form.value.extract_code,
    parent_id: parentId || form.value.share_parent_id, // 优先使用传入的 parentId，否则使用 share_parent_id
    // ...
})
```

### 5. 前端：打开选择起始文件弹窗时使用

在 `openStartFileDialog` 函数中，使用 `share_parent_id` 作为初始目录：

```javascript
const openStartFileDialog = async () => {
    // ...
    currentParentId.value = form.value.share_parent_id || ''
    await loadShareFiles(form.value.share_parent_id || '')
}
```

## 关键文件

- `internal/db/db.go` - Task 模型新增字段
- `web/src/views/Tasks.vue` - 前端表单和逻辑

## 验证

1. 创建 139 任务，选择子文件夹作为分享链接
2. 保存任务后，再次编辑该任务
3. 点击"选择起始转存文件"，确认显示的是子目录内容而非根目录
4. 对于夸克网盘，确认行为不受影响
