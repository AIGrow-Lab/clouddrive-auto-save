# 夸克网盘 (Quark) API 接口手册

本文档整理了 QuarkPan 及 quark-auto-save 项目中使用到的所有夸克网盘底层接口逻辑，供持续开发、接口扩展及维护参考。

---

## 1. 基础信息与认证

### 1.1 API 域名

| 域名 | 用途 | 标识 |
|------|------|------|
| `https://drive-pc.quark.cn/1/clouddrive` | PC 端主 API | `BASE_URL` |
| `https://drive.quark.cn/1/clouddrive` | 分享相关 API | `SHARE_BASE_URL` |
| `https://pan.quark.cn/account` | 账号信息 | — |
| `https://drive-m.quark.cn` | 移动端 API | `BASE_URL_APP` |

### 1.2 默认请求参数

所有请求均附带以下默认查询参数：

| 参数 | 值 | 说明 |
|------|-----|------|
| `pr` | `ucpro` | 产品标识 |
| `fr` | `pc` | 来源 (PC/Android) |
| `uc_param_str` | `""` | UC 参数 |
| `__t` | 当前毫秒时间戳 | 防缓存 |
| `__dt` | `1000` | 固定值 |

### 1.3 默认请求头

```
user-agent: Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 ...
origin: https://pan.quark.cn
referer: https://pan.quark.cn/
accept-language: zh-CN,zh;q=0.9
accept: application/json, text/plain, */*
content-type: application/json
cookie: <用户的 Cookie 字符串>
```

### 1.4 认证方式

使用 **Cookie 认证**，关键 Cookie 字段：`__kps`, `__uid`, `__pus`, `sign`, `vcode`。

移动端接口需从 Cookie 中提取 `kps`, `sign`, `vcode` 作为 URL 签名参数。

---

## 2. 登录接口

### 2.1 获取二维码登录 Token

- `GET https://uop.quark.cn/cas/ajax/getTokenForQrcodeLogin`
- 参数: `client_id: 532`, `v: 1.2`, `request_id: <UUID>`
- 响应: `status=2000000` 时成功，`data.members.token` 为二维码 token

### 2.2 检查二维码登录状态 / 获取 Service Ticket

- `GET https://uop.quark.cn/cas/ajax/getServiceTicketByQrcodeToken`
- 参数: `client_id: 532`, `v: 1.2`, `token: <二维码token>`, `request_id: <UUID>`
- 响应: `status=2000000`, `data.members.service_ticket` 存在表示登录成功
- 失败状态码: `50004001` = 等待扫码, `50004002/50004003/50004004` = 登录失败

### 2.3 获取用户信息

- `GET https://pan.quark.cn/account/info`
- 参数: `st: <service_ticket>`, `lw: scan` (登录时) 或 `fr: pc`, `platform: pc` (常规)
- 响应: `data` 包含 `nickname` 等用户信息

---

## 3. 文件管理接口

### 3.1 获取文件列表 (sort)

- `GET {BASE_URL}/file/sort`
- 参数:

| 参数 | 说明 |
|------|------|
| `pdir_fid` | 父文件夹 ID，`0` 为根目录 |
| `_page` | 页码 (从 1 开始) |
| `_size` | 每页数量 (默认 50) |
| `_sort` | 排序字段，格式 `field:order`，如 `file_name:asc` |
| `_fetch_total` | `1` 获取总数 |
| `_fetch_sub_dirs` | `0` 不获取子目录 |
| `_fetch_full_path` | `0/1` 是否获取完整路径 |
| `fetch_all_file` | `1` 获取所有文件 |
| `fetch_risk_file_name` | `1` 获取违规文件真实名 |

- 响应: `data.list` 数组，每项含 `fid`, `file_name`, `file_type`, `dir`, `size`, `pdir_fid`, `updated_at` 等

### 3.2 获取文件信息 (单个/批量)

- `GET {BASE_URL}/file`
- 参数: `fids`: 文件 ID（支持逗号分隔多个）
- 响应: `data.list` 数组，每项含 `fid`, `file_name`, `file_path`, `file_type`, `size`, `dir`

### 3.3 通过路径批量获取文件 ID

- `POST {BASE_URL}/file/info/path_list`
- 请求体: `{file_path: ["/path1", "/path2", ...], namespace: "0"}`
- 说明: 单次最多 50 个路径，超出需分批
- 响应: `data` 为 fid 数组，每项含 `file_path` 和 `fid`

### 3.4 创建文件夹

- `POST {BASE_URL}/file`
- 请求体: `{pdir_fid, file_name, dir_init_lock: false, dir_path}`
  - `dir_path`: 完整路径字符串 (如 `/a/b/c`)，与 `file_name` 二选一
- 响应: `data` 包含新创建的文件夹信息 (含 `fid`)

### 3.5 重命名文件/文件夹

- `POST {BASE_URL}/file/rename`
- 请求体: `{fid, file_name}`
- 响应: `code=0` 表示成功

### 3.6 删除文件/文件夹

- `POST {BASE_URL}/file/delete`
- 请求体: `{action_type: 2, filelist: [fid1, fid2, ...], exclude_fids: []}`
  - `action_type`: `2` = 删除
- 响应: `data.task_id` 用于查询删除任务状态

### 3.7 移动文件

- `POST {BASE_URL}/file/move`
- 请求体: `{action_type: 1, to_pdir_fid, filelist, exclude_fids}`
  - `action_type`: `1` = 移动
- 响应: `data.task_id`, `data.finish` (是否同步完成)

### 3.8 搜索文件

- `GET {BASE_URL}/file/search`
- 参数: `q`(关键词), `_page`, `_size`, `_fetch_total: 1`, `_sort`, `_is_hl: 1`(高亮)

### 3.9 获取文件夹树结构

- `GET {BASE_URL}/file/tree`
- 参数: `pdir_fid`, `max_depth`

---

## 4. 下载接口

### 4.1 获取下载链接

- `POST {BASE_URL}/file/download`
- 请求体: `{fids: [file_id1, file_id2, ...]}`
- 响应: `data` 数组，每项含 `fid`, `file_name`, `download_url`, `size`

---

## 5. 上传接口

### 5.1 预上传请求

- `POST {BASE_URL}/file/upload/pre`
- 请求体:
  ```json
  {
    "ccp_hash_update": true,
    "parallel_upload": true,
    "pdir_fid": "父文件夹ID",
    "dir_name": "",
    "size": 文件大小,
    "file_name": "文件名",
    "format_type": "MIME类型",
    "l_updated_at": 毫秒时间戳,
    "l_created_at": 毫秒时间戳
  }
  ```
- 响应: `data` 含 `task_id`, `auth_info`, `upload_id`, `obj_key`, `bucket` (默认 `ul-zb`), `callback`

### 5.2 更新文件哈希

- `POST {BASE_URL}/file/update/hash`
- 请求体: `{task_id, md5, sha1}`

### 5.3 获取上传授权

- `POST {BASE_URL}/file/upload/auth`
- 请求体:
  ```json
  {
    "task_id": "...",
    "auth_info": "...",
    "auth_meta": "PUT\n\n<content_type>\n<oss_date>\nx-oss-date:<oss_date>\n..."
  }
  ```
- 响应: `data.auth_key` 用于上传请求的 Authorization 头
- 上传 URL: `https://{bucket}.pds.quark.cn/{obj_key}?partNumber={N}&uploadId={ID}`

### 5.4 上传分片到 OSS

- `PUT https://{bucket}.pds.quark.cn/{obj_key}?partNumber={N}&uploadId={ID}`
- 请求头:
  - `Content-Type`: MIME 类型
  - `x-oss-date`: UTC 日期
  - `x-oss-user-agent`: `aliyun-sdk-js/1.0.0 ...`
  - `authorization`: 从授权 API 获取的 auth_key
  - `X-Oss-Hash-Ctx`: Base64 编码的 SHA1 增量哈希上下文 (分片 2+)
- 响应: HTTP 200，从 `ETag` 头获取分片标识
- **分片策略**: < 5MB 单分片, >= 5MB 多分片 (每片 4MB)

### 5.5 POST 完成合并 (OSS)

- `POST https://{bucket}.pds.quark.cn/{obj_key}?uploadId={ID}`
- 请求头: `Content-Type: application/xml`, `x-oss-callback`, `Content-MD5`, `authorization`
- 请求体 (XML):
  ```xml
  <?xml version="1.0" encoding="UTF-8"?>
  <CompleteMultipartUpload>
    <Part><PartNumber>1</PartNumber><ETag>"etag值"</ETag></Part>
  </CompleteMultipartUpload>
  ```
- 响应: HTTP 200 或 203 (203 表示 callback 失败但文件已上传)

### 5.6 完成上传 (通知夸克服务器)

- `POST {BASE_URL}/file/upload/finish`
- 请求体: `{task_id, obj_key}`

---

## 6. 分享接口

### 6.1 获取分享访问令牌 (stoken)

- `POST {SHARE_BASE_URL}/share/sharepage/token`
- 请求体: `{pwd_id, passcode, support_visit_limit_private_share: true}`
- 响应: `data.stoken`
- 也可用于验证资源是否失效

### 6.2 获取分享详情/文件列表

- `GET {SHARE_BASE_URL}/share/sharepage/detail`
- 参数:

| 参数 | 说明 |
|------|------|
| `pwd_id` | 分享 ID |
| `stoken` | 访问令牌 |
| `pdir_fid` | 父目录 ID (`0` 为根) |
| `force` | `0` |
| `_page`, `_size` | 分页 (默认 50) |
| `_fetch_total` | `1` |
| `_sort` | `file_type:asc,file_name:asc` |
| `ver` | `2` |
| `fetch_share_full_path` | `0/1` |

- 响应: `data.list` 含 `fid`, `file_name`, `dir`, `share_fid_token`, `size`, `obj_category` 等
- **异常处理**: `code: 0` 但 `data.list` 为空时，表示链接已失效/被取消/文件为空

### 6.3 转存分享文件

- `POST {SHARE_BASE_URL}/share/sharepage/save`
- 请求体:
  ```json
  {
    "fid_list": ["文件ID列表"],
    "fid_token_list": ["文件token列表"],
    "to_pdir_fid": "目标文件夹ID",
    "pwd_id": "分享ID",
    "stoken": "访问令牌",
    "pdir_fid": "0",
    "pdir_save_all": false,
    "exclude_fids": [],
    "scene": "link"
  }
  ```
- 响应: `data.task_id` 用于查询转存任务状态

### 6.4 创建分享链接

- `POST {SHARE_BASE_URL}/share`
- 请求体:
  ```json
  {
    "fid_list": ["文件ID列表"],
    "title": "分享标题",
    "url_type": 1,
    "expired_type": 1,
    "passcode": "提取码"
  }
  ```
  - `url_type`: `1` = 公开链接, `2` = 私密链接 (有密码)
  - `expired_type`: `1` = 永久, `2` = 有期限
- 响应: `data.task_id`

### 6.5 获取分享详情 (密码/链接)

- `POST {SHARE_BASE_URL}/share/password`
- 请求体: `{share_id}`
- 响应: `data` 包含完整分享信息 (含链接、密码等)

### 6.6 获取我的分享列表

- `GET {SHARE_BASE_URL}/share/mypage/detail`
- 参数: `_page`, `_size`, `_order_field: created_at`, `_order_type: desc`, `_fetch_total: 1`
- 响应: `data.list` 含 `share_id`, `share_url`, `title`, `status`, `created_at`, `expired_at`, `file_num`

### 6.7 删除分享

- `POST {SHARE_BASE_URL}/share/delete`
- 请求体: `{share_id}`

---

## 7. 任务查询接口

### 7.1 查询异步任务状态

- `GET {BASE_URL}/task`
- 参数: `task_id`, `retry_index`
- 响应 `data.status`:

| 状态 | 含义 |
|------|------|
| `0` | PENDING (等待中) |
| `1` | RUNNING (执行中) |
| `2` | 完成 |
| `3` | 失败 |

- 转存/删除/移动等操作均为异步，需轮询此接口等待完成 (推荐 500ms 间隔)

---

## 8. 存储/容量/会员接口

### 8.1 获取存储空间信息 (PC 端)

- `GET {BASE_URL}/capacity`

### 8.2 获取会员与容量信息 (PC Web 端)

- `GET https://pan.quark.cn/1/clouddrive/member?pr=ucpro&fr=pc`
- 响应:
  - `data.total_capacity`: 总空间 (Bytes)
  - `data.use_capacity`: 已用空间 (Bytes)
  - `data.member_type`: 会员类型 (`NORMAL`, `SUPER_VIP` 等)

### 8.3 获取成长信息 (签到进度/会员类型/总容量)

- `GET {BASE_URL_APP}/1/clouddrive/capacity/growth/info` (移动端)
- 参数: `pr`, `fr: android`, `kps`, `sign`, `vcode`
- 响应:
  - `member_type`: `NORMAL` / `EXP_SVIP` / `SUPER_VIP` / `Z_VIP`
  - `total_capacity`: 总空间
  - `cap_composition.sign_reward`: 签到累计获得空间
  - `cap_sign.sign_daily`: 今日是否已签到
  - `cap_sign.sign_daily_reward`: 今日签到奖励 (字节)
  - `cap_sign.sign_progress`: 连签进度
  - `cap_sign.sign_target`: 连签目标

---

## 9. 签到接口

### 9.1 执行每日签到

- `POST {BASE_URL_APP}/1/clouddrive/capacity/growth/sign` (移动端)
- 参数: `pr`, `fr: android`, `kps`, `sign`, `vcode`
- 请求体: `{sign_cyclic: true}`
- 响应: `data.sign_daily_reward` 为签到奖励 (字节)

---

## 10. 回收站接口

### 10.1 获取回收站列表

- `GET {BASE_URL}/file/recycle/list`
- 参数: `_page`, `_size`, `pr`, `fr`, `uc_param_str`
- 响应: `data.list`，每项含 `record_id`, `fid`

### 10.2 彻底删除回收站文件

- `POST {BASE_URL}/file/recycle/remove`
- 请求体: `{select_mode: 2, record_list: [record_id1, ...]}`

---

## 11. 移动端特殊参数

当 Cookie 中包含 `kps/sign/vcode` 时，自动切换到移动端域名 `drive-m.quark.cn`，并添加以下额外参数：

| 参数 | 值 |
|------|-----|
| `device_model` | `M2011K2C` |
| `entry` | `default_clouddrive` |
| `fr` | `android` |
| `ve` | `7.4.5.680` |
| `pr` | `ucpro` |
| `dt` | `phone` |
| `data_from` | `ucapi` |
| `kps/sign/vcode` | 从 Cookie 提取 |
| `app` | `clouddrive` |

此时请求头中不再包含 `cookie`。

---

## 12. 错误处理机制

### 12.1 HTTP 状态码

| 状态码 | 含义 |
|--------|------|
| 401 | 认证失败 |
| 403 | Cookie 过期 / 访问被拒 |
| >= 400 | 其他 HTTP 错误 |

### 12.2 API 响应码

- `code != 0`: API 调用失败
- 消息包含 `login` 或 `auth`: 认证问题
- `status == 500`: 网络异常

### 12.3 转存任务错误关键词

- `capacity limit`: 容量不足
- `permission denied`: 权限不足
- `share expired`: 分享已过期

### 12.4 致命错误判定 ([Fatal])

当驱动检测到不可恢复的业务错误时，返回带有 `[Fatal]` 前缀的错误信息。

**Quark 致命码**:
| 错误码 | 含义 |
|--------|------|
| `41010` | 违规内容 |
| `41008` | 未提供提取码 |
| `41007` / `24000` / `41009` | 提取码错误 |
| `24001` | 链接已失效 |

### 12.5 异常类层次结构 (QuarkPan)

```
QuarkClientError
├── AuthenticationError   # 认证相关
├── ConfigError           # 配置相关
├── APIError              # API 调用 (含 status_code, response_data)
├── NetworkError          # 网络相关
├── FileNotFoundError     # 文件未找到
├── ShareLinkError        # 分享链接相关
└── DownloadError         # 下载相关
```

### 12.6 交互逻辑联动

- **阻断**: 后端 `runTask` 会拦截带有 `[Fatal]` 标记的任务，防止无效请求
- **警示**: 前端 UI 会将此类任务标记为红色 "LINK ERROR"，并禁用运行按钮
- **解封**: 用户通过"编辑并保存"操作，可强制重置任务状态，解除 `[Fatal]` 锁定

---

## 附录: 接口汇总表

| 序号 | 接口名称 | 方法 | 路径 |
|------|---------|------|------|
| 1 | 获取二维码 Token | GET | `uop.quark.cn/cas/ajax/getTokenForQrcodeLogin` |
| 2 | 检查登录状态 | GET | `uop.quark.cn/cas/ajax/getServiceTicketByQrcodeToken` |
| 3 | 获取用户信息 | GET | `pan.quark.cn/account/info` |
| 4 | 文件列表 | GET | `/file/sort` |
| 5 | 文件信息 | GET | `/file` |
| 6 | 路径批量获取 fid | POST | `/file/info/path_list` |
| 7 | 创建文件夹 | POST | `/file` |
| 8 | 重命名 | POST | `/file/rename` |
| 9 | 删除文件 | POST | `/file/delete` |
| 10 | 移动文件 | POST | `/file/move` |
| 11 | 搜索文件 | GET | `/file/search` |
| 12 | 文件夹树 | GET | `/file/tree` |
| 13 | 获取下载链接 | POST | `/file/download` |
| 14 | 预上传 | POST | `/file/upload/pre` |
| 15 | 更新文件哈希 | POST | `/file/update/hash` |
| 16 | 获取上传授权 | POST | `/file/upload/auth` |
| 17 | 完成上传 | POST | `/file/upload/finish` |
| 18 | 获取 stoken | POST | `/share/sharepage/token` |
| 19 | 分享详情 | GET | `/share/sharepage/detail` |
| 20 | 转存文件 | POST | `/share/sharepage/save` |
| 21 | 创建分享 | POST | `/share` |
| 22 | 获取分享密码/链接 | POST | `/share/password` |
| 23 | 我的分享列表 | GET | `/share/mypage/detail` |
| 24 | 删除分享 | POST | `/share/delete` |
| 25 | 查询任务状态 | GET | `/task` |
| 26 | 存储空间 | GET | `/capacity` |
| 27 | 会员容量信息 | GET | `pan.quark.cn/1/clouddrive/member` |
| 28 | 成长/签到信息 | GET | `/capacity/growth/info` |
| 29 | 执行签到 | POST | `/capacity/growth/sign` |
| 30 | 回收站列表 | GET | `/file/recycle/list` |
| 31 | 彻底删除 | POST | `/file/recycle/remove` |

> 路径相对于 `https://drive-pc.quark.cn/1/clouddrive` 或 `https://drive.quark.cn/1/clouddrive`，完整 URL 需拼接对应 BASE_URL。
