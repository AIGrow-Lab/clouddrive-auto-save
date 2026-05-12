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
    ShareParentID string `gorm:"size:255" json:"share_parent_id"` // 139 分享链接的目录 ID (可选)
}
```

### 2. 后端：API 层支持

**文件**: `internal/api/router.go`

在 `updateTask` 函数的 `updateData` 中添加 `share_parent_id` 字段：

```go
updateData := map[string]interface{}{
    // ... existing fields ...
    "share_parent_id": task.ShareParentID,
    // ...
}
```

### 3. 前端：表单新增字段

**文件**: `web/src/views/Tasks.vue`

在表单初始化时添加 `share_parent_id` 字段：

```javascript
form.value = {
    // ... existing fields ...
    share_parent_id: ''
}
```

### 4. 前端：选择子文件夹时存储

在 `confirmSelectShareUrl` 函数中，对于 139 平台，存储 `share_parent_id`：

```javascript
if (account.platform === 'quark') {
    // 夸克网盘：替换 URL 中的 pdirFID
    // ...
    form.value.share_parent_id = ''
} else if (account.platform === '139') {
    // 移动云盘：URL 不变，但存储 share_parent_id
    form.value.share_parent_id = currentDirId || ''
    newUrl = originalUrl
}
```

### 5. 前端：打开选择起始文件弹窗时使用

在 `openStartFileDialog` 函数中，使用 `share_parent_id` 作为初始目录：

```javascript
const openStartFileDialog = async () => {
    // ...
    // 使用 share_parent_id 作为初始目录（139 平台）
    // 如果有 share_parent_id，将其作为新的根目录
    const initialParentId = form.value.share_parent_id || ''
    currentParentId.value = initialParentId
    await loadShareFiles(initialParentId)
}
```

### 6. 前端：浏览分享内容时使用

在 `openBrowseShareDialog` 函数中，根据当前分享链接确定初始目录：

```javascript
const openBrowseShareDialog = async () => {
    // ...
    // 根据当前分享链接确定初始目录
    const account = accounts.value.find(acc => acc.id === form.value.account_id)
    let initialParentId = ''

    if (account?.platform === 'quark') {
        // 夸克平台：从 URL 中解析 pdirFID
        const match = form.value.share_url.match(/\/s\/(\w+)#\/list\/share\/(\w+)/)
        if (match && match[2] && match[2] !== '0') {
            initialParentId = match[2]
        }
    } else if (account?.platform === '139') {
        // 139 平台：使用 share_parent_id
        initialParentId = form.value.share_parent_id || ''
    }

    currentParentId.value = initialParentId
    await loadShareFiles(initialParentId)
}
```

### 7. 前端：获取初始目录 ID

添加 `getInitialDirId` 辅助函数，统一获取初始目录 ID：

```javascript
const getInitialDirId = () => {
    const account = accounts.value.find(acc => acc.id === form.value.account_id)

    if (account?.platform === 'quark') {
        // 夸克平台：从 URL 中解析 pdirFID
        const match = form.value.share_url.match(/\/s\/(\w+)#\/list\/share\/(\w+)/)
        if (match && match[2] && match[2] !== '0') {
            return match[2]
        }
    } else if (account?.platform === '139') {
        // 139 平台：使用 share_parent_id
        return form.value.share_parent_id || ''
    }

    return ''
}
```

## 关键文件

- `internal/db/db.go` - Task 模型新增字段
- `internal/api/router.go` - API 层支持 share_parent_id
- `web/src/views/Tasks.vue` - 前端表单和逻辑

## 验证

1. 创建 139 任务，选择子文件夹作为分享链接
2. 保存任务后，再次编辑该任务
3. 点击"选择起始转存文件"，确认显示的是子目录内容而非根目录
4. 点击"浏览分享内容并选择目录"，确认显示的是当前分享链接对应的目录
5. 对于夸克网盘，确认从 URL 中正确解析 pdirFID
