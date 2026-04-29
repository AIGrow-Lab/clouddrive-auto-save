# TDD 单元测试防护网实施计划 (TDD Protection Plan)

**Goal:** 落实高意图密度的单元测试防护网，防止后续使用 AI 修改代码时破坏现有的核心转存逻辑（起始点截断、正则过滤、预计算去重）和 API 契约。

---

## Task 1: 升级 Mock 驱动为 Spy Mock

**Files:**

- Modify: `internal/core/worker/mock_test.go`

**Changes:**

1. 为 `MockDriver` 增加记录状态的字段，如 `SavedFileIDs []string` 和 `SaveLinkCalls int`。
2. 修改 `SaveLink` 方法，将传入的 `fileIDs` 追加到 `SavedFileIDs` 中，并递增调用计数。这使得测试用例能够“监视” Worker 到底向驱动提交了哪些文件。

## Task 2: 补充核心转存逻辑防护测试

**Files:**

- Modify: `internal/core/worker/worker_test.go`

**Changes:**

1. **新增 `TestManager_Execute_StartFileFilter`**: 验证当设置 `StartFileID` 时，只有该文件及其更新时间更晚的文件被转存，旧文件被安全忽略。
2. **新增 `TestManager_Execute_RegexFilter`**: 验证当设置 `Pattern` 时，只有匹配该正则的文件 ID 被传递给驱动的 `SaveLink` 方法。
3. **新增 `TestManager_Execute_Deduplication_With_Renamer`**: 验证结合重命名规则后，如果预计算出的新名字已经在目标目录存在，则该文件不会被转存。

## Task 3: 补充 API 契约防护测试

**Files:**

- Create: `internal/api/task_preview_test.go`

**Changes:**

1. 编写集成测试，启动 Gin 引擎调用 `POST /api/tasks/preview`。
2. 构造包含空正则、匹配正则和不匹配正则的请求 Payload。
3. **严格断言 JSON 响应**: 确保返回的 JSON 数组中，每个对象都包含 `original_name`, `new_name`, `matched`, `is_filtered` 字段，且 `matched` 的布尔值符合预期。防止未来因修改结构体标签导致前端字段缺失。

## Task 4: 运行检查并自动提交

- 执行 `make check` 确保所有新增测试通过。
- 提交代码到本地仓库（不推送）。
