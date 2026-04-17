# 目录树加载性能与交互优化计划 (Folder Tree Performance Optimization Plan)

## 1. Objective (目标)
消除在编辑任务选择目录时，深层目录展开产生的“一行行刷新”视觉差及无反馈卡顿感，提升整体交互的丝滑度。

## 2. Key Files & Context (核心影响文件)
- `web/src/views/Tasks.vue`
  - 核心逻辑：`loadFolders` (懒加载处理函数)。
  - UI 组件：`el-tree` 及其相关的 `v-loading` 状态。

## 3. Implementation Steps (实施步骤)

### Step 3.1: 增强全层级加载反馈
1. 修改 `loadFolders` 函数。
2. 移除原有的 `if (node.level === 1)` 限制，确保所有层级的网络请求期间都激活 `loadingFolders` 状态。
3. 优化 `v-loading` 的显示逻辑，使其覆盖整个树形容器。

### Step 3.2: 优化响应式数据处理
1. 在 `loadFolders` 接收到后端数据后，不再在循环中逐个更新 `pathIdMap`。
2. 先构建一个临时的局部 map。
3. 使用 `Object.assign` 或解构赋值一次性合并到 `pathIdMap.value` 中，触发单一的响应式更新流。

### Step 3.3: 提升渲染效率 (可选样式优化)
1. 检查并确保树节点的 CSS 动画简洁，避免不必要的过渡计算。

## 4. Verification & Testing (验证与测试)
- 点击根目录：确认出现加载动画，数据整体展现。
- 点击二级或更深层目录：确认依然有加载状态反馈，且节点内容刷新更加果断一致。
- 确认 `pathIdMap` 依然能正确记录各路径对应的 ID，不影响“新建文件夹”功能。
