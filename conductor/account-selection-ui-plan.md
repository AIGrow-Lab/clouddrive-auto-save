# 任务账号选择视觉增强计划 (Account Selection UI Enhancement Plan)

## 1. Objective (目标)
优化“创建/编辑任务”弹窗中执行账号的选择体验。通过视觉分组（139 vs Quark）和图标化，让用户能更直观地识别并选择正确的执行账号。

## 2. Key Files & Context (核心影响文件)
- `web/src/views/Tasks.vue`: 任务管理主页面，负责渲染对话框及账号下拉列表。

## 3. Implementation Steps (实施步骤)

### Step 3.1: 准备计算属性
1. 在 `Tasks.vue` 脚本部分，新增计算属性 `groupedAccounts`。
2. 该属性将原始的 `accounts` 列表按照 `platform` (139 和 quark) 进行归类，生成适合 `el-option-group` 渲染的数据结构。

### Step 3.2: 模板层重构 (账号下拉框)
1. 定位到执行账号的 `el-select` 部分。
2. 使用 `el-option-group` 替换单层的 `el-option` 循环。
3. 自定义选项内容（Slot）：
   - 在选项前根据平台显示不同的颜色图标或平台名称。
   - 增强选项的展示信息，包含昵称、剩余空间（如有）。
   - 增加状态标识（如果账号失效，禁用选项并显示“失效”）。

### Step 3.3: 样式美化
1. 在 `<style scoped>` 中增加针对账号下拉项的布局样式。
2. 确保平台分组标题美观，选项对齐。

## 4. Verification & Testing (验证与测试)
- **视觉验证**: 打开创建任务弹窗，检查下拉列表是否已按平台（移动云盘 / 夸克）清晰分组。
- **功能验证**:
  - 确认点击选项后能正确选中 `account_id`。
  - 确认禁用状态（失效账号）无法被选中。
  - 确认切换账号时，原有的 `handleAccountChange` 逻辑依然生效。
