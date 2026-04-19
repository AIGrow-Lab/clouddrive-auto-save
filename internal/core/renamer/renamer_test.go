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

func TestProcessor_Process_MagicVariables(t *testing.T) {
	p := NewProcessor()
	tests := []struct {
		name     string
		fileName string
		replace  string
		want     string
	}{
		{"YEAR", "Movie.2024.mp4", "{YEAR}", "2024"},
		{"DATE", "Doc.2024-04-19.pdf", "{DATE}", "20240419"},
		{"EXT", "video.mkv", "file.{EXT}", "file.mkv"},
		{"CHINESE", "电影.Movie.mp4", "{CHINESE}", "电影"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := RenameOptions{
				FileName:    tt.fileName,
				Replacement: tt.replace,
			}
			got, err := p.Process(opts)
			if err != nil {
				t.Fatalf("Process failed: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
