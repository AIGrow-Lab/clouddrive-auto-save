# Task Micro-Progress UI Revamp Plan (Database-Backed)

**Goal:** Transform the "Task Micro-Progress" card into a real, persistent tracker by storing progress in the database and syncing via polling + SSE.

### 1. Database Model (`internal/db/db.go`)
- [x] Add `Percent` (int) and `Stage` (string) to `Task` struct.

### 2. Worker Engine (`internal/core/worker/worker.go`)
- [x] Implement `updateProgress` helper to persist `Percent`, `Stage`, and `Message`.
- [x] Refactor `execute` and `finishTask` to use the helper.

### 3. Backend API (`internal/api/router.go`)
- [x] Update `getDashboardStats` to return tasks that are `running`, `failed`, or `success` within the last 15 seconds.

### 4. Frontend UI (`web/src/views/Dashboard.vue`)
- [x] Remove fragile `isReplay` and `setTimeout` logic.
- [x] Implement 5-second polling in `onMounted`.
- [x] Smart-merge API state with local reactive state in `fetchStats`.
- [x] Update `handleProgressMessage` for real-time sub-second updates.

