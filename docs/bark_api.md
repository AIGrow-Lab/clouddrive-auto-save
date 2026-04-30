# Bark 消息推送 API 使用说明

本文档基于 [Bark 官方教程](https://bark.day.app/#/tutorial) 整理，详细说明了如何通过 Bark API 发送推送通知。

## 1. 基础信息
- **服务地址**: `https://api.day.app` (官方默认) 或你的私有部署地址。
- **认证方式**: 使用 App 首页显示的 `device_key`。

## 2. 快捷请求 (GET)
最简单的调用方式，直接将参数拼接到 URL 中。

### URL 结构
- `https://api.day.app/{device_key}/{推送内容}`
- `https://api.day.app/{device_key}/{推送标题}/{推送内容}`
- `https://api.day.app/{device_key}/{标题}/{副标题}/{内容}`

### 示例
```bash
curl https://api.day.app/your_key/这是一条测试推送
```

## 3. 标准请求 (POST) - 推荐
对于复杂参数或较长内容，建议使用 POST 请求。

- **Endpoint**: `https://api.day.app/push`
- **Method**: `POST`
- **Content-Type**: `application/json`

### 请求体参数 (JSON)

| 参数名 | 类型 | 必填 | 说明 |
| :--- | :--- | :---: | :--- |
| `device_key` | string | 是 | App 首页显示的 Key |
| `body` | string | 是 | 推送正文内容 |
| `title` | string | 否 | 推送标题 |
| `subtitle` | string | 否 | 推送副标题 |
| `level` | string | 否 | 推送级别：`active` (默认), `timeSensitive` (时效性), `passive` (静默), `critical` (告警) |
| `badge` | int | 否 | App 图标角标数字 |
| `sound` | string | 否 | 推送铃声 (如 `minuet.caf`)，可在 App 设置中查看列表 |
| `icon` | string | 否 | 自定义图标 URL (仅限 iOS 15+) |
| `group` | string | 否 | 对推送进行分组，方便在通知中心分类 |
| `url` | string | 否 | 点击推送后跳转的 URL |
| `isArchive` | int | 否 | 值为 `1` 时自动保存到 App 历史记录 |
| `copy` | string | 否 | 指定点击推送时复制到剪贴板的内容 |
| `autoCopy` | int | 否 | 值为 `1` 时，收到推送自动复制 `copy` 字段的内容 |

### POST 示例
```bash
curl -X POST "https://api.day.app/push" \
     -H "Content-Type: application/json; charset=utf-8" \
     -d '{
           "device_key": "your_device_key",
           "title": "UCAS 系统通知",
           "body": "任务 [每日转存] 已成功完成",
           "group": "UCAS",
           "sound": "birdsong.caf",
           "isArchive": 1
         }'
```

## 4. 关键特性说明
1. **时效性通知 (Time Sensitive)**: 设置 `level=timeSensitive`，可在专注模式下正常弹出。
2. **告警通知 (Critical Alert)**: 设置 `level=critical`，即使手机处于静音或勿扰模式，也会以最大音量播放声音（需在 App 中开启相应权限）。
3. **自动复制**: 配合 `copy` 和 `autoCopy=1` 字段，非常适合发送验证码等需要快速提取的信息。
4. **私有化部署**: Bark 支持 Docker 部署，若使用私有服务器，只需将请求域名更换为私有地址。

## 5. 项目集成参考
在本项目中，Bark 相关的逻辑主要位于：
- **后端逻辑**: `internal/core/notify/bark.go`
- **设置界面**: `web/src/views/Settings.vue`
