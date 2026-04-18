# Fix Tasks UI Reactivity Plan

**Goal:** Resolve the issue where task statuses in the `Tasks.vue` table do not update automatically via SSE events, forcing the user to manually refresh the page.

**Context:** 
The `Tasks.vue` component listens to the `/api/dashboard/logs` SSE stream and attempts to merge incoming `[EVENT:{"type":"task_update",...}]` payloads into the local `taskList.value` array using `Object.assign`. However, if the incoming `task` object from the backend does not have the `Account` relation preloaded, `Object.assign` overwrites the existing `row.account` with a zeroed-out object, potentially breaking the UI or failing to trigger Vue's deep reactivity correctly for the specific `status` and `message` cells.

**Changes (`web/src/views/Tasks.vue`):**
1. Instead of using a blunt `Object.assign(taskList.value[idx], ev.task)`, selectively update only the progress-related fields that change during execution:
   - `status`
   - `message`
   - `last_run`
   - `percent`
   - `stage`
2. This surgical assignment guarantees that Vue 3's proxy intercepts the property mutations and re-renders the specific table cells without destroying nested relational data like `account`.
3. Add a fallback `fetchList(true)` when a task reaches `Success` or `Failed` to ensure the table reaches a fully consistent state upon task completion.
