# 修复夸克网盘 API 401 拦截问题 (Fix Quark 401 WAF Issue)

## 1. 目标 (Objective)
修复夸克网盘在执行 `ListFiles` 和 `CreateFolder` 等操作时遭遇 401 状态码和加密报错的问题。

## 2. 背景与原因 (Background & Cause)
在尝试通过前端编辑任务或者选择目录时，后端的 `getAccountFolders` 接口会调用夸克驱动的 `ListFiles` 方法。夸克网盘的接口具备 WAF (防爬虫) 机制，必须在每一次请求中都带上客户端标识参数 `pr=ucpro` 和 `fr=pc`。
如果缺少这些参数，服务器不会返回常规的 JSON 错误，而是返回 HTTP 401 和一串加密字符串（如 `AATFelIx...`），导致后端尝试将其作为 JSON 解析时触发 `invalid character 'A'` 报错。
目前代码中 `ListFiles` 和 `CreateFolder` 遗漏了这几个必需的 URL 查询参数。

## 3. 实施步骤 (Implementation Steps)

### 3.1 补充公共查询参数
1. 修改 `internal/core/quark/client.go` 中的 `ListFiles` 方法：
    * 向 `query` (类型为 `url.Values`) 中追加 `query.Set("pr", "ucpro")`。
    * 向 `query` 中追加 `query.Set("fr", "pc")`。

2. 修改 `internal/core/quark/client.go` 中的 `CreateFolder` 方法：
    * 新建 `query := url.Values{}`。
    * 增加 `query.Set("pr", "ucpro")` 和 `query.Set("fr", "pc")`。
    * 在调用 `doRequest` 时，将原本的 `nil` 参数替换为新创建的 `query`。

## 4. 验证与测试 (Verification & Testing)
1. 重新编译并运行后端服务。
2. 在前端访问任务编辑对话框，或者尝试在保存路径下拉框中请求夸克账号的目录树。
3. 验证网络请求不再返回 500/401 错误，且能够成功拉取到 JSON 格式的真实目录结构。