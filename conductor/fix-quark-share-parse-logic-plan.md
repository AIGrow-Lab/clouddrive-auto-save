# 修复夸克失效分享链接误判成功计划 (Fix Quark Invalid Share Link Success Plan)

## 背景与目标 (Objective)

当前在解析夸克网盘分享链接时，如果分享链接已失效或被取消，夸克的 `sharepage/detail` API 会返回 HTTP `200 OK` 状态码，且 `data.list` 为空数组。
现有的代码逻辑仅仅反序列化了 JSON 并没有对 `list` 的长度进行校验，导致后端认为“成功解析了 0 个文件”，后续逻辑由于没有文件可转存，最终将任务状态误判为 `success` 并提示“没有需要转存的文件”。
目标是在解析分享链接的阶段，如果检测到提取的文件列表为空，则抛出一个带有 `[Fatal]` 前缀的错误，从而中断任务并将其标记为失败，同时提示用户链接可能已失效。

## 关键文件 (Key Files & Context)

- `internal/core/quark/client.go`: 包含 `ParseShare` 方法的实现。

## 实施步骤 (Implementation Steps)

1. **修改 `internal/core/quark/client.go`**：
   定位到 `ParseShare` 方法中反序列化 `detailRes` 之后的代码块。
   添加对 `detailRes.Data.List` 长度的校验：

   ```go
   if err := json.Unmarshal(resp, &detailRes); err != nil {
    return nil, err
   }

   if len(detailRes.Data.List) == 0 {
    return nil, fmt.Errorf("[Fatal] 夸克分享链接无效、已取消或包含的文件为空")
   }

   var files []core.FileInfo
   ```

## 验证与测试 (Verification & Testing)

1. 在 `internal/core/quark/client_test.go` 中补充针对解析空列表的单元测试（如果有的话），或者运行现有的测试保证没有破坏原有逻辑。
2. 前端添加一个已知失效的夸克分享链接进行真实转存测试，验证任务状态是否变为了 `failed`，并能看到明确的错误提示信息。
