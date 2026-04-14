# 账号删除逻辑优化方案

## 1. 目标 (Goal)
防止用户删除当前仍在被转存任务关联使用的云盘账号。采用强一致性原则：如果账号下存在关联任务，则后端拒绝删除请求并提示用户先清理任务。

## 2. 涉及文件与范围
- `internal/api/router.go`: 修改后端的 `deleteAccount` API 接口。
- `web/src/views/Accounts.vue`: 修改前端删除二次确认的提示文案，并处理删除失败时的异常流程。

## 3. 实施细节 (Implementation Steps)

### 3.1. 后端接口强化 (`internal/api/router.go`)
在执行 `db.DB.Delete` 之前，使用 GORM 的 `Count` 聚合查询来检索 `Task` 表中是否存在 `account_id` 为目标账号 ID 的记录。
- 如果 `count > 0`：不执行删除操作，返回 HTTP 状态码 `409 Conflict` (或 `400 Bad Request`)，并附带明确的错误信息：`{"error": "该账号有关联的 X 个任务，请先删除任务"}`。
- 如果 `count == 0`：按原逻辑正常执行账号软删除。

### 3.2. 前端交互优化 (`web/src/views/Accounts.vue`)
- **调整删除确认文案**：原先的“确定要删除该账号吗？关联的任务可能无法执行。”需修改为“确定要删除该账号吗？只有在没有关联任务的情况下才能成功删除。”。
- **完善异步流处理**：优化 `handleDelete` 函数，包裹 `try...catch` 以避免后端拦截拒绝时继续执行成功反馈：
```javascript
const handleDelete = (row) => {
  ElMessageBox.confirm('确定要删除该账号吗？只有在没有关联任务的情况下才能成功删除。', '删除账号', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await deleteAccount(row.id)
      ElMessage.success('已删除')
      fetchList()
    } catch (err) {
      // 后端拦截了删除请求，这里仅捕获不抛出，错误提示将由全局 axios 拦截器负责显示。
    }
  }).catch(() => {})
}
```

## 4. 测试与验证 (Verification)
1. 创建一个云盘账号，并为其创建一个对应的转存任务。
2. 在账号列表点击删除该账号，预期前端会弹出报错提示“该账号有关联的 X 个任务...”，账号列表数据不刷新/不消失。
3. 删除步骤 1 中创建的关联任务，再次回到账号列表尝试删除该账号，预期顺利删除，并弹出“已删除”。