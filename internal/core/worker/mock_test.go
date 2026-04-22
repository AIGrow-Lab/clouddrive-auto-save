package worker

import (
	"github.com/zcq/clouddrive-auto-save/internal/core"
)

// 复用 core 中的 MockDriver 供 worker 测试使用
type MockDriver = core.MockDriver
