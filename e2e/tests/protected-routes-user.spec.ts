import { test, expect } from "@playwright/test";
import { loginAsUser } from "../lib/auth-helper";

test.describe("Protected Routes — Authenticated User", () => {
  test("should redirect normal user away from admin dashboard", async ({ page }) => {
    // 1. Login as normal user
    await loginAsUser(page);

    // 2. Try to access admin dashboard
    await page.goto("/adminDashboard");

    // 3. Should be redirected to /home
    await expect(page).toHaveURL(/.*home/);
  });
});
