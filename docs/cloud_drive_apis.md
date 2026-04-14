# 云盘底层 API 接口调用手册 (Cloud Drive APIs)

本文档整理并固化了 `cloudpan-auto-save` (移动云盘 139) 和 `quark-auto-save` (夸克网盘) 中使用到的所有底层网盘接口逻辑。这些接口已作为本项目核心驱动的基础，固化此处以备后续持续开发、接口扩展及维护。

---

## 1. 移动云盘 (139) API

移动云盘采用多子域名架构，接口分为基础信息、用户节点、分享节点以及私有文件（HCY）节点。

### 1.1 基础信息与认证
- **主要域名**:
  - 基础/会员: `https://yun.139.com`
  - 用户管理: `https://user-njs.yun.139.com`
  - 分享/转存: `https://share-kd-njs.yun.139.com`
  - 私有文件 (HCY): `https://personal-kd-njs.yun.139.com`
- **认证方式**:
  - `Authorization`: `Basic [base64(pc:手机号:token)]`
  - `Cookie`: 浏览器登录后的 Cookie。
- **签名机制 (mcloud-sign)**:
  用于 HCY 私有接口。格式为 `datetime,randomStr,hash`。
  - `hash` 计算：对 JSON Body 进行序列化 -> URI 编码 -> 字符排序 -> Base64 编码 -> MD5 哈希，再与时间戳+随机串的 MD5 进行二次哈希。

### 1.2 账号与用户接口
- **获取用户信息 (getUser)**: 
  - `POST /user/getUser` (User Host)
  - **重要返回结构**:
    - `userDomainId`: 用户核心标识（容量查询必填）。
    - `userProfileInfo.userName`: 最新的昵称字段所在路径。
    - `auth.memberLevel`: 部分版本在此处返回会员等级。
    - `loginName / account / msisdn / phoneNumber`: 用户的真实手机号。
    - `userServiceType`: 用户服务类型标识（如 "8" 代表移动云盘会员）。
- **获取云盘配额 (getPersonalDiskInfo)**: 
  - `POST /user/disk/getPersonalDiskInfo` (User Host)
  - Body: `{"userDomainId": "xxx"}`
  - 返回: `diskSize`, `freeDiskSize` (MB)。
- **获取会员等级 (queryUserBenefits)**: 
  - `POST /orchestration/group-rebuild/member/v1.0/queryUserBenefits` (Base Host)
  - **鉴权要求**: 必须携带 `mcloud-sign` 签名（基于 Body 计算）。
  - Body: `{"isNeedBenefit": 1, "commonAccountInfo": {"account": "手机号", "accountType": 1}}`
  - **返回解析**: 
    - 会员等级从 `data.userSubMemberList[0].memberLvName` 获取（如："白银会员"）。
    - **非会员时**: `data.userSubMemberList` 为空数组 `[]`。

### 1.3 核心鉴权与异常处理
- **动态签名 (mcloud-sign)**:
  用于 HCY 及 Orchestration 系列接口。
  - **Orchestration 接口签名**: 必须基于请求 Body、当前时间戳及 16 位随机字符串计算二次 MD5 哈希。
- **Token 异常状态码**:
  - `05050009`: 非法的 Token（通常指 Authorization 已过期或错误）。
  - `1010010003`: 登录已失效。
  - `1010010014`: 签名校验失败。

### 1.3 私有文件操作 (HCY 系列)
- **文件列表 (list)**: 
  - `POST /hcy/file/list` (Personal Host)
  - Body: 包含 `parentFileId`, `pageInfo`, `orderBy`。
- **创建文件夹 (create)**: 
  - `POST /hcy/file/create` (Personal Host)
  - Body: `{"parentFileId": "...", "name": "...", "type": "folder"}`
- **重命名 (update)**: 
  - `POST /hcy/file/update` (Personal Host)
  - Body: `{"fileId": "...", "name": "新名称"}`
- **删除到回收站 (batchTrash)**: 
  - `POST /hcy/recyclebin/batchTrash` (Personal Host)
  - Body: `{"fileIds": ["..."]}`
- **任务查询 (get)**: 
  - `POST /hcy/task/get` (Personal Host)
  - Body: `{"taskId": "..."}`

### 1.4 分享与转存接口
- **获取分享详情 (getOutLinkInfoV6)**: 
  - `POST /yun-share/richlifeApp/devapp/IOutLink/getOutLinkInfoV6` (Share Host)
  - 参数: `linkID`, `passwd`, `pCaID` (根目录为 'root')。
- **执行批量转存 (createOuterLinkBatchOprTask)**: 
  - `POST /yun-share/richlifeApp/devapp/IBatchOprTask/createOuterLinkBatchOprTask` (Share Host)
  - Body: `{"createOuterLinkBatchOprTaskReq": {"msisdn": "手机号", "linkID": "...", "taskInfo": {"newCatalogID": "目标ID", "contentInfoList": ["parentID/fileID"], ...}}}`

---

## 2. 夸克网盘 (Quark) API

夸克网盘分为 PC 端网页接口和 App 移动端模拟接口。

### 2.1 基础信息与认证
- **域名**:
  - PC 端: `https://drive-pc.quark.cn`
  - 移动端: `https://drive-m.quark.cn`
- **认证方式**: 
  - 主要是 HTTP `Cookie`。
  - 调用移动端接口（容量、签到）需从 Cookie 中提取 `kps`, `sign`, `vcode` 作为 URL 参数。

### 2.2 账号与用户接口
- **获取基础信息 (昵称)**: 
  - `GET https://pan.quark.cn/account/info`
- **会员与容量信息 (PC Web端 - 最新推荐)**:
  - `GET https://pan.quark.cn/1/clouddrive/member?pr=ucpro&fr=pc`
  - **重要返回结构**:
    - `data.total_capacity`: 总空间 (Bytes)。
    - `data.use_capacity`: 已用空间 (Bytes)。
    - `data.member_type`: 会员类型 (字符串，如 "NORMAL", "SUPER_VIP")。
- **备选容量接口 (PC Web端)**:
  - `GET https://drive-pc.quark.cn/1/clouddrive/capacity?pr=ucpro&fr=pc`
  - 返回: `data.cap_info.total`, `data.cap_info.used`。
- **会员容量信息 (App端 - 需鉴权参数)**: 
  - `GET /1/clouddrive/capacity/growth/info` (App Host)
  - 参数: `pr=ucpro&fr=android&kps=...&sign=...&vcode=...`
  - 返回: `total_capacity`, `used_capacity`, `member_type`。
- **每日签到 (App端)**: 
  - `POST /1/clouddrive/capacity/growth/sign` (App Host)
  - Body: `{"sign_cyclic": true}`

### 2.3 文件操作接口 (PC端)
- **获取目录列表 (sort)**: 
  - `GET /1/clouddrive/file/sort`
  - 参数: `pdir_fid` (根目录为 "0"), `_page`, `_size`, `_sort`。
- **创建文件夹 (mkdir)**: 
  - `POST /1/clouddrive/file`
  - Body: `{"pdir_fid": "...", "file_name": "...", "dir_path": "..."}`
- **重命名 (rename)**: 
  - `POST /1/clouddrive/file/rename`
  - Body: `{"fid": "...", "file_name": "..."}`
- **批量删除 (delete)**: 
  - `POST /1/clouddrive/file/delete`
  - Body: `{"action_type": 2, "filelist": ["fid1", "fid2"]}`
- **路径解析 (path_list)**: 
  - `POST /1/clouddrive/file/info/path_list`
  - 将绝对路径数组转换为对应的 `fid` 详情。

### 2.4 分享与转存接口 (PC端)
- **获取分享 Token (token)**: 
  - `POST /1/clouddrive/share/sharepage/token`
  - Body: `{"pwd_id": "...", "passcode": "..."}`
  - 返回: `stoken`。
- **获取分享页详情 (detail)**: 
  - `GET /1/clouddrive/share/sharepage/detail`
  - 参数: `pwd_id`, `stoken`, `pdir_fid`。
- **执行保存 (save)**: 
  - `POST /1/clouddrive/share/sharepage/save`
  - Body: `{"fid_list": [...], "fid_token_list": [...], "to_pdir_fid": "目标ID", "pwd_id": "...", "stoken": "..."}`
- **查询异步任务 (task)**: 
  - `GET /1/clouddrive/task`
  - 参数: `task_id`。状态 `2` 表示成功。
