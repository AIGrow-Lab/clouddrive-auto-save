# Database-Backed Micro-Progress Plan

**Goal:** Resolve all progress bar race conditions, missing tasks, and reappearance bugs by making the database the single source of truth for task progress, as suggested by the user.

**Architecture Shift:**
Currently, progress (`Percent`, `Stage`) only exists in transient SSE messages. If a page refresh occurs, the frontend tries to reconstruct this state by replaying logs, which is brittle and causes timing conflicts. We will move `Percent` and `Stage` into the `Task` database model. 

**Changes:**

### 1. Database Model (`internal/db/db.go`)
- Add `Percent int` and `Stage string` to the `Task` struct so progress state is persistent.

### 2. Worker Engine (`internal/core/worker/worker.go`)
- Create a helper `updateProgress(task, percent, stage, message)` that updates the `Task` in the DB and emits the SSE log.
- Replace all standalone `[PROGRESS...]` logs in `execute` with calls to `updateProgress`.
- Update `finishTask` to also set `Percent = 100` and `Stage = "Success" | "Failed"` in the database.

### 3. Backend API (`internal/api/router.go`)
- In `getDashboardStats`, change the `runningTasksList` query to fetch tasks that are:
  - `status = "running"`
  - OR `status = "failed"` (so failures stay visible for intervention)
  - OR `status = "success" AND last_run >= time.Now().Add(-10 * time.Second)` (so successes naturally expire from the API response 10 seconds after completion).

### 4. Frontend UI (`web/src/views/Dashboard.vue`)
- **Remove** the fragile `setTimeout` / `dismissTask` and `isReplay` logic.
- **Remove** the historical log replay for progress messages.
- Add a 5-second `setInterval` to poll `fetchStats()`.
- When `fetchStats()` returns `running_tasks_list`, smartly merge it with `runningTasks.value` (update existing, add new, remove stale) to prevent UI flickering.
- `handleProgressMessage` will now only be responsible for real-time sub-second visual updates between the 5-second API polls.

This fundamentally simplifies the frontend state management and guarantees 100% accuracy.
