package scheduler

import (
	"testing"
)

func TestScheduler_AddAndRemoveTask(t *testing.T) {
	s := New(nil)
	s.Start()
	defer s.Stop()

	s.UpdateTask(1, "custom", "0 * * * * *")

	if len(s.CustomEntryIDs) != 1 {
		t.Errorf("Expected 1 task in scheduler, got %d", len(s.CustomEntryIDs))
	}

	s.RemoveTask(1)
	if len(s.CustomEntryIDs) != 0 {
		t.Errorf("Expected 0 tasks after removal, got %d", len(s.CustomEntryIDs))
	}
}
