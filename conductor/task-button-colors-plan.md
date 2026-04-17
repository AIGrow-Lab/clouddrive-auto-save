# 优化任务列表操作按钮颜色计划 (Optimize Task Actions Button Colors Plan)

## 1. Objective (目标)
优化任务列表操作栏中“运行”、“编辑”和“删除”三个按钮的颜色展示，使其更加符合操作语义并提高用户的视觉区分度。

## 2. Key Files & Context (核心影响文件)
- `web/src/views/Tasks.vue`
  - 第 52-64 行附近的 `<el-button-group>` 中的按钮配置。

## 3. Implementation Steps (实施步骤)
目前“运行”和“编辑”按钮都是 `type="primary"`（蓝色），这在视觉上不够清晰。
- **运行 (Run)**：作为一个积极的执行操作，将其 `type` 从 `primary` 改为 `success`（绿色），以传达“开始、成功、正向”的语意。
- **编辑 (Edit)**：保持 `type="primary"`（蓝色），符合信息修改的常规配色。
- **删除 (Delete)**：保持 `type="danger"`（红色），符合破坏性操作的常规配色。

## 4. Verification & Testing (验证与测试)
1. 刷新前端页面，查看任务列表的“操作”列。
2. 确认三个按钮的颜色依次为：绿色（运行）、蓝色（编辑）、红色（删除），视觉区分清晰明了。
