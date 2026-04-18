# Optimize Cron Input Validation Plan

**Goal:** Improve the user experience and system reliability by validating custom Cron expressions both on the frontend (UI hints/validation) and backend (API rejection).

**Changes:**

### 1. Backend (`internal/core/scheduler/scheduler.go`)
- Implement a `ValidateCron(cronExpr string) error` function. Since we use `cron.WithSeconds()`, the easiest reliable way to validate is by temporarily instantiating `cron.New(cron.WithSeconds())` and calling `AddFunc` with an empty function to capture any parsing errors.

### 2. Backend API (`internal/api/router.go`)
- In `createTask` and `updateTask`: Before saving to the database, check if `task.ScheduleMode == "custom"`. If so, validate `task.Cron` using `scheduler.ValidateCron`. Return a `400 Bad Request` with a user-friendly error message if it's invalid.
- In `updateScheduleSettings`: Validate the `input.Cron` using `scheduler.ValidateCron` before saving it to the database or updating the global schedule.

### 3. Frontend (`web/src/views/Tasks.vue`)
- Ensure the `cron` input fields (both in the global settings card and the task dialog) have clear placeholders.
- Leverage the backend's validation errors (which will now cleanly return 400 with the exact parsing failure) by relying on the existing axios interceptor. To improve UX further, we can add a small tip underneath the input box explaining the 6-field format: `秒 分 时 日 月 周`.

This plan ensures that no invalid cron expressions can ever enter the database or crash the scheduler logic silently.
