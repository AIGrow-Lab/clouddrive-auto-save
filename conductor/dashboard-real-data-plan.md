# 仪表盘数据真实化优化计划 (Dashboard Real Data Integration)

## 1. 目标 (Objective)
移除前端 Dashboard.vue 中写死的假数据，通过开发后端统计接口并对接前端，实现仪表盘运行状态和实时动态的真实数据展示。

## 2. 涉及的关键文件 (Key Files & Context)
*   **后端 API 路由:** `internal/api/router.go`
*   **后端模型:** `internal/db/db.go`
*   **前端 API 请求:** `web/src/api/dashboard.js` (新增)
*   **前端视图:** `web/src/views/Dashboard.vue`

## 3. 实施步骤 (Implementation Steps)

### 3.1 后端接口开发 (Backend Development)
1.  **新增统计接口函数 `getDashboardStats`:** 在 `internal/api/router.go` 中实现统计逻辑。
2.  **注册路由:** 在 `api` 组中注册 `GET /dashboard/stats` 路由。
3.  **计算逻辑:**
    *   **运行中任务:** `SELECT COUNT(*) FROM tasks WHERE status = 'running'`
    *   **已保存容量:** 查询所有状态正常账号的已用容量总和: `SELECT SUM(capacity_used) FROM accounts WHERE status = 1`
    *   **今日完成:** `SELECT COUNT(*) FROM tasks WHERE status = 'success' AND DATE(last_run) = DATE('now')`
    *   **活跃账号:** `SELECT COUNT(*) FROM accounts WHERE status = 1`
    *   **实时动态:** `SELECT * FROM tasks ORDER BY last_run DESC LIMIT 5` (取最近的 5 次任务执行记录)

### 3.2 前端 API 对接 (Frontend API Integration)
1.  **新建 API 文件:** 在 `web/src/api/` 下创建 `dashboard.js`，导出 `getStats` 方法，使用项目中现有的请求实例。
2.  **修改 Dashboard 视图:**
    *   在 `web/src/views/Dashboard.vue` 中引入 `getStats` 和 Vue 响应式 API。
    *   在 `onMounted` 钩子中获取数据。
    *   将模板中的写死数字替换为绑定的响应式变量。
    *   实现一个简单的文件大小格式化方法将字节转换为 TB/GB。

## 4. 验证与测试 (Verification & Testing)
1.  重新编译运行后端程序和前端开发服务器。
2.  刷新 Dashboard 页面。
3.  验证页面上的四项关键指标数据是否与本地 sqlite 数据库的实际记录相符。