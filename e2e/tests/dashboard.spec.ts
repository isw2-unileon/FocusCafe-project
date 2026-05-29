import { test, expect } from "@playwright/test";
import { loginAsUser } from "../lib/auth-helper";

test.describe("Dashboard view", () => {
  test("should display user profile data after login", async ({ page }) => {
    // 1. Login as user
    await loginAsUser(page);

    // 2. Navigate to dashboard
    await page.goto("/dashboard");
    await expect(page).toHaveURL(/.*dashboard/);

    // 3. Verify profile data is visible
    await expect(page.getByText("Test User")).toBeVisible();
    // Username (strict mode fix: use .first() since @testuser appears twice)
    await expect(page.getByText("@testuser").first()).toBeVisible();
    // Level badge
    await expect(page.getByText(/Level\s+5/)).toBeVisible();
    // Rank badge (level 5 = Focus Apprentice)
    await expect(page.getByText("Focus Apprentice")).toBeVisible();
    // Energy
    await expect(page.getByText(/Energy/)).toBeVisible();
    await expect(page.getByText(/300/)).toBeVisible();
    // Experience
    await expect(page.getByText(/Experience/)).toBeVisible();
    await expect(page.getByText(/2500/)).toBeVisible();
    // Email
    await expect(page.getByText("user@user.com")).toBeVisible();
    // Member Since
    await expect(page.getByText(/Member Since/)).toBeVisible();
  });
});
