import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  timeout: 30_000,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: "html",

  use: {
    baseURL: "http://localhost:5173",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },

  // NOTE: Start backend and frontend manually before running tests:
  //   Terminal 1: make run-backend    (port 8080)
  //   Terminal 2: cd frontend && npm run dev  (port 5173)
  // webServer: [
  //   {
  //     command: "go run ./backend/cmd/server",
  //     cwd: "..",
  //     port: 8080,
  //     reuseExistingServer: !process.env.CI,
  //   },
  //   {
  //     command: "cd frontend && npm run dev",
  //     cwd: "..",
  //     port: 5173,
  //     reuseExistingServer: !process.env.CI,
  //   },
  // ],

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    {
      name: "firefox",
      use: { ...devices["Desktop Firefox"] },
    },
    {
      name: "webkit",
      use: { ...devices["Desktop Safari"] },
    },
  ],
});
