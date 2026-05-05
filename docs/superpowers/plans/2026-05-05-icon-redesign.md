# 项目图标与首页视觉重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 按照设计规范重构项目的 Favicon 和首页图标，实现“层叠云朵”视觉方案及动态效果。

**Architecture:** 采用 SVG 矢量图形实现核心图标，通过 Vue 组件封装 Logo 以支持 CSS 动画（呼吸灯、流光），并替换现有的静态资源。

**Tech Stack:** SVG, Vue 3, CSS Animations.

---

### Task 1: 更新 Favicon

**Files:**
- Modify: `web/public/favicon.svg`

- [ ] **Step 1: 编写新的 Favicon SVG 代码**
将 `web/public/favicon.svg` 的内容替换为简化的层叠云朵设计。

```xml
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32">
  <defs>
    <linearGradient id="cloudGrad" x1="0%" y1="0%" x2="100%" y2="100%">
      <stop offset="0%" style="stop-color:#4facfe;stop-opacity:1" />
      <stop offset="100%" style="stop-color:#00f2fe;stop-opacity:1" />
    </linearGradient>
  </defs>
  <path fill="url(#cloudGrad)" d="M22 10c-3.3 0-6 2.7-6 6s2.7 6 6 6 6-2.7 6-6-2.7-6-6-6zm-10 0C8.7 10 6 12.7 6 16s2.7 6 6 6c.7 0 1.3-.1 1.9-.3C15 13.8 17.5 12 20.5 12c.5 0 1 .1 1.4.2C21.1 9.7 18.7 8 16 8c-3.9 0-7 3.1-7 7 0 .4 0 .9.1 1.3-.3-.2-.6-.3-1-.3-.5 0-1 .2-1.3.6C5.9 15.1 4.3 14 2.4 14c-2.1 0-3.9 1.3-4.6 3.1-.5-.7-1.3-1.1-2.2-1.1-1.1 0-1.9.9-1.9 1.9V22h0z" transform="translate(5, 2)" />
  <circle cx="21" cy="18" r="2.5" fill="white" fill-opacity="0.6" />
</svg>
```

- [ ] **Step 2: 验证 Favicon**
在浏览器中打开项目，检查标签页图标是否已更新且清晰。

- [ ] **Step 3: Commit**

```bash
git add web/public/favicon.svg
git commit -m "style: 更新项目 Favicon 为层叠云朵设计"
```

---

### Task 2: 创建 CloudLogo.vue 组件

**Files:**
- Create: `web/src/components/CloudLogo.vue`

- [ ] **Step 1: 编写组件基础代码**
创建一个可配置尺寸的 SVG 组件。

```vue
<template>
  <div class="cloud-logo" :style="{ width: size + 'px', height: size + 'px' }">
    <svg viewBox="0 0 100 100" fill="none" xmlns="http://www.w3.org/2000/svg">
      <defs>
        <linearGradient :id="id" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stop-color="#4facfe" />
          <stop offset="100%" stop-color="#00f2fe" />
        </linearGradient>
      </defs>
      <!-- 云朵主体 -->
      <path
        d="M75 65C66.7 65 60 58.3 60 50C60 41.7 66.7 35 75 35C83.3 35 90 41.7 90 50C90 58.3 83.3 65 75 65ZM42 65C32.1 65 24 56.9 24 47C24 37.1 32.1 29 42 29C44.2 29 46.3 29.4 48.2 30.1C51.9 21.6 60.3 15.8 70.1 15.8C82.8 15.8 93.1 26.1 93.1 38.8C93.1 40.3 93 41.7 92.7 43.1C91.8 42.5 90.7 42.1 89.5 42.1C87.8 42.1 86.3 42.8 85.2 43.9C82.5 38.8 77.1 35.5 70.8 35.5C63.8 35.5 57.9 39.7 55.4 45.7C53.7 43.5 51 42.1 48 42.1C44.5 42.1 41.7 44.9 41.7 48.3V65H42Z"
        :fill="'url(#' + id + ')'"
        class="main-cloud"
      />
      <!-- 核心/状态圆点 -->
      <circle cx="70" cy="55" r="8" fill="white" fill-opacity="0.7" class="core-node" />
    </svg>
  </div>
</template>

<script setup>
import { computed } from 'vue';
const props = defineProps({
  size: { type: Number, default: 100 },
  id: { type: String, default: 'cloud-grad' }
});
</script>

<style scoped>
.cloud-logo {
  display: inline-block;
  filter: drop-shadow(0 8px 16px rgba(79, 172, 254, 0.2));
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/CloudLogo.vue
git commit -m "feat(web): 新增 CloudLogo 组件基础实现"
```

---

### Task 3: 实现呼吸灯与流光动画

**Files:**
- Modify: `web/src/components/CloudLogo.vue`

- [ ] **Step 1: 添加动画样式**
在 `CloudLogo.vue` 的 `<style>` 中增加动画定义。

```css
<style scoped>
.cloud-logo {
  display: inline-block;
  filter: drop-shadow(0 8px 16px rgba(79, 172, 254, 0.2));
}

.core-node {
  animation: breathe 3s ease-in-out infinite;
}

@keyframes breathe {
  0%, 100% { fill-opacity: 0.4; transform: scale(0.95); transform-origin: 70px 55px; }
  50% { fill-opacity: 1; transform: scale(1.05); transform-origin: 70px 55px; }
}

.main-cloud {
  transition: filter 0.3s ease;
}

.cloud-logo:hover .main-cloud {
  filter: brightness(1.1);
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/CloudLogo.vue
git commit -m "style(web): 为 CloudLogo 组件添加呼吸灯动画"
```

---

### Task 4: 更新 Dashboard 首页

**Files:**
- Modify: `web/src/views/Dashboard.vue`

- [ ] **Step 1: 引入并使用 CloudLogo 组件**
替换 `Dashboard.vue` 中原本用于展示 logo 或 hero 的部分。

```javascript
// ... existing imports ...
import CloudLogo from '../components/CloudLogo.vue';
```

修改模版，将原本的图片或简单的文字标题替换为 `CloudLogo`。

```html
<template>
  <!-- 找到头部或欢迎区域 -->
  <div class="welcome-section">
    <CloudLogo :size="120" />
    <h1>统一云盘自动转存系统</h1>
    <!-- ... -->
  </div>
</template>
```

- [ ] **Step 2: 移除旧资源引用**
检查并移除对 `web/src/assets/hero.png` 的引用及相关样式。

- [ ] **Step 3: 验证**
启动开发服务器 (`make dev-web`)，确认首页视觉效果符合预期，且动画运行平滑。

- [ ] **Step 4: Commit**

```bash
git add web/src/views/Dashboard.vue
git commit -m "style(web): 在仪表盘首页集成新的 CloudLogo 组件"
```

---

### Task 5: 最终检查与推送

- [ ] **Step 1: 运行 Lint 检查**
Run: `make lint`

- [ ] **Step 2: 推送更改**
Run: `git push origin main`
