# Fix Stuck Running Task Plan

**Goal:** Ensure that tasks do not get permanently stuck in the `"running"` state if the backend server is restarted abruptly or crashes.

**Target File:** `cmd/server/main.go`

**Changes:**
1. In the backend initialization phase (right after `db.InitDB`), execute a database update query.
2. Find all tasks with `status = "running"`.
3. Reset their status to `"pending"` (or `"failed"`) and set a message indicating that the task was interrupted by a server restart.
   - Example query: `db.DB.Model(&db.Task{}).Where("status = ?", "running").Updates(map[string]interface{}{"status": "pending", "message": "Server restarted during execution"})`
4. This ensures that when the cron scheduler or the worker manager starts, no phantom running tasks block future executions.
