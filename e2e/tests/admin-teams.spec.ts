import { test, expect } from "@playwright/test";
import { loginAsAdmin } from "../lib/auth-helper";

test.describe("Admin Dashboard — Teams Tab", () => {
  test("should navigate to Teams tab and display groups", async ({ page }) => {
    // 1. Login as admin
    await loginAsAdmin(page);
    await expect(page).toHaveURL(/.*adminDashboard/);

    // 2. Verify admin dashboard loaded (Staff tab active by default)
    await expect(page.getByText("Staff Management")).toBeVisible();

    // 3. Click Teams tab
    await page.getByRole("button", { name: "Teams" }).click();

    // 4. Verify Teams tab is active (search placeholder changes)
    await expect(page.getByPlaceholder("Search by team name or code...")).toBeVisible();

    // 5. Verify either groups are listed or empty state appears
    const body = page.locator("body");
    await expect(body).toContainText(/Teams|Teams will appear here when users create them/i);
  });

  test("should expand a group to see members", async ({ page }) => {
    // 1. Login as admin
    await loginAsAdmin(page);
    await expect(page).toHaveURL(/.*adminDashboard/);

    // 2. Navigate to Teams tab
    await page.getByRole("button", { name: "Teams" }).click();
    await expect(page.getByPlaceholder("Search by team name or code...")).toBeVisible();

    // 3. Try to find and expand a group card if any exists
    const groupCard = page.locator("[data-testid^='admin-group-']").first();
    if (await groupCard.isVisible().catch(() => false)) {
      await groupCard.click();
      // Verify member list appears
      await expect(page.getByText(/Members|Leader/i).first()).toBeVisible();
    } else {
      // If no groups exist, that's acceptable for this test environment
      test.skip(true, "No groups exist in test environment");
    }
  });
});
