# 数据库设计说明 (Database Schema)

本项目采用 **SQLite** 作为存储引擎，利用 **GORM** 实现对象关系映射。

为了实现零依赖的跨平台编译与发布，系统采用了 **CGO-free** 的纯 Go 驱动 [github.com/glebarez/sqlite](https://github.com/glebarez/sqlite)，这使得程序可以在没有 C 编译器（如 gcc）的环境下直接运行。

---

## 1. 账号表 (`accounts`)

存储云盘账号的认证信息、容量快照及状态。

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | uint (PK) | 账号唯一标识 |
| `platform` | string | `139` 或 `quark` |
| `nickname` | string | 云盘真实昵称 |
| `account_name` | string | 用户备注名/登录名 |
| `cookie` | text | 认证 Cookie 串 |
| `auth_token` | text | (仅 139) 抓包获取的 Authorization Token |
| `status` | int | 1: 正常, 0: 失效 |
| `vip_name` | string | 会员等级名称 |
| `capacity_used`| int64 | 已用容量 (Bytes) |
| `capacity_total`| int64 | 总容量 (Bytes) |
| `last_check` | datetime | 最后一次校验时间 |

---

## 2. 任务表 (`tasks`)

核心业务表，存储转存任务的配置、调度逻辑及持久化实时进度。

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | uint (PK) | 任务唯一标识 |
| `account_id` | uint (FK) | 关联的账号 ID |
| `name` | string | 任务名称 |
| `share_url` | string | 分享链接 |
| `extract_code` | string | 提取码 |
| `save_path` | string | 保存路径 |
| `pattern` | string | 重命名正则表达式 |
| `replacement` | string | 替换模板/魔法变量 |
| `share_parent_id` | string | **[新]** 139 分享链接的目录 ID，用于指定浏览子目录时的起始位置 |
| `schedule_mode`| string | `global` (跟随全局), `custom` (自定义), `off` (手动) |
| `cron` | string | 当模式为 custom 时的标准 6 位 Cron 表达式 |
| `status` | string | `pending`, `running`, `success`, `failed` |
| `percent` | int | **[持久化]** 当前执行进度百分比 (0-100) |
| `stage` | string | **[持久化]** 阶段标识: `Parsing`, `Checking`, `Saving`, `Renaming`, `Success`, `Failed` |
| `last_run` | datetime | 最后一次开始运行的时间 |
| `message` | text | 最后运行的结果摘要或 `[Fatal]` 错误信息 |

---

## 3. 常用目录表 (`common_folders`)

记录用户在文件夹选择器中点击过的或收藏的路径，实现路径历史记忆。

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | uint (PK) | 唯一标识 |
| `account_id` | uint | 所属账号 |
| `path` | string | 完整绝对路径 |
| `name` | string | 文件夹显示名称 |

---

## 4. 系统设置表 (`settings`)

用于持久化系统的全局配置，如全局调度开关。

| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `key` | string (PK) | 配置项键名 |
| `value` | text | 配置内容（如 `global_schedule_enabled`: "true"） |

---

## 5. 核心状态枚举说明

### 任务状态 (`Status`)

- **`pending`**: 待命中，等待调度。
- **`running`**: 正在执行中。
- **`success`**: 上次执行成功。
- **`failed`**: 上次执行遇到错误（包含超时或逻辑错误）。

### 调度模式 (`ScheduleMode`)

- **`global`**: 受全局开关控制，统一频率执行。
- **`custom`**: 忽略全局开关，按自身 `cron` 频率独立执行。
- **`off`**: 完全关闭自动调度，仅支持手动触发。
