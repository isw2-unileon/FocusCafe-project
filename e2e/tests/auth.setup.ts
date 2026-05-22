import { test as setup, expect } from "@playwright/test";

const authDir = "playwright/.auth";

setup("authenticate as user", async ({ page }) => {
  await page.goto("/login");
  await page.getByPlaceholder("Email").fill("user@user.com");
  await page.getByPlaceholder("Password", { exact: true }).fill("user123");
  await page.getByRole("button", { name: "Enter the Café" }).click();
  await page.waitForURL(/.*home/);

  // Wait for token to be in localStorage
  await page.waitForFunction(() => localStorage.getItem("token") !== null);

  // Save storage state
  await page.context().storageState({ path: `${authDir}/user.json` });
});

setup("authenticate as admin", async ({ page }) => {
  await page.goto("/login");
  await page.getByPlaceholder("Email").fill("admin@admin.com");
  await page.getByPlaceholder("Password", { exact: true }).fill("admin123");
  await page.getByRole("button", { name: "Enter the Café" }).click();
  await page.waitForURL(/.*adminDashboard/);

  // Wait for token to be in localStorage
  await page.waitForFunction(() => localStorage.getItem("token") !== null);

  // Save storage state
  await page.context().storageState({ path: `${authDir}/admin.json` });
});
