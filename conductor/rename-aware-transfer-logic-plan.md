# 实现重命名感知转存与自动重命名计划 (Rename-Aware Transfer Plan)

## 1. Objective (目标)
将重命名引擎整合进任务执行核心流程（Worker）。确保任务运行时：1. 基于重命名后的文件名进行同名跳过；2. 转存成功后自动执行网盘内的重命名操作。

## 2. Key Files & Context (核心影响文件)
- `internal/core/drive.go`: 升级 `CloudDrive` 接口，增加 `RenameFile` 方法。
- `internal/core/quark/client.go`: 实现夸克网盘的重命名 API。
- `internal/core/cloud139/client.go`: 实现移动云盘的重命名 API。
- `internal/core/worker/worker.go`: 整合 `renamer` 引擎，实现去重比对与转存后的重命名触发。

## 3. Implementation Steps (实施步骤)

### Step 3.1: 接口与驱动升级
1. **修改 `internal/core/drive.go`**:
   - 在 `CloudDrive` 中添加：`RenameFile(ctx context.Context, fileID, newName string) error`。
2. **实现驱动方法**:
   - **Quark**: 调用 `POST /1/clouddrive/file` (或专用的 rename 接口) 修改文件名。
   - **139**: 调用移动云盘的 `batchUpdate` 或相应重命名接口。

### Step 3.2: Worker 逻辑重构 (核心)
1. **引入 Renamer**: 在 `execute` 方法开始处初始化重命名处理器。
2. **重命名感知比对**:
   - 修改 `existingMap` 的检查逻辑。
   - 对每个待转存文件 `f`，先计算其预览名 `targetName`。
   - 检查 `existingMap[targetName]` 而非 `existingMap[f.Name]`。
3. **批量转存与逐个重命名**:
   - 维持现有的 `SaveLink` 批量转存（效率最高）。
   - **重命名同步**: 转存完成后，由于我们需要将转存后的文件重命名，这里存在一个技术细节：批量转存后的文件 ID 通常需要通过一次 `ListFiles` 重新获取，或者通过解析 `SaveLink` 的响应（如果包含 FID 映射）。
   - *优化方案*：转存完成后，再次拉取目标目录，匹配刚刚转存的文件并执行 `RenameFile`。

## 4. Verification & Testing (验证与测试)
- **去重验证**: 开启正则重命名，确认如果目标目录已存在“新名字”的文件，任务会正确跳过。
- **重命名验证**: 确认转存成功后，网盘中的文件确实被改成了正则定义的名称。
- **编译测试**: 确保全链路代码无语法错误。
