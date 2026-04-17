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
| `share_url` | string | 是 | 139/Quark 分享链接 (支持子目录链接) |
| `extract_code`| string | 否 | 分享提取码 |
| `save_path` | string | 是 | 云盘内保存路径 (如 `/Movies`) |
| `start_file_id`| string | 否 | 起始文件 ID，系统将从此文件开始(含)向前转存 |
| `start_file_name`| string | 否 | 起始文件名称，用于前端 UI 快速回显，减少接口解析等待 |
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

---

## 6. 任务重命名与预览
在保存任务前，基于真实分享链接的内容，预览重命名规则的执行效果。

- **URL**: `/tasks/preview`
- **Method**: `POST`
- **Payload**:
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `account_id` | uint | 模拟执行的账号 ID |
| `share_url` | string | 待解析的分享链接 |
| `extract_code`| string | 提取码 |
| `pattern` | string | 正则表达式 |
| `replacement` | string | 替换规则 |
| `name` | string | 任务名称（用于 `{TASKNAME}` 变量替换） |

- **Response**: `Array<PreviewResult>`

### PreviewResult 对象结构
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `original_name`| string | 原始文件名 |
| `new_name` | string | 重命名后的名称 |
| `matched` | bool | 正则是否匹配成功 |

---

## 7. 解析分享链接详情
仅用于拉取分享链接内的文件列表，供用户在 UI 中选择起始转存点。

- **URL**: `/tasks/parse_share`
- **Method**: `POST`
- **Payload**:
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `account_id` | uint | 关联的账号 ID |
| `share_url` | string | 待解析的分享链接 |
| `extract_code`| string | 提取码 |

- **Response**: `Array<FileInfo>` (包含文件名、ID、更新时间等)
