# 修复分享链接跳转缺失提取码的问题

## 目标 (Objective)

修复在新建标签页打开网盘分享链接时，未能携带提取码导致需手动填写的问题。根据 API 文档对各网盘提取码参数的定义，分别对夸克网盘和移动云盘采用不同的 URL 参数拼接规则。针对移动云盘网页端不支持参数识别的缺陷，增加自动复制提取码至剪贴板的功能。

## 具体改动 (Changes)

1. **修改模板传参**
   在 `web/src/views/Tasks.vue` 中，找到 `分享链接` 的输入框附属按钮 `@click="openExternalLink(form.share_url)"`，修改为同时传入提取码：`@click="openExternalLink(form.share_url, form.extract_code)"`。

2. **区分网盘拼接 URL 参数与自动复制剪贴板**
   修改 `openExternalLink` 函数。根据 `docs/cloud_drive_apis.md` 的接口规范判定网盘类型并拼接不同参数名：
   - **移动云盘 (139)**：API 接口定义中，其提取码参数在多个场景中体现为 `passwd` 或是直接携带 `pwd`。实际上移动云盘对拼接提取码到 URL 以供浏览器自动识别并不提供官方直接支持参数，但业界/脚本常使用 `pwd`。考虑到后端 API 抓取通常读取 `passwd` 和 `pwd`（见 `internal/core/cloud139/client.go`），我们在外部 URL 拼接上统一使用更被客户端环境广泛兼容的 `pwd=提取码`。
   - **夸克网盘 (Quark)**：API 接口定义中提取码在授权接口中作为 `passcode` 使用。对于 URL 分享跳转，夸克官方 Web 版通常识别 `pwd=提取码`。我们将统一向其追加 `pwd=提取码`。

   通过 URL 对象安全拼接。如果前端输入的是有效的 URL，则使用 `url.searchParams.set('pwd', extractCode)`；如果无法解析，使用简单的字符串拼接 fallback。

   **新增复制剪贴板逻辑：** 在跳转新标签页前，利用 `navigator.clipboard.writeText(extractCode)` 将提取码复制到剪贴板，并使用 `ElMessage.success` 提示用户“提取码已复制，请在新页面粘贴”，以提升未安装网盘脚本插件时的用户体验。

## 验证方式 (Verification)

1. 在前端页面输入带有提取码的 139 分享链接和提取码，点击外跳，检查新开页面的 URL 是否携带 `pwd=提取码`，同时观察是否出现提取码复制成功的提示。
2. 在前端页面输入夸克分享链接和提取码，点击外跳，检查新开页面的 URL 是否携带 `pwd=提取码`，并确认剪贴板是否已写入提取码。
