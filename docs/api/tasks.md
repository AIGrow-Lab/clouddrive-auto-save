# 任务管理 (Tasks)

## 1. 获取任务列表
获取所有自动转存任务，默认包含关联的账号信息。

- **URL**: `/tasks`
- **Method**: `GET`

---

## 2. 创建新任务
- **URL**: `/tasks`
- **Method**: `POST`
- **Payload**:
| 字段 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `name` | string | 是 | 任务显示名称 |
| `account_id` | uint | 是 | 关联的账号 ID |
| `share_url` | string | 是 | 139/Quark 分享链接 |
| `extract_code`| string | 否 | 分享提取码 |
| `save_path` | string | 是 | 云盘内保存路径 (如 `/Movies`) |
| `pattern` | string | 否 | 正则匹配表达式 |
| `replacement` | string | 否 | 替换模板及魔法变量 |

---

## 3. 更新任务配置
修改已存在任务的链接或规则。

- **URL**: `/tasks/:id`
- **Method**: `PUT`
- **Payload**: 与创建任务相同。

---

## 4. 删除任务
- **URL**: `/tasks/:id`
- **Method**: `DELETE`

---

## 5. 手动运行任务
将指定任务立即提交至 Worker 队列执行。

- **URL**: `/tasks/:id/run`
- **Method**: `POST`
- **Error**: 若任务已在运行，返回 `400`。
