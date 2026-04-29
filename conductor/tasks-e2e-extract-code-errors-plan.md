# 增加提取码错误处理的 E2E 测试覆盖计划

## 目标 (Objective)

在最近优化的网盘提取码错误提示逻辑基础上，增加端到端 (E2E) 测试用例，确保 139 移动云盘和夸克网盘在提取码缺失或错误时，系统能正确触发 `[Fatal]` 级别的报错，并在 UI 上展示正确的中文提示，阻断任务继续执行。

## 具体改动 (Changes)

### 1. 更新 HTTP Mock 数据 (`internal/core/mock_http.go`)

- 在拦截 Quark 分享链接详情和 Token 获取接口 (`drive-pc.quark.cn/1/clouddrive/share/sharepage/...`) 的逻辑中，新增对 `mock_quark_missing_code` 和 `mock_quark_wrong_code` 关键词的拦截：
  - `mock_quark_missing_code` 触发返回 `{"code": 41008, "message": "请先输入提取码"}`。
  - `mock_quark_wrong_code` 触发返回 `{"code": 41007, "message": "提取码错误"}`。
- 在拦截 139 移动云盘获取分享信息接口 (`share-kd-njs.yun.139.com/yun-share/.../getOutLinkInfoV6`) 的逻辑中，新增对 `mock_139_wrong_code` 关键词的拦截：
  - 触发返回 `{"code": "9188", "message": "提取码校验失败"}`。

### 2. 增加 E2E 测试用例 (`e2e/tests/tasks/execute.spec.ts`)

在任务执行相关的测试文件中，增加三个测试用例：

- **139 移动云盘：验证提取码错误或未提供场景**：使用链接 `https://yun.139.com/w/#/share/link/mock_139_wrong_code` 创建任务，触发执行后验证是否抛出 `LINK ERROR`，且提示语为“提取码错误或未提供提取码，请检查后再试。”。
- **夸克网盘：验证未提供提取码场景**：使用链接 `https://pan.quark.cn/s/mock_quark_missing_code` 创建任务，触发执行后验证提示语是否为“当前分享链接需要提取码，请填写提取码 (code: 41008)”。
- **夸克网盘：验证提取码错误场景**：使用链接 `https://pan.quark.cn/s/mock_quark_wrong_code` 创建任务，触发执行后验证提示语是否为“提取码错误，请检查后再试 (code: 41007)”。

## 验证方式 (Verification)

执行 E2E 测试命令 `cd e2e && npx playwright test tests/tasks/execute.spec.ts`，确保所有新增加的测试用例均通过。
