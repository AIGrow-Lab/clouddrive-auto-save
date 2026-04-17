package scheduler

import (
	"testing"
)

func TestScheduler_AddAndRemoveTask(t *testing.T) {
	s := New()
	s.Start()
	defer s.Stop()

	err := s.UpdateTask(1, "0 * * * * *")
	if err != nil {
		t.Fatalf("Failed to add task: %v", err)
	}

	if len(s.EntryIDs) != 1 {
		t.Errorf("Expected 1 task in scheduler, got %d", len(s.EntryIDs))
	}

	s.RemoveTask(1)
	if len(s.EntryIDs) != 0 {
		t.Errorf("Expected 0 tasks after removal, got %d", len(s.EntryIDs))
	}
}
