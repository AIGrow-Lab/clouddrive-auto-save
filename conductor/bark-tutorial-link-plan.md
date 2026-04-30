# Bark 教程链接添加计划

## Objective (目标)
在设置页面的 Bark 消息推送标题旁添加查看教程的链接，方便用户快速访问 Bark 的官方教程。

## Key Files & Context (涉及文件)
- `web/src/views/Settings.vue`

## Implementation Steps (实施步骤)
1. 修改 `web/src/views/Settings.vue` 中的 Bark 消息推送卡片头部 (`<div class="header-title">`)。
2. 在 `<span>Bark 消息推送</span>` 旁边添加一个 `<el-link>` 组件：
   - 链接地址 (`href`) 为 `https://bark.day.app/`
   - 设置新标签页打开 (`target="_blank"`)
   - 移除默认下划线 (`:underline="false"`)
   - 增加适当的左边距 (`style="margin-left: 8px; font-size: 13px;"`) 调整排版
   - 链接文本为“查看教程”

## Verification & Testing (验证与测试)
1. 启动前端开发服务器 (`make dev-web`)。
2. 访问设置页面。
3. 检查 Bark 推送卡片标题旁边是否正确渲染了“查看教程”的超链接，且样式协调。
4. 点击该链接，验证是否能正常在新标签页打开 `https://bark.day.app/`。