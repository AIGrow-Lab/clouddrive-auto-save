# 移动云盘 (139) API 接口手册

本文档整理了移动云盘 139 驱动中使用到的所有底层网盘接口逻辑，供持续开发、接口扩展及维护参考。

---

## 1. 基础信息与认证

### 1.1 域名架构

移动云盘采用多子域名架构：

| 域名 | 用途 |
|------|------|
| `https://yun.139.com` | 基础/会员接口 |
| `https://user-njs.yun.139.com` | 用户管理 |
| `https://share-kd-njs.yun.139.com` | 分享/转存 |
| `https://personal-kd-njs.yun.139.com` | 私有文件 (HCY) |

### 1.2 认证方式

- `Authorization`: `Basic [base64(pc:手机号:token)]`
- `Cookie`: 浏览器登录后的 Cookie

### 1.3 签名机制 (mcloud-sign)

用于 HCY 私有接口和 Orchestration 接签。格式为 `datetime,randomStr,hash`。

**hash 计算流程**：
1. 对 JSON Body 进行序列化
2. URI 编码
3. 字符排序
4. Base64 编码
5. MD5 哈希
6. 与时间戳 + 随机串的 MD5 进行二次哈希

---

## 2. 账号与用户接口

### 2.1 获取用户信息 (getUser)

- `POST /user/getUser` (User Host)
- **重要返回结构**：
  - `userDomainId`: 用户核心标识（容量查询必填）
  - `auditNickName`: 审核通过的昵称。未手动修改时为 null
  - `userProfileInfo.userName`: 最新的昵称字段所在路径
  - `auth.memberLevel`: 部分版本在此处返回会员等级
  - `loginName / account / msisdn / phoneNumber`: 用户的真实手机号
  - `userServiceType`: 用户服务类型标识（如 "8" 代表移动云盘会员）

### 2.2 获取云盘配额 (个人与家庭)

- `POST /user/disk/getPersonalDiskInfo` (获取个人空间)
- `POST /user/disk/getFamilyDiskInfo` (获取家庭空间)
- Body: `{"userDomainId": "xxx"}`
- 返回: `diskSize`, `freeDiskSize` (MB)
- 系统将两者累加得出真实的总可用空间

### 2.3 获取会员等级 (queryUserBenefits)

- `POST /orchestration/group-rebuild/member/v1.0/queryUserBenefits` (Base Host)
- **鉴权要求**: 必须携带 `mcloud-sign` 签名（基于 Body 计算）
- Body: `{"isNeedBenefit": 1, "commonAccountInfo": {"account": "手机号", "accountType": 1}}`
- **返回解析**:
  - 会员等级从 `data.userSubMemberList[0].memberLvName` 获取（如："白银会员"）
  - **非会员时**: `data.userSubMemberList` 为空数组 `[]`

---

## 3. 核心鉴权与异常处理

### 3.1 动态签名 (mcloud-sign)

用于 HCY 及 Orchestration 系列接口。Orchestration 接口签名必须基于请求 Body、当前时间戳及 16 位随机字符串计算二次 MD5 哈希。

### 3.2 Token 异常状态码

| 状态码 | 含义 |
|--------|------|
| `05050009` | 非法的 Token（Authorization 已过期或错误） |
| `1010010003` | 登录已失效 |
| `1010010014` | 签名校验失败 |

---

## 4. 私有文件操作 (HCY 系列)

### 4.1 文件列表 (list)

- `POST /hcy/file/list` (Personal Host)
- Body: 包含 `parentFileId`, `pageInfo`, `orderBy`

### 4.2 创建文件夹 (create)

- `POST /hcy/file/create` (Personal Host)
- Body: `{"parentFileId": "...", "name": "...", "type": "folder"}`

### 4.3 重命名 (update)

- `POST /hcy/file/update` (Personal Host)
- Body: `{"fileId": "...", "name": "新名称"}`

### 4.4 删除到回收站 (batchTrash)

- `POST /hcy/recyclebin/batchTrash` (Personal Host)
- Body: `{"fileIds": ["..."]}`

### 4.5 任务查询 (get)

- `POST /hcy/task/get` (Personal Host)
- Body: `{"taskId": "..."}`

---

## 5. 分享与转存接口

### 5.1 获取分享详情 (getOutLinkInfoV6)

- `POST /yun-share/richlifeApp/devapp/IOutLink/getOutLinkInfoV6` (Share Host)
- 参数: `linkID`, `passwd`, `pCaID` (根目录为 'root')

### 5.2 执行批量转存 (createOuterLinkBatchOprTask)

- `POST /yun-share/richlifeApp/devapp/IBatchOprTask/createOuterLinkBatchOprTask` (Share Host)
- Body: `{"createOuterLinkBatchOprTaskReq": {"msisdn": "手机号", "linkID": "...", "taskInfo": {"newCatalogID": "目标ID", "contentInfoList": ["parentID/fileID"], ...}}}`

---

## 6. 接口监控与排错

### 6.1 响应日志广播

系统在底层 `doRequest` 逻辑中统一拦截了原始 JSON 响应。

- **广播机制**: 使用结构化日志系统将原始响应体发出
- **等级控制**: 归类为 `DEBUG` 等级，生产配置 (`LOG_LEVEL=INFO`) 下不打印
- **排错模式**: 设置 `LOG_LEVEL=DEBUG` 查看原始 JSON
- **关键日志标识**: `139 API 响应`

### 6.2 常见错误诊断

- 若响应包含 `05050009`，通常意味着 `Authorization` 头部失效，需重新抓包
- 容量异常: 家庭空间容量若计算不准，请参考 `client.go` 中的累加逻辑

### 6.3 人性化错误清洗与分类

#### 致命错误判定 ([Fatal])

当驱动检测到不可恢复的业务错误时，返回带有 `[Fatal]` 前缀的错误信息。

**139 致命码**:
| 错误码 | 含义 |
|--------|------|
| `200000727` | 链接不存在 |
| `200000728` / `9188` | 提取码错误或未提供 |
| `200000732` | 链接已过期 |

#### 交互逻辑联动

- **阻断**: 后端 `runTask` 会拦截带有 `[Fatal]` 标记的任务，防止无效请求
- **警示**: 前端 UI 会将此类任务标记为红色 "LINK ERROR"，并禁用运行按钮
- **解封**: 用户通过"编辑并保存"操作，可强制重置任务状态，解除 `[Fatal]` 锁定
