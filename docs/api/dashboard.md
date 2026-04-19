# 仪表盘统计 (Dashboard)

## 1. 获取系统运行统计
获取当前系统的运行状态汇总数据，包括任务统计、容量汇总和近期动态。

- **URL**: `/dashboard/stats`
- **Method**: `GET`
- **Response**: `Object`

### 响应字段说明
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `active_accounts` | int64 | 当前状态正常的云盘账号总数 |
| `capacity_used` | int64 | 所有活跃账号已使用的云盘空间总和 (Bytes) |
| `scheduled_tasks` | int64 | 当前已规划自动调度的任务总数 |
| `today_completed` | int64 | 今日（零点起）成功执行完成的任务次数 |
| `running_tasks_list`| Array | **[新]** 当前正在执行中、最近报错或最近 8 秒内成功且未被清空的活跃任务列表 |
| `recent_activities`| Array | 最近执行过的 15 条有消息记录的任务历史列表 |

---

## 2. 实时系统总线 (SSE)
建立 Server-Sent Events 连接，实时获取系统日志、任务进度及全量同步事件。

- **URL**: `/dashboard/logs`
- **Method**: `GET`
- **Headers**: `Accept: text/event-stream`

### 协议格式说明
后端通过 SSE 推送的消息包含以下三种类型：

1.  **分级日志**: 带有等级前缀的文本日志（受 `LOG_LEVEL` 环境变量过滤）。
    *   格式: `[LEVEL] Message` (例如: `[INFO] Worker 启动`)
    *   **注意**: 原始网盘 API 响应已被降级为 `DEBUG` 等级。
2.  **任务进度 (`[PROGRESS:...]`)**:
    *   格式: `[PROGRESS:TaskID:Percent:Stage:Message]`
    *   作用: 实时驱动前端仪表盘的“任务执行监控”条。
3.  **系统事件 (`[EVENT:...]`)**:
    *   格式: `[EVENT:{"type":"task_update", "task":{...}}]`
    *   作用: 实现后端驱动的 UI 同步（如：任务列表字段变更、任务删除）。
    *   **过滤逻辑**: 该类纯数据消息**不再存入历史记录**，仅在连接建立后实时广播，以保持日志流的整洁。

---

## 3. 获取近期日志历史
获取内存中缓存的最近若干条日志记录。

- **URL**: `/dashboard/logs/recent`
- **Method**: `GET`
- **Response**: `Array<string>`
- **说明**: 历史记录中仅包含文本类日志，已自动过滤掉 `[EVENT:...]` 格式的纯数据事件。

---

## 4. 清空系统日志与状态
清空后端内存中缓存的所有日志记录，并同步重置所有任务的运行摘要。

- **URL**: `/dashboard/logs/recent`
- **Method**: `DELETE`
- **Response**: `{"message": "logs and task summaries cleared"}`
- **Side Effect**: 此操作会同时清空数据库中所有任务的 `message` 和 `stage` 字段，使仪表盘视图回归纯净状态。
