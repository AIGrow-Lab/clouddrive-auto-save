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
仅用于拉取分享链接内的文件列表，供用户在 UI 中选择起始转存点。**此接口已支持重命名预览及同名预检。**

- **URL**: `/tasks/parse_share`
- **Method**: `POST`
- **Payload**:
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `account_id` | uint | 关联的账号 ID |
| `share_url` | string | 待解析的分享链接 |
| `extract_code`| string | 提取码 |
| `save_path` | string | 预期的保存路径（用于执行同名检查） |
| `pattern` | string | 正则表达式（用于重命名预览） |
| `replacement` | string | 替换规则 |
| `name` | string | 任务名称 |

- **Response**: `Array<FileInfo>`

### FileInfo 结构说明 (含增强字段)
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | string | 文件 ID |
| `name` | string | 原始文件名 |
| `new_name` | string | **[增强]** 经过重命名规则处理后的预览名 |
| `is_existed` | bool | **[增强]** 此文件（基于新名）是否已存在于目标路径中 |
| `is_folder` | bool | 是否为文件夹 |
| `size` | int64 | 文件大小 |
| `updated_at` | string | 更新时间 |

---

## 8. 自动重命名与去重逻辑说明 (Workflow)

当任务执行（无论是定时还是手动触发）时，系统遵循以下核心逻辑以确保效率与整洁：

1.  **重命名预览生成**：系统首先解析分享链接，并根据任务配置的正则规则，为每个待转存文件计算出“入库文件名”。
2.  **智能同名去重**：
    *   系统会拉取目标网盘目录的完整文件列表。
    *   **比对逻辑**：拿“入库文件名”去目标目录中检索。
    *   **结果**：如果目标目录已存在同名文件，系统将直接**跳过**该文件的转存，避免产生冗余副本（如 `filename(1).mp4`）。
3.  **转存与自动更名**：
    *   对于通过去重校验的新文件，系统发起批量转存。
    *   **自动重命名**：转存成功后，系统会自动调用网盘底层的重命名接口（139 的 `/hcy/file/update` 或 Quark 的 `/1/clouddrive/file/rename`），将文件修改为最终的入库名称。
4.  **状态回传**：执行结果（包含跳过的文件数）会实时更新至任务的 `message` 字段。
