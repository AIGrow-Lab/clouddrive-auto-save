# 修复夸克网盘会员等级和容量无法获取的问题计划 (Fix Quark VIP and Capacity Fetching Plan)

## 1. 背景与问题描述
用户反馈在网页端成功获取了夸克网盘的昵称，但会员等级和存储空间仍然显示为空（或为 0 显示为 `-`）。
经过分析，夸克网盘在获取“容量和会员信息”时，我们当前代码调用的是移动端 App 专有的接口 `/1/clouddrive/capacity/growth/info`。该接口强依赖 Cookie 中的 `kps`, `sign`, `vcode` 等参数。
当用户通过 PC 浏览器抓取 Cookie 时，Cookie 中是不包含 `kps` 的，这就导致代码逻辑 `if q.mparam["kps"] != ""` 被直接跳过，从而完全没有去获取容量信息。

## 2. 影响范围
- 后端 Quark 驱动：`internal/core/quark/client.go` 的 `GetInfo` 方法。

## 3. 实施步骤
为了让即使只有基础 Cookie（无 `kps` 参数）的账号也能获取到容量和会员，我们需要补充调用夸克网盘的 PC 端容量接口作为备用或保底方案。

修改 `internal/core/quark/client.go` 中的 `GetInfo` 逻辑：
1. 保留原本的 App `growth/info` 接口调用逻辑，因为该接口能准确返回像 `88VIP`, `SVIP` 等更细分的会员标签。
2. 增加逻辑：如果当前账号的 Cookie 中没有 `kps`（即未调用 App 接口），或者 App 接口调用/解析失败，则**降级回退使用 PC 端网页容量接口** `GET https://drive-pc.quark.cn/1/clouddrive/capacity`。
3. 解析 PC 接口返回的 JSON（包含 `cap_info.total`, `cap_info.used` 以及 `member_type` 等字段）。
4. 对 `member_type` 字段（0：普通用户，1：VIP，2：SVIP）进行宽容类型转换和判断。

## 4. 验证方式
执行上述替换后退出计划模式，并重启 Go 服务器。然后在前端界面点击夸克网盘账号的“校验”按钮，验证即使在缺少 `kps` 参数的 Cookie 情况下，仍能展示正确的“VIP”等级和带有使用比例的存储空间进度条。