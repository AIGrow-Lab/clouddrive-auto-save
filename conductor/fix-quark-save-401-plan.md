# 修复夸克转存逻辑 401 报错 (Fix Quark Save 401 Error)

## 1. 目标 (Objective)
修复夸克网盘转存过程中持续出现的 401 `require login [guest]` 报错。

## 2. 背景与分析 (Background & Analysis)
通过对比 `quark-auto-save` 的 Python 实现，我发现了几个关键的差异：
1. **参数缺失**：未开启 AppParams 模拟时，基础的请求（如 `getStoken`）没有附带必要的 `pr=ucpro` 和 `fr=pc` 查询参数，这会触发 WAF 拦截或导致生成的 Token 上下文不一致。
2. **App 模拟覆盖不全**：Python 版本中，只要请求 URL 包含 `share`，就会自动升级为移动端 App 请求（替换域名并附加一堆设备参数）。在 Go 代码中，`getStoken` 使用的是 `useAppParams = false`，导致它与后续使用 App 模拟的 `ParseShare` 和 `SaveLink` 环境不一致（Cookie 与无 Cookie 环境的混用）。
3. **转存参数遗漏**：在实际转存 `SaveLink` 和获取详情 `ParseShare` 时，遗漏了 `uc_param_str=""`, `_page=1`, `_size=50`, `__dt` 等官方模拟参数。

## 3. 实施步骤 (Implementation Steps)

### 3.1 增强底层 doRequest
修改 `internal/core/quark/client.go` 中的 `doRequest`：
* 在未启用 `useAppParams` 的 `else` 分支中，自动为 `query` 附加 `pr=ucpro` 和 `fr=pc`，保障所有普通 PC 接口（如目录浏览）的畅通。

### 3.2 统一 share 接口环境
* 将 `getStoken` 的 `doRequest` 调用的 `useAppParams` 参数改为 `true`（或者根据 URL 中是否包含 share 动态判断，这里我们直接给分享相关的接口传 true），使其与 `detail` 和 `save` 处于同一个设备标识和鉴权环境。

### 3.3 补全 SaveLink 和 Detail 请求参数
* 在获取详情 `detailQuery` 中，补全 `_page=1`, `_size=100`, `force=0` 等参数。
* 在转存请求 `saveQuery` 中，补全 `uc_param_str=""` 和随机延迟参数 `__dt`。

## 4. 验证测试
保存修改后重启服务器，重试转存任务，检查 401 是否消失。