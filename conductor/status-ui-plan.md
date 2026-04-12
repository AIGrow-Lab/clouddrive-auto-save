# 账号状态 UI 优化实施计划 (Account Status UI Optimization Plan)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 优化账号列表中的“状态”列展示效果，将原本不协调的 `el-badge`（带右上角小圆点）替换为“浅色胶囊标签 (Light Tag)”，使界面更加现代和柔和。

**Architecture:** 
1. 在 `web/src/views/Accounts.vue` 中，将 `<el-badge>` 替换为 `<el-tag type="success|danger" effect="light" round class="status-tag">`。
2. 将该列内容居中对齐，并增加一点自定义的内边距样式以显得更饱满。

**Tech Stack:** Vue 3, Element Plus, CSS

---

### Task 1: 替换状态列模板

**Files:**
- Modify: `web/src/views/Accounts.vue`

- [ ] **Step 1: 替换 HTML 模板**
  找到“状态”列的 `<el-table-column>`，将其内容替换为圆角、浅色效果的 `el-tag`。

  ```vue
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" effect="light" round class="status-tag">
              {{ row.status === 1 ? '正常' : '失效' }}
            </el-tag>
          </template>
        </el-table-column>
  ```

- [ ] **Step 2: 添加 CSS 样式**
  在 `<style scoped>` 中为 `.status-tag` 添加稍微宽一点的内边距和加粗字体，以增强胶囊的质感。

  ```css
  .status-tag {
    padding: 0 12px;
    font-weight: 500;
  }
  ```

### Task 2: 验证与提交

- [ ] **Step 1: 验证展示效果**
  重新启动或热更新前端项目，检查状态列是否变成了圆角的浅绿/浅红胶囊标签。

- [ ] **Step 2: 提交代码**
  ```bash
  git add web/src/views/Accounts.vue
  git commit -m "style(web): 优化账号状态列 UI 为浅色胶囊标签"
  ```