import { test, expect } from "@playwright/test";

test.use({ storageState: "playwright/.auth/user.json" });

test.describe("Protected Routes — Authenticated User", () => {
  test("should redirect normal user away from admin dashboard", async ({ page }) => {
    // Already authenticated as normal user via storageState
    await page.goto("/adminDashboard");

    // Should be redirected to /home
    await expect(page).toHaveURL(/.*home/);
  });
});
