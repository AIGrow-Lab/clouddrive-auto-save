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
| `share_parent_id`| string | 否 | **[新]** 139 分享链接的目录 ID，用于指定浏览子目录时的起始位置 |
| `schedule_mode`| string | 否 | **[新]** 调度模式: `global` (跟随全局), `custom` (自定义), `off` (手动)。默认为 `global` |
| `cron` | string | 否 | **[新]** 当 `schedule_mode` 为 `custom` 时必填。Cron 表达式 (秒 分 时 日 月 周) |
| `percent` | int | 否 | **[内部]** 实时进度百分比 |
| `stage` | string | 否 | **[内部]** 当前执行阶段 |
| `pattern` | string | 否 | 正则匹配表达式 |
| `replacement` | string | 否 | 替换模板及魔法变量 |

---

## 3. 更新任务配置

修改已存在任务的链接或规则。

- **URL**: `/tasks/:id`
- **Method**: `PUT`
- **Payload**: 与创建任务相同。
- **业务逻辑**: 如果修改了 `share_url` 或 `extract_code`，系统会自动将任务的 `status` 重置为 `pending` 并清空 `message`，以此来自动解除因原链接失效导致的 `[Fatal]` 封锁。

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

## 5.5 批量运行所有任务

一键启动系统中所有健康的待转存任务。

- **URL**: `/tasks/run_all`
- **Method**: `POST`
- **逻辑**:
  - 自动跳过状态为 `running` 的任务。
  - 自动跳过 message 中包含 `[Fatal]` 关键字的故障任务。
- **Response**: `{"message": "批量执行已开启，共触发 X 个任务", "count": X}`

---

## 5.6 忽略/消除错误状态 (Dismiss)

手动清除任务的错误消息及状态，使其恢复为待命中状态。常用于网盘容量不足清理后，或网络波动导致失败后的重试。

- **URL**: `/tasks/:id/dismiss`
- **Method**: `POST`
- **Side Effect**: 将任务的 `status` 置为 `pending`，清空 `message` 和 `stage`。

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
| `parent_id` | string | **[新]** 父目录 ID，用于浏览子目录内容 |
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

1. **分享内容解析与排序**：系统首先解析分享链接，并强制将文件列表按 **“更新时间”从新到旧 (Descending)** 进行排序。
2. **双重条件过滤 (AND 关系)**：系统会根据以下两个条件过滤待转存文件：
    - **起始点过滤**：如果设置了“起始文件”，系统将仅保留该文件及其之后更新（即时间轴上更靠后）的所有文件。
    - **正则匹配过滤**：仅当文件名符合“重命名正则”规则时，才会被加入转存队列。
    - **注意**：只有同时满足上述两个条件的文件才会被处理。
3. **预测式智能去重**：
    - 在正式转存前，系统会利用重命名引擎预先计算出每个文件的“入库文件名”。
    - 系统会拉取目标网盘目录的完整文件列表。
    - **比对逻辑**：拿“预测后的入库文件名”去目标目录中检索。
    - **结果**：如果目标目录已存在同名文件，系统将直接 **跳过** 该文件的转存，避免产生冗余副本。
4. **转存与异步等待 (Quark 专用)**：
    - 对于夸克网盘，转存操作是异步的。系统会获取 `task_id` 并自动轮询，直到网盘返回“转存成功”状态。
5. **自动更名确认**：转存成功后，系统会自动调用网盘底层的重命名接口，将文件修改为最终的入库名称。
6. **状态回传**：执行结果会实时更新至任务的 `message` 字段。

---

## 9. 魔法变量说明 (Magic Variables)

魔法变量允许您无需编写复杂的正则捕获组，即可自动从原始文件名中提取关键信息。

### 支持的变量列表

| 变量名 | 提取规则 (Regex) | 说明 | 示例 (原名: `2024.04.20 功夫熊猫4.mp4`) |
| :--- | :--- | :--- | :--- |
| `{YEAR}` | `\b(?:18\|19\|20)\d{2}\b` | 提取 4 位年份数字 | `2024` |
| `{DATE}` | `\b(?:18\|19\|20)?\d{2}[\.\-/年]\d{1,2}[\.\-/月]\d{1,2}\b` | 提取日期并自动清洗为纯数字 | `20240420` |
| `{CHINESE}`| `\p{Han}{2,}` | 提取连续的中文名称 | `功夫熊猫` |
| `{EXT}` | `\.(\w+)$` | 提取文件后缀名 (不含点) | `mp4` |
| `{OLDNAME}` | (内置变量) | **原始完整文件名** | `2024.04.20 功夫熊猫4.mp4` |
| `{TASKNAME}`| (内置变量) | 当前任务的显示名称 | - |

### 使用方法

在“替换规则 (Replacement)”中直接引用即可。例如：`[{DATE}] {TASKNAME}.{EXT}`。
