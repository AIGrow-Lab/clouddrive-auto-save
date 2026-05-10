# 系统设置 (Settings)

## 1. 获取全局调度设置

获取系统当前的全局定时触发规则及开关状态。

- **URL**: `/settings/schedule`
- **Method**: `GET`
- **Response**:

```json
{
  “enabled”: true,
  “cron”: “0 0 2 * * *”
}
```

---

## 2. 更新全局调度设置

修改全局调度规则。修改后系统会立即更新后台调度引擎。

- **URL**: `/settings/schedule`
- **Method**: `POST`
- **Payload**:
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `enabled` | bool | 全局总开关 |
| `cron` | string | 全局 Cron 表达式 (例如 `0 0 2 * * *` 代表每天凌晨 2 点) |

---

## 3. 调度优先级逻辑 (Logic Priority)

- **跟随全局 (Global)**: 任务执行受全局开关 `enabled` 与全局 `cron` 的共同约束。
- **自定义 (Custom)**: 任务拥有独立的 Cron 表达式，不受全局配置影响。
- **手动 (Off)**: 系统不自动执行，仅在用户点击 UI 上的”运行”按钮时触发。

---

## 4. OpenList 扫描配置

### 4.1 配置项说明

通过 `GET/POST /settings/global` 接口读写以下配置：

| 配置键 | 类型 | 说明 |
| :--- | :--- | :--- |
| `openlist_enabled` | string | 是否启用自动扫描 (`”true”` / `”false”`) |
| `openlist_api_url` | string | OpenList API 地址 (例如 `http://127.0.0.1:23541`) |
| `openlist_api_token` | string | OpenList API 认证 Token |

### 4.2 配置示例

```json
{
  “openlist_enabled”: “true”,
  “openlist_api_url”: “http://127.0.0.1:23541”,
  “openlist_api_token”: “openlist-913f51c8-720f-45ec-b1a2-a388ff5c50f4”
}
```

### 4.3 自动触发逻辑

- **单任务执行**：任务完成且有新转存文件时，自动触发 OpenList 扫描
- **批量任务执行**：所有任务完成后，统一触发一次扫描
- **延迟合并**：3 秒内的多次触发会合并为一次扫描

---

## 5. 手动触发 OpenList 扫描

立即触发 OpenList 文件扫描，用于手动生成 strm 文件。

- **URL**: `/openlist/scan`
- **Method**: `POST`
- **Headers**: 无特殊要求
- **Request Body**: 无

- **Response (成功)**:

```json
{
  "message": "扫描已触发"
}
```

- **Response (失败)**:

```json
{
  "error": "加载 OpenList 配置失败"
}
```

或

```json
{
  "error": "API 返回错误状态码 500: internal error"
}
```

**说明**:
- 调用此接口会先从数据库重新加载 OpenList 配置
- 如果 OpenList 未配置或已禁用，接口会返回成功但不会实际触发扫描
- 扫描请求超时时间为 30 秒
