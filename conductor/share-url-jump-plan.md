# 添加分享链接快捷跳转按钮计划 (Share URL Quick Jump Plan)

## 1. Objective (目标)
在“创建/编辑任务”弹窗中，为“分享链接”输入框增加一个附加按钮。点击该按钮，系统将在浏览器的新标签页中打开当前填写的分享链接，方便用户快速核对网盘内容。

## 2. Key Files & Context (核心影响文件)
- `web/src/views/Tasks.vue`
  - 第 106 行附近的 `<el-input v-model="form.share_url">`。

## 3. Implementation Steps (实施步骤)
1. **引入新图标**: 在 `Tasks.vue` 的 `<script setup>` 中，从 `lucide-vue-next` 库中额外引入 `ExternalLink` (外链) 图标。
2. **重构输入框**: 将分享链接的 `el-input` 改造成带有附加组件的形式。
   - 使用 Vue 3 的插槽 `<template #append>`。
   - 在插槽内添加一个 `el-button`，当 `form.share_url` 有值时，点击该按钮通过 `window.open(form.share_url, '_blank')` 打开新标签页。
   - 若链接为空，按钮处于禁用状态。

## 4. Verification & Testing (验证与测试)
1. 刷新前端页面，打开任务编辑框。
2. 确认输入分享链接后，右侧出现一个带有外链图标的跳转按钮。
3. 点击该按钮，浏览器成功新开标签页跳转至该分享链接。