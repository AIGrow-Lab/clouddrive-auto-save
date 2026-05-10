# OpenList 扫描 API 使用说明

本文档说明了如何通过 OpenList API 触发文件扫描，以便快速生成本地 strm 文件。

## 1. 基础信息
- **服务地址**: 默认 `http://127.0.0.1:23541`（根据你的 OpenList 部署地址配置）
- **认证方式**: 使用 `Authorization` Header 传递 Token
- **Token 格式**: `openlist-{uuid}{random_string}`

## 2. 触发扫描接口

立即触发 OpenList 文件扫描，扫描完成后会自动生成 strm 文件。

- **Endpoint**: `/api/admin/scan/start`
- **Method**: `POST`
- **认证**: `Authorization` Header

### 请求示例

```bash
curl --location --request POST 'http://127.0.0.1:23541/api/admin/scan/start' \
--header 'Authorization: openlist-913f51c8-720f-45ec-b1a2-a388ff5c50f4mTNNbFg6GHoVd6sycC9r0qgwzLU27ps1Nz8yrRtlSlOiaNSNeA1dAZZBtWgXhqhQ'
```

### 响应说明

- **成功**: HTTP 200 OK
- **失败**: HTTP 4xx/5xx + 错误信息

## 3. 在 UCAS 中的集成方式

### 3.1 配置项

通过 UCAS 设置页面配置以下参数：

| 配置项 | 说明 | 示例值 |
| :--- | :--- | :--- |
| 启用自动扫描 | 是否启用转存完成后自动触发扫描 | 开启/关闭 |
| API 地址 | OpenList 服务地址 | `http://127.0.0.1:23541` |
| API Token | 认证 Token | `openlist-913f51c8-...` |

### 3.2 自动触发逻辑

- **单任务执行**: 任务完成且有新转存文件时，自动触发扫描
- **批量任务执行**: 所有任务完成后，统一触发一次扫描
- **延迟合并**: 3 秒内的多次触发会合并为一次扫描，避免频繁请求

### 3.3 手动触发

在设置页面点击「手动扫描」按钮，立即触发一次扫描。

## 4. 关键特性说明

1. **自动触发**: 转存任务完成后自动通知 OpenList 扫描新文件
2. **延迟合并**: 避免短时间内多次扫描，减少 OpenList 服务压力
3. **批量优化**: 批量任务只触发一次扫描，提高效率
4. **错误隔离**: 扫描失败不影响转存任务本身

## 5. 项目集成参考

在本项目中，OpenList 相关的逻辑主要位于：
- **客户端**: `internal/core/openlist/client.go`
- **扫描管理器**: `internal/core/openlist/scanner.go`
- **设置界面**: `web/src/views/Settings.vue`
- **API 接口**: `POST /api/openlist/scan`
