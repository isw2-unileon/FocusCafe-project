import { test, expect } from "@playwright/test";

test.use({ storageState: "playwright/.auth/user.json" });

test.describe("Dashboard view", () => {
  test("should display user profile data after login", async ({ page }) => {
    // Navigate directly to dashboard (already authenticated via storageState)
    await page.goto("/dashboard");
    await expect(page).toHaveURL(/.*dashboard/);

    // Verify profile data is visible
    // Name
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
