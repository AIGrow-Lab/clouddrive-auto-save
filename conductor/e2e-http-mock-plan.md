# E2E 彻底下沉至 HTTP Mock 层计划 (E2E HTTP-Level Mock Plan)

**Goal:** 将 E2E 测试环境的 Mock 拦截层从 Service 层（`MockDriver`）下沉至 HTTP 传输层（`http.RoundTripper`），强制底层网盘驱动（`Quark`, `Cloud139`）去解析真实的模拟 JSON 数据，从而建立起防范 JSON 解析和参数拼接错误的绝对防线。

---

### Task 1: 定义全局可被替换的 HTTP 传输通道

**Files:**
- Modify: `internal/core/drive.go`

**Changes:**
1. 新增导出的全局变量 `var HTTPTransport http.RoundTripper = http.DefaultTransport`。

### Task 2: 改造底层驱动以支持 HTTP 注入

**Files:**
- Modify: `internal/core/quark/client.go`
- Modify: `internal/core/cloud139/client.go`

**Changes:**
1. 在两个驱动的初始化函数 (`NewQuark`, `NewCloud139`) 中，将默认的 `http.Client` 修改为：
   `client: &http.Client{Timeout: 30 * time.Second, Transport: HTTPTransport}`

### Task 3: 编写 HTTP Mock 拦截器

**Files:**
- Create: `internal/core/mock_http.go`

**Changes:**
1. 实现一个自定义的 `mockTransport` 结构体，实现 `RoundTrip(*http.Request)` 方法。
2. 在该方法中，根据请求的 `URL` 路径拦截请求，直接返回包含原生 JSON 的 `http.Response`。
   - 必须模拟夸克接口：账号/容量 (`/account/info`, `/1/clouddrive/member`)、分享解析 (`/sharepage/detail`)、预检 (`/file/sort`)、新建目录 (`/file`)、转存提交 (`/sharepage/save` 返回 `task_id`)、轮询任务 (`/task` 返回 `status: 2`)。
   - 必须模拟移动云盘接口：账号 (`/user/getUser`)、容量 (`/getPersonalDiskInfo`)。
3. 导出一个 `SetupE2EHTTPMock` 方法，调用时将 `core.HTTPTransport` 赋值为此拦截器。

### Task 4: 切换后端的 E2E 启动逻辑

**Files:**
- Modify: `cmd/server/main.go`

**Changes:**
1. 移除旧的 Service Mock 注入。
2. 调用 `core.SetupE2EHTTPMock()`。
3. 将测试账号的 `Platform` 改回真实的 `"139"` 和 `"quark"`。

### Task 5: 验证并提交

- 运行 Playwright E2E 测试，确认功能。
- （先不执行本地提交）
