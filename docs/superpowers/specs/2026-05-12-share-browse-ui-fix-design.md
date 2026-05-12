# 分享链接浏览功能 UI 优化设计

## 背景

当前分享链接浏览功能存在三个 UI 问题：
1. 两个按钮（浏览分享内容 + 在新标签页打开）在输入框右侧重叠
2. 浏览分享内容模式中，选择逻辑不直观（需要先进入子目录再返回选择）
3. 选择起始转存文件模式中，"进入"按钮与 radio 选择逻辑重复

## 设计方案

### 1. 按钮布局优化

**当前问题**：两个按钮都放在 `el-input` 的 `#append` 区域，导致视觉重叠。

**解决方案**：保留两个按钮，中间添加分隔线，使用 flex 布局对齐。

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
  margin: 0 6px;
}

:deep(.el-input-group__append) {
  display: flex;
  align-items: center;
  gap: 4px;
}

:deep(.el-input-group__append .el-button) {
  margin-left: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
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
- 保留 radio 列用于选择文件（仅在初始目录显示）
- 保留"进入"列用于浏览子目录
- 点击文件夹名直接进入子目录
- 双击文件选中为起始文件

```html
<!-- startFile 模式且在初始目录时显示 radio 列 -->
<el-table-column v-if="browseMode === 'startFile' && isInitialDir" width="40" align="center">
  <template #default="{ row }">
    <el-radio v-if="!row.is_folder" v-model="tempStartFileId" :label="row.id" class="naked-radio"><span></span></el-radio>
  </template>
</el-table-column>

<!-- selectShareUrl 模式显示进入按钮 -->
<el-table-column v-if="browseMode === 'selectShareUrl'" label="操作" width="80" align="center">
  <template #default="{ row }">
    <el-button v-if="row.is_folder" type="primary" link size="small" @click="enterFolder(row)">
      进入
    </el-button>
  </template>
</el-table-column>
```

### 4. 139 平台分享链接子目录支持

**当前问题**：移动云盘（139）的分享链接 URL 不支持通过参数区分目录。

**解决方案**：
- 后端 Task 模型新增 `share_parent_id` 字段
- 前端表单新增 `share_parent_id` 字段
- 选择 139 子文件夹时，存储 `share_parent_id`
- 打开选择起始文件弹窗时，使用 `share_parent_id` 作为初始目录
- 浏览分享内容时，以当前分享链接对应的目录为起始

## 模式差异对比

| 特性 | selectShareUrl | startFile |
|------|----------------|-----------|
| 表格 radio 列 | 无 | 有（仅在初始目录显示） |
| 操作列 | 有（进入按钮） | 无 |
| 点击文件夹 | 进入子目录 | 进入子目录 |
| 双击文件 | 无操作 | 选中为起始文件 |
| 底部按钮 | 选择当前目录（目录名） | 确认选择（需先选中文件） |
| 按钮禁用条件 | 无（始终可点击） | 未选中文件时禁用 |

## 关键文件

- `internal/db/db.go` - Task 模型新增 share_parent_id 字段
- `internal/api/router.go` - API 层支持 share_parent_id
- `web/src/views/Tasks.vue` - 前端主要修改文件

## 验证

1. 按钮布局：两个按钮清晰分隔，图标居中对齐
2. 浏览分享内容模式：
   - 点击文件夹进入子目录
   - 底部按钮显示当前目录名
   - 点击按钮后分享链接更新为当前目录地址
   - 夸克网盘从 URL 解析当前目录，139 使用 share_parent_id
3. 选择起始转存文件模式：
   - 点击文件夹名进入子目录
   - 双击文件选中为起始文件
   - radio 仅在初始目录显示
   - 139 平台支持 share_parent_id 作为初始目录
