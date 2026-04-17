# 跳过同名文件转存设计优化计划 (Skip Duplicate Files Plan)

## 1. Objective (目标)
优化转存任务的执行逻辑：在转存前预先检查目标网盘目录。如果待转存的分享文件中存在与目标目录同名的文件，则直接跳过该文件，不再进行冗余转存，提升转存效率和云盘整洁度。

## 2. Key Files & Context (核心影响文件)
- **核心接口**: `internal/core/drive.go`
  - 在 `CloudDrive` 接口中新增 `PrepareTargetPath` 方法（原先驱动内为私有方法），以便 `Worker` 可以获取目标目录的真实 ID。
- **网盘驱动**: `internal/core/quark/client.go`, `internal/core/cloud139/client.go`
  - 将原私有方法 `prepareTargetPath` 重命名为对外公开的 `PrepareTargetPath` 并调整内部调用。
- **执行引擎**: `internal/core/worker/worker.go`
  - 在遍历过滤待转存文件时，增加目标目录同名检查逻辑，构建去重名单，精确缩减需要转存的文件 ID 列表。

## 3. Implementation Steps (实施步骤)

### Step 3.1: 升级 CloudDrive 接口
1. 打开 `internal/core/drive.go`。
2. 在 `CloudDrive` 接口中追加方法：
   `PrepareTargetPath(ctx context.Context, path string) (string, error)`

### Step 3.2: 调整驱动内部方法权限
1. 打开 `internal/core/quark/client.go`，将 `prepareTargetPath` 重命名为 `PrepareTargetPath`，并修改 `SaveLink` 里的相应调用。
2. 打开 `internal/core/cloud139/client.go`，同样重命名为 `PrepareTargetPath`，修改相应调用。

### Step 3.3: 优化 Worker 执行逻辑 (核心拦截)
1. 打开 `internal/core/worker/worker.go`。
2. 在过滤阶段（`2.2 日期与 ID 过滤`）之前，提前调用：
   `targetID, err := driver.PrepareTargetPath(ctx, task.SavePath)`
3. 调用 `driver.ListFiles(ctx, targetID)` 获取目标目录下已存在的文件列表，并生成基于文件名的 `map[string]bool`。
4. 修改遍历过滤 `for _, f := range files { ... }` 逻辑，增加判断：
   如果 `f.Name` 在前面生成的 `map` 中存在且不为文件夹（或包括同名文件夹），则作为已存在文件跳过，记入跳过统计。
5. 优化完成日志输出：如果在这一步所有文件均被跳过（被判定为已存在），则直接完成任务，显示类似 “No new files to transfer (all files already exist)”。

## 4. Verification & Testing (验证与测试)
- **编译检查**: 运行 `make build` 确保接口变动没有引发语法错误。
- **重复执行测试**: 对于一个已经完成且成功的转存任务，点击“运行”。预期控制台输出全部文件被识别为同名并跳过。
- **部分重名测试**: 向同一个目标目录转存包含新旧文件的分享链接，确认只有新文件产生转存动作。