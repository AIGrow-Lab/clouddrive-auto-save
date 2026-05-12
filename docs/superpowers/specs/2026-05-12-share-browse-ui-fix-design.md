# 分享链接浏览功能 UI 优化设计

## 背景

当前分享链接浏览功能存在三个 UI 问题：
1. 两个按钮（浏览分享内容 + 在新标签页打开）在输入框右侧重叠
2. 浏览分享内容模式中，选择逻辑不直观（需要先进入子目录再返回选择）
3. 选择起始转存文件模式中，"进入"按钮与 radio 选择逻辑重复

## 设计方案

### 1. 按钮布局优化

**当前问题**：两个按钮都放在 `el-input` 的 `#append` 区域，导致视觉重叠。

**解决方案**：保留两个按钮，中间添加分隔线。

```html
<template #append>
  <el-button :icon="FolderOpen" title="浏览分享内容并选择目录" ... />
  <div class="append-divider"></div>
  <el-button :icon="ExternalLink" title="在新标签页中打开链接" ... />
</template>
```

CSS:
```css
.append-divider {
  width: 1px;
  height: 20px;
  background: var(--el-border-color);
  margin: 0 4px;
}
```

### 2. 浏览分享内容模式（selectShareUrl）

**当前问题**：
- 需要先选中某个文件夹（radio），再点确认
- 左侧 radio 圆圈在浏览模式中无意义

**解决方案**：
- 移除表格中的 radio 列
- 表格仅用于浏览目录内容
- 底部按钮动态显示"选择当前目录（目录名）"
- 点击文件夹进入子目录，按钮自动更新为当前目录名

```html
<template #footer>
  <el-button @click="startFileDialogVisible = false">取消</el-button>
  <el-button type="primary" @click="confirmSelectShareUrl">
    选择当前目录（{{ currentDirName }}）
  </el-button>
</template>
```

其中 `currentDirName` 根据 `breadcrumbs` 计算：
- 根目录时显示"根目录"
- 子目录时显示当前目录名（breadcrumbs 最后一项的 name）

### 3. 选择起始转存文件模式（startFile）

**当前问题**：
- 已选中文件夹后，"进入"按钮仍然显示，造成混淆
- radio 选择与"进入"操作并存，逻辑不清晰

**解决方案**：
- 移除"进入"列
- 移除 radio 列
- 点击文件夹名直接进入子目录
- 双击文件选中为起始文件

```html
<el-table-column label="文件名" show-overflow-tooltip min-width="200">
  <template #default="{ row }">
    <div class="name-main" :class="{ 'folder-clickable': row.is_folder }" 
         @click="row.is_folder && enterFolder(row)"
         @dblclick="!row.is_folder && selectStartFile(row)">
      <el-icon size="16">
        <Folder v-if="row.is_folder" color="#eab308" />
        <File v-else color="#64748b" />
      </el-icon>
      <span>{{ row.name }}</span>
    </div>
  </template>
</el-table-column>
```

## 模式差异对比

| 特性 | selectShareUrl | startFile |
|------|----------------|-----------|
| 表格 radio 列 | 无 | 无 |
| 操作列 | 无 | 无 |
| 点击文件夹 | 进入子目录 | 进入子目录 |
| 双击文件 | 无操作 | 选中为起始文件 |
| 底部按钮 | 选择当前目录（目录名） | 确认选择（需先双击选中文件） |
| 按钮禁用条件 | 无（始终可点击） | 未选中文件时禁用 |

## 关键文件

- `web/src/views/Tasks.vue` - 主要修改文件

## 验证

1. 按钮布局：两个按钮清晰分隔，无重叠
2. 浏览分享内容模式：
   - 点击文件夹进入子目录
   - 底部按钮显示当前目录名
   - 点击按钮后分享链接更新为当前目录地址
3. 选择起始转存文件模式：
   - 点击文件夹名进入子目录
   - 双击文件选中为起始文件
   - 底部按钮显示选中的文件名
