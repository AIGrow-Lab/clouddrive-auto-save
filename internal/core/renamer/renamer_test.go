package renamer

import (
	"testing"
)

func TestProcessor_Process_TaskName(t *testing.T) {
	p := NewProcessor()
	opts := RenameOptions{
		TaskName:    "MyTask",
		FileName:    "movie.mp4",
		Replacement: "{TASKNAME}.mp4",
	}
	got, err := p.Process(opts)
	if err != nil {
		t.Fatalf("Process failed: %v", err)
	}
	want := "MyTask.mp4"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
