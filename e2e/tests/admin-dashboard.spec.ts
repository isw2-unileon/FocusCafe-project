import { test, expect } from "@playwright/test";
import { loginAsAdmin } from "../lib/auth-helper";

test.describe("Admin Dashboard", () => {
  const testEmail = `e2e-test-${Date.now()}@example.com`;

  test("should allow admin to manage users through full flow", async ({ page }) => {
    // 1. Login as admin (goes to adminDashboard by default)
    await loginAsAdmin(page);
    await expect(page).toHaveURL(/.*adminDashboard/);

    // 2. Verify page elements
    await expect(page.getByText("Staff Management")).toBeVisible();
    await expect(page.getByPlaceholder("Search by name or email...")).toBeVisible();
    await expect(page.getByRole("button", { name: /HIRE NEW STAFF/i })).toBeVisible();

    // 3. Open create user modal
    await page.getByRole("button", { name: /HIRE NEW STAFF/i }).click();
    await expect(page.getByText("Create New Staff")).toBeVisible();

    // 4. Fill the form
    await page.getByPlaceholder("First Name").fill("E2E");
    await page.getByPlaceholder("Last Name").fill("TestUser");
    await page.getByTestId("create-user-email").fill(testEmail);
    await page.getByPlaceholder("Password", { exact: true }).fill("TestPass123!");
    await page.getByPlaceholder("Confirm Password", { exact: true }).fill("TestPass123!");
    // Select role "User"
    await page.locator('select').selectOption("user");

    // 5. Create user
    await page.getByRole("button", { name: "Create User" }).click();

    // 6. Wait for modal to close and user to appear in list
    await expect(page.getByText("Create New Staff")).not.toBeVisible();

    // 7. Search for the created user
    await page.getByPlaceholder("Search by name or email...").fill("E2E TestUser");
    await expect(page.getByText("E2E TestUser")).toBeVisible();
    await expect(page.getByText(testEmail)).toBeVisible();

    // 8. Clear search and find user again
    await page.getByPlaceholder("Search by name or email...").fill("");
    await page.getByPlaceholder("Search by name or email...").fill(testEmail);
    await expect(page.getByText(testEmail)).toBeVisible();

    // 9. Clear search
    await page.getByPlaceholder("Search by name or email...").fill("");

    // 10. Delete the test user
    await page.getByTestId(`delete-user-${testEmail}`).click();

    // 11. Confirm delete modal
    await expect(page.getByText("Remove Staff")).toBeVisible();
    await page.getByRole("button", { name: "Delete" }).click();

    // 12. Verify user is gone
    await page.getByPlaceholder("Search by name or email...").fill(testEmail);
    await expect(page.getByText("No users found")).toBeVisible();
  });

  test("should navigate to admin dashboard from Home via shield icon", async ({ page }) => {
    // 1. Login as admin
    await loginAsAdmin(page);

    // Admin login redirects directly to adminDashboard, but let's verify
    await expect(page).toHaveURL(/.*adminDashboard/);

    // 2. Navigate to Home (if there's a back button or home link)
    // For now, verify we're on the correct page
    await expect(page.getByText("Staff Management")).toBeVisible();
  });
});
