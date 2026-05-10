# UCAS 后端 API 接口文档

本文档详细说明了 **统一云盘自动转存系统 (UCAS)** 后端服务提供的所有 REST API 接口。

## 基础信息

- **Base URL**: `http://localhost:8080/api`
- **Content-Type**: `application/json`
- **认证方式**: 暂无（内网环境使用），后续计划增加 API Key。

## 模块导航

1. [仪表盘统计 (Dashboard)](./dashboard.md) - 系统运行状态、容量汇总与实时动态。
2. [账号管理 (Accounts)](./accounts.md) - 139/Quark 账号的增删改查与校验。
3. [任务管理 (Tasks)](./tasks.md) - 转存任务的生命周期管理与手动触发。
4. [系统设置 (Settings)](./settings.md) - 全局调度规则、通知推送与 OpenList 扫描配置。
5. [数据库设计 (Database Schema)](./database.md) - 核心数据表结构与字段说明。
6. [Bark 消息推送 (Bark API)](../bark_api.md) - 外部通知推送接口与参数说明。
7. [OpenList 扫描 (OpenList API)](../openlist_api.md) - 文件扫描触发接口与配置说明。

## 全局响应格式

所有接口均返回 JSON 格式数据。

### 成功响应

```json
{
  "id": 1,
  "platform": "quark",
  ...
}
```

### 错误响应

```json
{
  "error": "错误描述信息"
}
```
