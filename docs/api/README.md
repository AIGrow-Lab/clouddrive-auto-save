# UCAS 后端 API 接口文档

本文档详细说明了 **统一云盘自动转存系统 (UCAS)** 后端服务提供的所有 REST API 接口。

## 基础信息
- **Base URL**: `http://localhost:8080/api`
- **Content-Type**: `application/json`
- **认证方式**: 暂无（内网环境使用），后续计划增加 API Key。

## 模块导航
1. [账号管理 (Accounts)](./accounts.md) - 139/Quark 账号的增删改查与校验。
2. [任务管理 (Tasks)](./tasks.md) - 转存任务的生命周期管理与手动触发。

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
