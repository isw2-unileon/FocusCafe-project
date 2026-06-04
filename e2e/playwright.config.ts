import { defineConfig, devices } from "@playwright/test";
import * as dotenv from "dotenv";
import * as path from "path";

dotenv.config({ path: path.resolve(__dirname, "../.env") });


export default defineConfig({
  testDir: "./tests",
  timeout: 30_000,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 4 : 1,
  reporter: "html",
  globalTeardown: require.resolve("./global-teardown"),

  use: {
    baseURL: "http://localhost:5173",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
  },

  webServer: [
    {
      command: process.platform === "win32"
        ? "cd .. && go build -o backend/bin/server.exe ./backend/cmd/server && .\\backend\\bin\\server.exe"
        : "cd .. && make build-backend && ./backend/bin/server",
      port: Number(process.env.PORT) || 8080,
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,

      env: {
        DATABASE_URL: process.env.DATABASE_URL   || "postgresql://postgres:postgres@127.0.0.1:54322/postgres?sslmode=disable",
        SUPABASE_URL: process.env.SUPABASE_URL || "http://127.0.0.1:54321",
        SUPABASE_ANON_KEY: process.env.SUPABASE_ANON_KEY || process.env.SUPABASE_KEY || "sb_secret_local_development_key_mock_value_for_testing_purposes",
        SUPABASE_SERVICE_ROLE_KEY: process.env.SUPABASE_SERVICE_ROLE_KEY || "sb_publishable_local_development_key_mock_value_for_testing_purposes",
        PORT: process.env.PORT || "8080",
        GIN_MODE: "debug",
      }
    },
    {
      command: process.env.CI 
        ? "cd ../frontend && npx vite preview --port 5173" 
        : "cd ../frontend && npm run dev",
      port: 5173,
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
      stdout: "pipe",
    },
  ],

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