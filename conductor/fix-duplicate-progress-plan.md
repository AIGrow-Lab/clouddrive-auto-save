# Fix Duplicate Progress UI Plan

**Goal:** Fix the issue where clicking "Retry" causes a duplicate progress card to appear (one with the task's real name and one with a "Task #ID" placeholder).

**Context:** 
1. When a task is triggered, the SSE progress log (`handleProgressMessage`) creates a new entry in `runningTasks.value` with `id` as a string (`"1"`) and a placeholder name (`"任务 #1"`).
2. It then calls `fetchStats()`, which gets the `running_tasks_list` from the backend API. The API returns the `id` as a number (`1`) and its real name.
3. Inside `fetchStats()`, the code checks `runningTasks.value.find(t => t.id === task.id)`. Because it uses strict equality (`===`) and comparing a string (`"1"`) to a number (`1`), it returns `false`.
4. As a result, `fetchStats` pushes a *second* entry into the array, causing the duplicate UI.

**Changes (`web/src/views/Dashboard.vue`):**
- Update the `find` condition in `fetchStats` to use string conversion: `String(t.id) === String(task.id)`.
- If the task *does* exist, check if its name starts with `"任务 #"`. If it does, update the placeholder name with the real task name (`task.name`) returned from the backend.
