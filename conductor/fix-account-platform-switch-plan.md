# 修复编辑账号时切换平台导致凭据残留计划 (Fix Account Platform Switch Credential Leak Plan)

## 背景与目标 (Objective)

在前端“账号管理”页面编辑或添加账号时，如果用户在对话框中切换“网盘平台”（例如从夸克切换到移动云盘），原平台填写的 `Cookie` 或 `Authorization` 仍然保留在输入框中。由于不同平台的凭据格式完全不同，这种残留会导致误导或校验失败。
目标是在切换平台时自动清空已填写的凭据信息。

## 关键文件 (Key Files & Context)

- `web/src/views/Accounts.vue`

## 实施步骤 (Implementation Steps)

1. **修改模板**：
   在 `web/src/views/Accounts.vue` 中，找到 `el-radio-group` (绑定 `accountForm.platform` 的部分)，为其添加 `@change` 事件监听器：

   ```html
   <el-radio-group v-model="accountForm.platform" @change="handlePlatformChange">
   ```

2. **添加处理函数**：
   在 `<script setup>` 块中添加 `handlePlatformChange` 函数，用于清空凭据：

   ```javascript
   const handlePlatformChange = () => {
     accountForm.value.cookie = ''
     accountForm.value.auth_token = ''
   }
   ```

## 验证与测试 (Verification & Testing)

1. 启动前端开发服务器 (`make dev-web`)。
2. 打开“账号管理”页面。
3. 点击“添加账号”，先选择“夸克网盘”并输入一些内容到 Cookie 框。
4. 切换到“移动云盘”，验证 Cookie 和 Authorization 框是否已清空。
5. 点击某个已有账号的“编辑”，尝试切换平台，验证同样生效。
