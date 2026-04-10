# 修复 Dashboard 导入路径问题 (Fix Dashboard Import Path)

## 1. 目标 (Objective)
修复前端 `web/src/views/Dashboard.vue` 文件中对于 `@/api/dashboard` 的解析错误。

## 2. 涉及的关键文件 (Key Files & Context)
*   **前端视图:** `web/src/views/Dashboard.vue`

## 3. 实施步骤 (Implementation Steps)
1.  **修改导入路径:** 将 `web/src/views/Dashboard.vue` 中第 97 行左右的引入语句 `import { getStats } from '@/api/dashboard'` 替换为使用相对路径的 `import { getStats } from '../api/dashboard'`。

## 4. 验证与测试 (Verification & Testing)
1.  保存修改后，Vite 将自动热重载。
2.  观察浏览器控制台是否不再报错，且仪表盘能正确渲染真实数据。