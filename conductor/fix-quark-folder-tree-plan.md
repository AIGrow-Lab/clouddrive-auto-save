# 修复夸克网盘目录树加载问题 (Fix Quark Folder Tree Loading)

## 1. 目标 (Objective)
修复夸克网盘在任务管理模块中无法正确获取和展示目录树的问题。通过对齐 `quark-auto-save` 项目的 API 调用逻辑，确保 FID (File ID) 的正确传递和根目录的准确识别。

## 2. 涉及的关键文件 (Key Files & Context)
*   **驱动接口:** `internal/core/drive.go`
*   **后端 API:** `internal/api/folder.go`
*   **139 驱动:** `internal/core/cloud139/client.go`
*   **夸克驱动:** `internal/core/quark/client.go`
*   **前端视图:** `web/src/views/Tasks.vue`

## 3. 实施步骤 (Implementation Steps)

### 3.1 统一驱动接口定义
1.  修改 `internal/core/drive.go` 中的 `CloudDrive` 接口：
    *   将 `ListFiles` 的参数 `parentPath string` 改为 `parentID string`。
    *   将 `CreateFolder` 的参数 `parentPath string` 改为 `parentID string`。

### 3.2 更新后端 API 逻辑
1.  修改 `internal/api/folder.go`：
    *   将 `getAccountFolders` 中的 Query 参数从 `parent_path` 改为 `parent_id`。
    *   默认值从 `/` 改为空字符串 `""`（由驱动内部处理根目录映射）。
    *   在返回 `FolderItem` 时，`ID` 字段明确对应 `f.ID`。

### 3.3 优化夸克驱动实现
1.  修改 `internal/core/quark/client.go`：
    *   在 `ListFiles` 中，完善查询参数，增加 `_fetch_total: 1` 和 `fetch_risk_file_name: 1` 等，对齐 `quark-auto-save` 的稳定调用。
    *   确保根目录映射逻辑稳健：`if parentID == "" || parentID == "/" { parentID = "0" }`。

### 3.4 适配 139 驱动
1.  由于接口变更，同步修改 `internal/core/cloud139/client.go` 中的方法签名，将 `parentPath` 重命名为 `parentID`（保持原有路径作为 ID 的逻辑）。

### 3.5 前端调用适配
1.  修改 `web/src/views/Tasks.vue` 中的 `loadFolders` 方法：
    *   将请求参数名从 `parent_path` 修改为 `parent_id`。
    *   初始加载根目录时，传递空字符串。

## 4. 验证与测试 (Verification & Testing)
1.  重新编译运行后端服务。
2.  在前端“创建任务”对话框中，选择一个夸克账号。
3.  点击“保存路径”的 TreeSelect，验证是否能加载出根目录列表。
4.  点击子目录，验证是否能正确异步加载二级目录。