# 系统设置隐藏与下一阶段规划

**目标：** 在当前版本中隐藏尚不完善的“系统设置”侧边栏入口，并将其全功能重构列入后续路线图。

**修改内容：**

### 1. 前端：移除侧边栏菜单项
- **文件**：`web/src/layout/MainLayout.vue`
- **动作**：
  - 移除 `<el-menu-item index="/settings">`。
  - 从 `currentPageTitle` 计算属性中移除 `'/settings'` 映射。
  - 从 `import` 中移除不再使用的 `Settings` 图标。

### 2. 文档：同步 Roadmap
- **文件**：`RESTRUCTURE_PLAN.md`
- **动作**：
  - 在第五阶段明确“系统设置中心”的开发点：包含通知推送配置、Worker 并发数调整、自动清理策略、AI API Key 管理。

### 3. 构建与验证
- 重新构建前端并运行，确保侧边栏不再显示“系统设置”。
