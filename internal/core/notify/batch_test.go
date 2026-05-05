package notify

import (
	"strings"
	"testing"
	"time"
)

func TestBuildBatchBody_AllSuccess(t *testing.T) {
	results := []BatchResult{
		{TaskName: "任务A", Status: "success", Message: "转存成功 (新增 5 个文件)", Files: []string{"a.mp4", "b.mp4"}, Duration: 45 * time.Second},
		{TaskName: "任务B", Status: "success", Message: "转存成功 (新增 3 个文件)", Files: []string{"c.mkv"}, Duration: 32 * time.Second},
	}
	body := buildBatchBody(results, 2*time.Minute)

	if !strings.Contains(body, "总耗时") {
		t.Error("body should contain '总耗时'")
	}
	if !strings.Contains(body, "✅ 任务A") {
		t.Error("body should contain success icon for 任务A")
	}
	if !strings.Contains(body, "转存文件列表") {
		t.Error("body should contain file list section")
	}
	if !strings.Contains(body, "a.mp4") {
		t.Error("body should list file names")
	}
}

func TestBuildBatchBody_MixedResults(t *testing.T) {
	results := []BatchResult{
		{TaskName: "任务A", Status: "success", Message: "ok", Files: []string{"a.mp4"}, Duration: 10 * time.Second},
		{TaskName: "任务B", Status: "failed", Message: "解析分享失败: 链接过期", Duration: 2 * time.Second},
	}
	body := buildBatchBody(results, 30*time.Second)

	if !strings.Contains(body, "❌ 任务B") {
		t.Error("body should contain failure icon for 任务B")
	}
	if !strings.Contains(body, "链接过期") {
		t.Error("body should contain failure reason")
	}
}

func TestBuildBatchTitle_AllSuccess(t *testing.T) {
	results := []BatchResult{
		{TaskName: "A", Status: "success"},
		{TaskName: "B", Status: "success"},
	}
	title := buildBatchTitle(results)
	if title != "📊 批量转存完成: 全部 2 个任务成功" {
		t.Errorf("unexpected title: %s", title)
	}
}

func TestBuildBatchTitle_PartialFailure(t *testing.T) {
	results := []BatchResult{
		{TaskName: "A", Status: "success"},
		{TaskName: "B", Status: "failed"},
		{TaskName: "C", Status: "success"},
	}
	title := buildBatchTitle(results)
	if title != "📊 批量转存完成: 2成功 / 1失败" {
		t.Errorf("unexpected title: %s", title)
	}
}

func TestBuildBatchBody_FileListTruncation(t *testing.T) {
	files := make([]string, 25)
	for i := range files {
		files[i] = "file.mp4"
	}
	results := []BatchResult{
		{TaskName: "BigTask", Status: "success", Message: "ok", Files: files, Duration: 10 * time.Second},
	}
	body := buildBatchBody(results, 10*time.Second)

	if !strings.Contains(body, "等共 25 个文件") {
		t.Error("body should truncate file list and show total count")
	}
}
