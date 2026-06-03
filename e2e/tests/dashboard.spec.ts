import { test, expect } from "@playwright/test";
import { RegisterAndloginAsUser } from "../lib/auth-helper";

test.describe("Dashboard view", () => {
  test("should display user profile data after navigating from Home", async ({ page }) => {
    // 1. Login
    await RegisterAndloginAsUser(page);

    // 2. Should be on Home after login
    await expect(page).toHaveURL(/.*home/, { timeout: 10000 });

    // 3. Navigate to Dashboard by clicking the avatar icon in header
    await page.getByTestId("nav-dashboard").click();

    // 4. Verify we're on Dashboard
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

    // 5. Verify profile data is visible
    await expect(page.getByTestId("profile-full-name")).toBeVisible({ timeout: 10000 });
    await expect(page.getByTestId("profile-username")).toBeVisible({ timeout: 10000 });
    await expect(page.getByTestId("profile-level")).toBeVisible({ timeout: 10000 });
    await expect(page.getByTestId("profile-rank")).toBeVisible({ timeout: 10000 });
    await expect(page.getByTestId("stat-energy-value")).toBeVisible({ timeout: 10000 });
    await expect(page.getByTestId("stat-experience-value")).toBeVisible({ timeout: 10000 });
    await expect(page.getByTestId("profile-email")).toBeVisible({ timeout: 10000 });
    await expect(page.getByTestId("profile-member-since")).toBeVisible({ timeout: 10000 });
  });

  test("should navigate back to Home from Dashboard", async ({ page }) => {
    // 1. Login
    await RegisterAndloginAsUser(page);

    // 2. Go to Dashboard
    await page.getByTestId("nav-dashboard").click();
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

    // 3. Click back button (ArrowLeft)
    await page.locator("button").filter({ has: page.locator("svg") }).first().click();

    // 4. Should be back on Home
    await expect(page).toHaveURL(/.*home/, { timeout: 10000 });
    await expect(page.getByRole('heading', { name: 'Player Stats' })).toBeVisible({ timeout: 10000 });
  });
});
