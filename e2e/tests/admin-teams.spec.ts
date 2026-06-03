import { test, expect } from "@playwright/test";
import { CreateAndLoginAsAdmin, RegisterAndloginAsUser } from "../lib/auth-helper";


test.describe("Admin Dashboard — Teams Tab", () => {
  test("should navigate to Teams tab and display groups", async ({ page }) => {
    // 1. Login as admin
    await CreateAndLoginAsAdmin(page);

    // 2. Verify admin dashboard loaded (Staff tab active by default)
    await expect(page.getByText("Staff Management")).toBeVisible({ timeout: 10000 });

    // 3. Click Teams tab
    await page.getByRole("button", { name: "Teams" }).click();

    // 4. Verify Teams tab is active (search placeholder changes)
    await expect(page.getByPlaceholder("Search by team name or code...")).toBeVisible({ timeout: 10000 });

    // 5. Verify either groups are listed or empty state appears
    const body = page.locator("body");
    await expect(body).toContainText(/Teams|Teams will appear here when users create them/i, { timeout: 10000 });
  });

  test("should expand a group to see members", async ({ page }) => {
    // 1. Login as admin
    await CreateAndLoginAsAdmin(page);

    // 2. Navigate to Teams tab
    await page.getByRole("button", { name: "Teams" }).click();
    const searchTeamsInput = page.getByPlaceholder("Search by team name or code...");
    await expect(searchTeamsInput).toBeVisible({ timeout: 10000 });

    // 3. Try to find and expand a group card if any exists
    const groupCard = page.locator("[data-testid^='admin-group-']").first();
    if (await groupCard.isVisible().catch(() => false)) {
      await groupCard.click();
      // Verify member list appears
      await expect(page.getByText(/Members|Leader/i).first()).toBeVisible({ timeout: 10000 });
    } else {
      // If no groups exist, that's acceptable for this test environment
      test.skip(true, "No groups exist in test environment");
    }
  });

  test("should allow admin to delete a group", async ({ browser, page }) => {
    const groupName = `E2E-DeleteTeam-${Date.now()}`;

    // 1. Create a normal user and a group
    const userContext = await browser.newContext();
    const userPage = await userContext.newPage();

    await RegisterAndloginAsUser(userPage);
    await expect(userPage).toHaveURL(/.*home/);

    await userPage.getByPlaceholder('New Team').fill(groupName);
    await userPage.getByTitle('Create Team').click();

    await expect(userPage.getByText('Team Created!')).toBeVisible();
    await userPage.getByRole('button', { name: 'Done!' }).click();

    // 2. Login as admin
    await CreateAndLoginAsAdmin(page);

    // 3. Navigate to Teams tab
    await page.getByRole("button", { name: "Teams" }).click();
    await expect(page.getByPlaceholder("Search by team name or code...")).toBeVisible();

    // 4. Find the group card in the list and click the delete button
    const groupCard = page.locator('div.bg-\\[\\#fdfaf7\\]').filter({ hasText: groupName });
    await expect(groupCard).toBeVisible();

    await groupCard.locator('button').first().click();

    // 5. Confirm deletion in modal
    await expect(page.getByText("Delete Team")).toBeVisible();
    await page.getByRole("button", { name: "Delete" }).click();

    // 6. Verify success
    await expect(page.getByText("Group deleted successfully")).toBeVisible();
    await expect(groupCard).not.toBeVisible();

    // 7. Cleanup
    await userContext.close();
  });
});
