# 修复创建同名文件夹冲突问题 (前端事前拦截)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在前端 `web/src/views/Tasks.vue` 的新建文件夹操作中，增加对当前目录下是否已存在同名文件夹的校验。如果存在，则直接拦截并提示用户，避免夸克网盘抛出 400 错误，同时防止移动云盘静默创建带有 `(1)` 后缀的副本。

**Architecture:** 
- 修改 `web/src/views/Tasks.vue` 中的 `handleInlineCreateFolder` 方法。
- 利用 Element Plus 的 `folderTreeRef` 获取当前选中节点的子节点列表，比对 `label` 与输入的新文件夹名称 `newFolderName`。

**Tech Stack:** Vue 3, Element Plus

---

### Task 1: 实现前端同名文件夹校验

**Files:**
- Modify: `web/src/views/Tasks.vue`

- [ ] **Step 1: 在 `handleInlineCreateFolder` 中添加事前拦截逻辑**
  在获取到 `currentPath` 后，真正发起 API 请求前，利用 `folderTreeRef` 获取当前节点的 `childNodes`，遍历检查是否有与 `newFolderName.value.trim()` 相同的 `label`。

```javascript
<<<<
  const currentPath = selectedTreePath.value || '/'
  // 从选中状态取 ID，若为空则默认取虚拟根节点的空字符串 ''
  const currentID = selectedTreeId.value || '' 

  creatingFolder.value = true
  try {
====
  const currentPath = selectedTreePath.value || '/'
  // 从选中状态取 ID，若为空则默认取虚拟根节点的空字符串 ''
  const currentID = selectedTreeId.value || '' 

  // 前端事前拦截：检查是否已存在同名文件夹
  if (folderTreeRef.value) {
    const currentNode = folderTreeRef.value.getNode(currentPath)
    if (currentNode && currentNode.childNodes) {
      const isDuplicate = currentNode.childNodes.some(
        child => child.data && child.data.label === newFolderName.value.trim()
      )
      if (isDuplicate) {
        return ElMessage.warning('该目录下已存在同名文件夹')
      }
    }
  }

  creatingFolder.value = true
  try {
>>>>
```

### Task 2: 验证与提交

- [ ] **Step 1: 编译测试**
  在 `web/` 目录下运行 `npm run build` 确保无语法错误。
- [ ] **Step 2: 自动提交并推送**
  按照功能逻辑分批 commit 并推送到远端 `main` 分支。
