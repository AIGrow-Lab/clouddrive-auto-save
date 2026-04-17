# 修复夸克网盘重命名接口请求失败计划 (Fix Quark Rename API Plan)

## 1. Objective (目标)
修复夸克网盘在转存文件后的“重命名操作”阶段持续报错 `400 [同名冲突]` 的严重 Bug。

## 2. Key Files & Context (核心影响文件)
- `internal/core/quark/client.go`：包含 Quark 云盘的所有 API 封装。其中 `RenameFile` 方法目前的 URL 路径有误。

## 3. Root Cause Analysis (根本原因分析)
经过分析原版 Python 脚本 (`quark_auto_save.py`)，我们发现：
- 正确的 Quark 重命名 API 路径为：`/1/clouddrive/file/rename`，并需要附带 Query 参数 `?pr=ucpro&fr=pc`。
- 我们现有的 Go 代码中，API 路径错误地写成了 `/1/clouddrive/file`，且缺少了 Query 参数。这个路径在 Quark 的系统中是用于**新建文件夹或保存文件**的操作。
- 因为传入的参数是一致的，系统误以为我们要在此路径新建一个名为 `193.mp4` 的文件，但由于底层的某种锁机制或该目录本身的属性，触发了 `file is doloading[同名冲突]` 异常。

## 4. Implementation Steps (实施步骤)
1. 打开 `internal/core/quark/client.go`。
2. 定位到 `RenameFile` 方法。
3. 将请求 URL 从 `BaseURL + "/1/clouddrive/file"` 修改为 `BaseURL + "/1/clouddrive/file/rename"`。
4. 补充 URL Query 参数（`pr=ucpro`, `fr=pc`），并将其正确传入 `q.doRequest` 函数。

## 5. Verification & Testing (验证与测试)
1. 编译并启动服务。
2. 触发一次 Quark 网盘的转存重命名任务。
3. 查看日志中是否不再出现 `StatusCode=400`，并打印出成功的重命名记录。
