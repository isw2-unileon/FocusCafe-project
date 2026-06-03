import { test, expect } from "@playwright/test";
import { RegisterAndloginAsUser } from "../lib/auth-helper";

test.describe("Protected Routes — Authenticated User", () => {
  test("should not show admin navigation for normal users", async ({ page }) => {
    // 1. Login as normal user
    await RegisterAndloginAsUser(page);
    await expect(page).toHaveURL(/.*home/);

    // 2. Verify admin shield icon is NOT visible in header
    const adminShield = page.locator('a[href="/adminDashboard"]');
    await expect(adminShield).not.toBeVisible();

    // 3. Try to navigate to admin dashboard directly
    await page.goto("/adminDashboard");

    // 4. Should be redirected to /home
    await expect(page).toHaveURL(/.*home/);
  });
});
