# 账号管理 (Accounts)

## 1. 获取账号列表
获取系统中所有已绑定的云盘账号。

- **URL**: `/accounts`
- **Method**: `GET`
- **Response**: `Array<Account>`

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

## 3. 账号有效性校验
手动触发后端模拟登录，校验凭证是否有效并更新昵称。

- **URL**: `/accounts/:id/check`
- **Method**: `POST`
- **Response**: 返回更新后的账号对象或 401 错误。

---

## 4. 删除账号
彻底移除该账号信息。

- **URL**: `/accounts/:id`
- **Method**: `DELETE`
