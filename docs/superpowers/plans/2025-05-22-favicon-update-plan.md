# Favicon 更新实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将项目 Favicon 替换为现代化的层叠云朵设计，优化视觉效果并减小资源体积。

**Architecture:** 直接替换公共静态资源目录中的 SVG 文件。

**Tech Stack:** SVG, Git

---

### Task 1: 更新 Favicon 文件

**Files:**
- Modify: `web/public/favicon.svg`

- [ ] **Step 1: 写入新的 SVG 内容**

```bash
cat <<EOF > web/public/favicon.svg
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
EOF
```

- [ ] **Step 2: 验证文件内容与大小**

Run: `ls -l web/public/favicon.svg && cat web/public/favicon.svg`
Expected: 文件大小显著减小（约 800 字节），内容匹配。

- [ ] **Step 3: 提交更改**

```bash
git add web/public/favicon.svg
git commit -m "style: 更新项目 Favicon 为层叠云朵设计"
```
