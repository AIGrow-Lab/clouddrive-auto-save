# 修复任务保存路径与文件夹 ID 混淆的 Bug (Fix Tree Select Path vs ID Bug)

## 1. 目标 (Objective)
修复在前端选择目录后，`save_path` 输入框显示乱码 (UUID) 且后端执行时将该 UUID 作为文件夹名称创建的 Bug。

## 2. 背景与原因 (Background & Cause)
目前 `Tasks.vue` 中的 `<el-tree-select>` 组件使用 `value: 'id'` 进行数据绑定。由于夸克网盘的目录 ID 是一串类似 UUID 的 Hex 字符串 (如 `b996...`)，用户选择目录后，表单的 `save_path` 实际上被赋值成了这个 ID。
而底层的 `prepareTargetPath` 逻辑是基于真实的绝对路径（如 `/资源/电影`）来递归解析和创建文件夹的。当后端收到保存路径为 `b996...` 时，它会误认为用户想要在根目录下创建一个名字叫 `b996...` 的文件夹，从而导致了该错误。

## 3. 实施步骤 (Implementation Steps)

### 3.1 修改后端目录结构 (`internal/api/folder.go`)
1. 在 `FolderItem` 结构体中新增 `Path string` 字段。
2. `getAccountFolders` 接收新的 Query 参数 `parent_path`（默认 `/`）。
3. 在遍历组装子目录时，基于 `parent_path` 拼接出完整的绝对路径并赋值给 `FolderItem.Path`。
4. 在 `createAccountFolder` 接口中接收 `parent_path` 并返回构建好的新 `Path`。

### 3.2 修改前端 API 声明 (`web/src/api/account.js`)
1. 将 `getFolders` 的参数从 `(accountId, parentID)` 修改为 `(accountId, parentID, parentPath)`，并在请求中携带 `parent_path`。

### 3.3 修改前端交互逻辑 (`web/src/views/Tasks.vue`)
1. 修改 `<el-tree-select>` 的 `:props` 属性，将 `value: 'id'` 更改为 `value: 'path'`。这将使得用户的 `form.save_path` 再次变回人类可读的真实绝对路径。
2. 增加响应式字典 `pathIdMap`，在懒加载目录时将 `path` 映射到 `id`。
3. 修改 `loadFolders`：获取当前节点的 `path`，连同 `id` 一起发送给后端，并更新 `pathIdMap`。
4. 优化 `handleCreateFolder`：取消立即请求后端的强绑定，而是采用追加文本路径 (`form.save_path += "/" + name`) 的方式。后端 `prepareTargetPath` 会在任务运行时自动递归创建新文件夹，完美兼顾体验和底层逻辑。

## 4. 验证与测试 (Verification & Testing)
1. 重启服务，进入前端任务创建弹窗。
2. 展开夸克网盘目录并选择一个子目录。
3. 检查输入框内显示的应为完整的绝对路径 (如 `/云盘/视频`)，不再是乱码。
4. 点击新建目录，输入框能自动追加新目录名称。
5. 提交任务并运行，前往夸克网盘检查是否按真实名称创建/保存了文件，而不再出现乱码文件夹。