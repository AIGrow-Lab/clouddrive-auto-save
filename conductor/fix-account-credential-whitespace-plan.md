# 账号凭据空白字符过滤修复计划 (Fix Account Credential Whitespace Plan)

## 背景与目标 (Objective)

当前系统在添加或更新账号时，如果 `Cookie` 或 `AuthToken`（Authorization）的末尾存在多余的换行符，可能导致后续调用底层云盘驱动校验接口时失败（如 HTTP 请求头异常或解析错误）。
目标是在业务数据入库和发起校验前，拦截并剔除这些凭据信息首尾的空白字符与换行符。

## 关键文件 (Key Files & Context)

- `internal/api/router.go`

## 实施步骤 (Implementation Steps)

在 `internal/api/router.go` 文件内：

1. **`createAccount` 接口**：
   在 `c.ShouldBindJSON(&account)` 完成反序列化后，且在写入数据库和校验之前，添加以下清理逻辑：

   ```go
   account.Cookie = strings.TrimSpace(account.Cookie)
   account.AuthToken = strings.TrimSpace(account.AuthToken)
   ```

2. **`updateAccount` 接口**：
   在 `c.ShouldBindJSON(&account)` 完成反序列化后，添加相同的清理逻辑：

   ```go
   account.Cookie = strings.TrimSpace(account.Cookie)
   account.AuthToken = strings.TrimSpace(account.AuthToken)
   ```

*(注：`strings` 包目前已在该文件中导入，无需额外引入。)*

## 验证与测试 (Verification & Testing)

1. 运行 `make check` 确保代码格式和现存测试通过。
2. 通过前端界面（或构造接口请求），尝试添加或更新 `Cookie` 和 `Authorization` 字段末尾或开头故意带上换行符与空格的账号。
3. 验证接口不再报错，校验正常通过，且数据库中最终保存的凭据内容已成功去除了换行符。
