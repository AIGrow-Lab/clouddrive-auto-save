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
| `recent_activities`| Array | 最近执行过的 5 条任务记录列表 |

### 响应示例
```json
{
  "active_accounts": 2,
  "capacity_used": 103291877039,
  "scheduled_tasks": 3,
  "today_completed": 5,
  "recent_activities": [
    {
      "id": 12,
      "name": "夸克任务 [黑镜]",
      "status": "success",
      "last_run": "2026-04-10T14:30:00Z",
      ...
    }
  ]
}
```

---

## 2. 获取实时系统日志 (SSE)
建立 Server-Sent Events 连接，实时获取网盘 API 请求与响应日志。

- **URL**: `/dashboard/logs`
- **Method**: `GET`
- **Headers**: `Accept: text/event-stream`
- **Response**: `EventStream`
    - **Event: `message`**: 包含格式化的日志文本。

---

## 3. 获取近期日志历史
获取内存中缓存的最近若干条日志记录。

- **URL**: `/dashboard/logs/recent`
- **Method**: `GET`
- **Response**: `Array<string>`

---

## 4. 清空日志历史
清空后端内存中缓存的所有日志记录。

- **URL**: `/dashboard/logs/recent`
- **Method**: `DELETE`
- **Response**: `{"message": "logs cleared"}`
