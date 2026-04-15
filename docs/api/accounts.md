# 账号管理 (Accounts)

## 1. 获取账号列表
获取系统中所有已绑定的云盘账号。

- **URL**: `/accounts`
- **Method**: `GET`
- **Response**: `Array<Account>`

### Account 对象结构
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | uint | 账号唯一标识 |
| `platform` | string | `139` 或 `quark` |
| `nickname` | string | 云盘真实昵称 |
| `vip_name` | string | 会员等级（如：白银会员、SVIP） |
| `capacity_used`| int64 | 已用容量 (Bytes) <br>*(139 含个人+家庭空间)* |
| `capacity_total`| int64 | 总容量 (Bytes) <br>*(139 含个人+家庭空间)* |
| `status` | int | 状态 (1:正常, 0:失效) |
| `last_check` | string | 最后一次校验的时间戳 |

---

## 2. 添加新账号
绑定一个新的移动云盘或夸克网盘账号。

- **URL**: `/accounts`
- **Method**: `POST`
- **Payload**:
| 字段 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `platform` | string | 是 | `139` 或 `quark` |
| `account_name`| string | 是 | 备注名或手机号 |
| `cookie` | string | 否 | 浏览器全量 Cookie |
| `auth_token` | string | 否 | 仅 139 支持，抓包获取的 Basic 串 |

---

## 3. 更新账号信息
修改已存在的账号配置。

- **URL**: `/accounts/:id`
- **Method**: `PUT`
- **Payload**: 与添加账号一致（仅需传递待修改字段）。

---

## 4. 账号有效性校验
手动触发后端模拟登录，校验凭证是否有效并更新昵称、容量及会员信息。

- **URL**: `/accounts/:id/check`
- **Method**: `/POST`
- **Response**: 返回更新后的账号对象。
- **Errors**:
    - `401 Unauthorized`: 登录凭证失效（如 Token 过期），Body 中包含 `{"error": "登录已失效 (Token Invalid)"}`。
    - `404 Not Found`: 账号不存在。
    - `500 Internal Server Error`: 驱动加载失败或系统内部错误。

---

## 5. 删除账号
从系统中移除指定的云盘账号。

- **URL**: `/accounts/:id`
- **Method**: `DELETE`
- **Success Response**: `200 OK`
- **Errors**:
    - `409 Conflict`: 账号当前有关联的任务。Body: `{"error": "该账号有关联的 X 个任务，请先删除任务"}`。
    - `404 Not Found`: 账号不存在。

---

## 6. 获取网盘目录树
获取指定账号下特定目录的子文件夹列表（支持懒加载）。

- **URL**: `/accounts/:id/folders`
- **Method**: `GET`
- **Params**:
    - `parent_id`: 父目录 ID（夸克为 FID，139 为路径或 ID）。默认为空（根目录）。
    - `parent_path`: 父目录的绝对路径。用于后端自动拼接并返回子目录的完整 `path` 字段。
- **Response**: `Array<FolderItem>`

### FolderItem 对象结构
| 字段 | 类型 | 说明 |
| :--- | :--- | :--- |
| `id` | string | 文件夹唯一标识 (FID) |
| `path` | string | 文件夹的绝对路径 |
| `label` | string | 文件夹显示名称 |
| `isLeaf` | bool | 是否为叶子节点（目前统一返回 false 以支持展开） |

---

## 7. 新建网盘文件夹
在指定账号的特定目录下创建一个新文件夹。

- **URL**: `/accounts/:id/folders`
- **Method**: `POST`
- **Payload**:
| 字段 | 类型 | 必填 | 说明 |
| :--- | :--- | :--- | :--- |
| `parent_id` | string | 否 | 父目录 ID |
| `parent_path`| string | 是 | 父目录绝对路径 |
| `name` | string | 是 | 新文件夹名称 |

- **Response**: 返回新建成功的 `FolderItem` 对象。
