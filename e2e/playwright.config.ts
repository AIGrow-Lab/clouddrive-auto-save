import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:8080',
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  // 自动启动后端服务（可选，目前假设后端手动启动或在流水线中已启动）
  /* webServer: {
    command: 'E2E_TEST_MODE=true ./bin/ucas',
    url: 'http://localhost:8080',
    reuseExistingServer: !process.env.CI,
  }, */
});
