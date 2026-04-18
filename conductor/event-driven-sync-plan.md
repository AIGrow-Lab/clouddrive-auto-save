# Backend-Driven Event Sync Plan

**Goal:** Eliminate all frontend polling (`setInterval`) and achieve "ultimate real-time synchronization" by utilizing Server-Sent Events (SSE) to push database and state changes directly to the frontend.

**Architecture Shift:**
Currently, `Tasks.vue` and `Dashboard.vue` rely on a 5-second `setInterval` loop to fetch updates from the backend to ensure consistency. We will replace this with a backend-driven event system leveraging our existing SSE log stream.

### 1. Backend Event System (`internal/utils/events.go`)
- Create an event broadcasting helper that serializes structured JSON events (e.g., `task_update`, `task_delete`, `stats_update`) and pushes them through the existing `GlobalBroadcaster` with a special prefix `[EVENT:{...}]`.

### 2. Backend Event Triggers
- **`worker.go`**: Emit `task_update` whenever progress changes. Emit `stats_update` when a task finishes.
- **`router.go`**: Emit `task_update` and `stats_update` on task creation, modification, deletion, and manual execution triggers.

### 3. Frontend: `Tasks.vue`
- Remove the 5-second `setInterval`.
- Initialize an `EventSource` connection to `/api/dashboard/logs`.
- Parse incoming `[EVENT:...]` messages:
  - On `task_update`: Find the task in `taskList.value` and do an in-place `Object.assign` to update status, last run time, etc., with zero latency.
  - On `task_delete`: Remove the task from the list.

### 4. Frontend: `Dashboard.vue`
- Remove the 5-second `setInterval`.
- Update the existing SSE message handler to parse `[EVENT:...]` messages.
- On `stats_update`: Trigger `fetchStats(false)` exactly once to refresh the dashboard numbers and running list, guaranteeing 100% accuracy without blind polling.
- Ensure that `[EVENT:...]` messages are skipped from the terminal log display.

This transformation ensures that the UI reacts instantaneously (within milliseconds) to any backend changes, reducing network traffic and providing a perfectly synced user experience.
