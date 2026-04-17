# 修复账号管理界面初始化闪烁计划 (Fix Accounts UI Flicker Plan)

## 1. Objective (目标)
修复进入账号管理界面时，由于异步数据加载导致的“暂无账号”提示短暂闪烁的问题，确保加载期间始终显示 Loading 状态。

## 2. Key Files & Context (核心影响文件)
- `web/src/views/Accounts.vue`: 账号管理视图文件。

## 3. Root Cause Analysis (根本原因分析)
1. **Typo 错误**: 在 `fetchList` 函数中，代码错误地写成了 `loading.ref = true`，这在 Vue 3 中是无效的（应为 `loading.value`），导致 Loading 状态从未被正确激活。
2. **初始状态竞争**: `loading` 变量初始值为 `false`。这意味着在 `onMounted` 触发的请求发出前，界面处于“非加载且列表为空”的状态，从而触发了 `el-empty` 的显示。
3. **空状态判断不严谨**: `el-empty` 仅检查了 `accountList.length === 0`，没有结合 `loading` 状态。

## 4. Implementation Steps (实施步骤)
1. **修正初始状态**: 将 `loading` 响应式变量的初始值改为 `true`。
2. **修复逻辑错误**: 在 `fetchList` 函数中，将 `loading.ref = true` 修正为 `loading.value = true`。
3. **优化空状态渲染**: 修改模板中的 `el-empty` 判定条件，增加 `!loading` 的前置检查。

## 5. Verification & Testing (验证与测试)
1. 刷新浏览器并切换到账号管理页面。
2. 确认在数据返回前，界面显示 Loading 动画。
3. 确认在加载完成后：
   - 若有数据，显示列表/卡片。
   - 若无数据，才显示“暂无账号”。
