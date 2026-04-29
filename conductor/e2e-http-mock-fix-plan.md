# E2E HTTP Mock 修复计划

> **给智能体的指示:** 推荐使用 subagent-driven-development 或 executing-plans 技能来逐项执行本计划。进度通过勾选框 (`- [ ]`) 进行追踪。

**目标:** 修复 `make e2e-test` 失败的问题。通过调整 HTTP Mock 响应数据以匹配最新版 Quark 和 139 网盘客户端的解析逻辑，同时更新 E2E 测试脚本以提供合法的 Quark Cookie。

**架构说明:** 修改 E2E 测试代码绕过客户端对 Quark 的 `__uid=` Cookie 强校验；同时重构 `internal/core/mock_http.go`，精准返回 `quark/client.go` 和 `cloud139/client.go` 当前所期望解析的 JSON 结构。

**技术栈:** Go, Playwright, HTTP Mocking

---

## 任务 1: 更新 E2E 测试中的 Quark Cookie

**涉及文件:**

- 修改: `e2e/tests/core.spec.ts:25-27`

- [ ] **步骤 1: 修改 Playwright 测试中的 Quark Cookie 输入值**

```typescript
    // 4. 交互：添加 Quark 账号
    await page.getByRole('button', { name: '添加账号' }).click();
    await page.getByText('夸克网盘').click();
    await page.getByLabel('Cookie 全量字符串').fill('__uid=mock; mock_cookie');
```

## 任务 2: 重构 Quark 的 HTTP Mock 响应

**涉及文件:**

- 修改: `internal/core/mock_http.go`

- [ ] **步骤 1: 更新 Quark 分享详情接口响应**

在 `internal/core/mock_http.go` 中，更新 `sharepage/detail` 路由的返回值，以符合客户端解析 `dir` (布尔值) 和 `share_fid_token` 的期望。

```go
 } else if strings.Contains(url, "drive-pc.quark.cn/1/clouddrive/share/sharepage/detail") {
  // 模拟返回文件列表
  respBody = `{"code": 0, "data": {"list": [{"fid": "file1", "file_name": "[2024.04.20] E2E测试电影.mp4", "size": 1024, "updated_at": 1612345678000, "dir": false, "share_fid_token": "mock_token_1"}, {"fid": "file2", "file_name": "readme.txt", "size": 100, "updated_at": 1612345679000, "dir": false, "share_fid_token": "mock_token_2"}]}}`
```

## 任务 3: 重构 139 网盘的 HTTP Mock 响应

**涉及文件:**

- 修改: `internal/core/mock_http.go`

- [ ] **步骤 1: 更新 139 获取用户信息接口响应**

为 `getUser` 的响应补充 `"code": "0000"`。

```go
 // 2. 模拟 139 相关接口
 if strings.Contains(url, "user-njs.yun.139.com/user/getUser") {
  respBody = `{"code": "0000", "success": true, "data": {"auditNickName": "E2E移动云盘用户", "userName": "E2E移动云盘用户", "userDomainId": "mock_domain", "loginName": "13800000000"}}`
```

- [ ] **步骤 2: 更新 139 获取容量与权益接口响应**

为获取容量和权益查询接口补充 `"code": "0"`。

```go
 } else if strings.Contains(url, "user-njs.yun.139.com/user/disk/getPersonalDiskInfo") || strings.Contains(url, "user-njs.yun.139.com/user/disk/getFamilyDiskInfo") {
  respBody = `{"code": "0", "success": true, "data": {"diskSize": "1048576", "freeDiskSize": "524288"}}`
 } else if strings.Contains(url, "yun.139.com/orchestration/group-rebuild/member/v1.0/queryUserBenefits") {
  respBody = `{"code": "0", "success": true, "data": {"userSubMemberList": [{"memberLvName": "黄金会员"}]}}`
```

- [ ] **步骤 3: 更新 139 获取文件列表接口响应**

将 `nodeList` 变更为 `items`，这是目前客户端期望的文件列表字段。

```go
 } else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/list") {
  respBody = `{"code": "0", "success": true, "data": {"items": []}}`
```

- [ ] **步骤 4: 更新 139 更新文件名称接口响应**

为接口增加 `"code": "0"`。

```go
 } else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/update") {
  respBody = `{"code": "0", "success": true}`
```

- [ ] **步骤 5: 更新 139 创建文件夹接口响应**

移除外层的 `node` 对象，直接平铺 `fileId` 和 `fileName`。

```go
 } else if strings.Contains(url, "personal-kd-njs.yun.139.com/hcy/file/create") {
  respBody = `{"code": "0", "success": true, "data": {"fileId": "mock_dir_139", "fileName": "mock_dir"}}`
 }
```

## 任务 4: 运行 E2E 测试进行验证

**涉及文件:**

- 无

- [ ] **步骤 1: 运行命令确认修复成功**

```bash
make e2e-test
```

预期结果: 测试无报错并成功通过。
