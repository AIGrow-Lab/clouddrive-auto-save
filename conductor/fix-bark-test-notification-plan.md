# 修复 Bark 测试通知逻辑 (Fix Bark Test Notification)

## 1. 背景与问题

用户反馈在前端“系统设置”页面填写 Bark 配置并点击“发送测试消息”后，手机无法收到推送。分析发现原因如下：

- 当前后端测试接口复用了 `notify.SendBark`，该函数会去数据库读取配置，并检查 `bark_enabled` 开关是否开启。
- 如果用户在前端填写了配置但尚未保存，或者尚未打开开关，测试消息将被跳过。

## 2. 修复方案

### 后端修复 (`internal/core/notify/bark.go` & `internal/api/router.go`)

1. **拆分推送函数**: 在 `notify` 包中提取一个底层函数 `SendBarkDirect(server, key, title, body string) error`。这个函数不读取数据库，不检查开关，直接使用传入的参数发送 HTTP 请求。
2. **修改原有函数**: 原有的 `SendBark` 函数仍然负责读取数据库配置并进行开关判断，但在最终发送时调用新提取的 `SendBarkDirect`。
3. **重构测试接口**: 更新 `/api/settings/test_bark` 路由。它应该接收包含 `server` 和 `key` 的 JSON 请求体，并直接调用 `notify.SendBarkDirect`，从而允许前端验证未保存的配置。

### 前端修复 (`web/src/api/task.js` & `web/src/views/Settings.vue`)

1. **更新 API 定义**: 将 `testBark()` 修改为 `testBark(data)`，使其能够发送 POST 实体数据。
2. **更新视图逻辑**: 在 `Settings.vue` 中，当用户点击“发送测试消息”时，收集当前表单中填写的 `bark_server` 和 `bark_device_key`，作为参数传递给 `testBark()` API。

## 3. 验证步骤

- 启动前后端服务。
- 进入“系统设置”页面。
- 保持“Bark 消息推送”开关**关闭**。
- 填写有效的 Bark Server 和 Device Key。
- 点击“发送测试消息”，观察是否提示成功，并检查手机是否收到推送。
- 清空 Device Key，点击测试，验证是否提示错误。
