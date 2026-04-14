# 修复 139 云盘家庭空间已用容量重复计算问题

## 1. Objective (目标)
修复 `internal/core/cloud139/client.go` 中对于 139 云盘总已用容量的计算逻辑，避免将家庭空间的容量状态错误累加到已用容量中，导致容量显示不准确。

## 2. Key Files & Context (关键文件与上下文)
- `internal/core/cloud139/client.go`: 包含 139 云盘账号信息及容量拉取的逻辑。
- 现状：在分别拉取个人空间（`getPersonalDiskInfo`）和家庭空间（`getFamilyDiskInfo`）时，会将两者的总容量（`DiskSize`）和已用容量（计算出的 `total - free`）都进行了累加。
- 问题：由于家庭空间的真实使用状态在 139 的机制下并不准确反映在该接口的 `usedSize` 或 `freeDiskSize` 中，实际上已用容量仅需要通过个人的 `freeDiskSize` 来倒推（允许为负数），家庭空间的接口仅用于扩充总容量（`totalCapacity`）。

## 3. Implementation Steps (实施步骤)
修改 `internal/core/cloud139/client.go` 中针对 `getFamilyDiskInfo` 接口响应的处理逻辑：

1. 保持原有的拉取 `getFamilyDiskInfo` 逻辑不变。
2. 在解析该接口响应时，仅提取 `DiskSize` 并累加到 `totalCapacity` 中。
3. 移除根据 `FreeDiskSize` 计算 `usedCapacity += (total - free) * 1024 * 1024` 的行。
4. 清理对家庭空间中无需使用的 `FreeDiskSize` 的相关解析或赋值，保持代码整洁。

## 4. Verification & Testing (验证与测试)
- 编译并运行项目，查看包含 139 家庭空间的账号容量显示。
- 确认界面的已用容量准确反映为“个人总容量 - 个人剩余容量”。
- 确认总容量包含个人和家庭空间的相加之和。
- 验证当个人剩余容量为负数时，整体容量的计算仍然准确。