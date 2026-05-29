import { test, expect } from "@playwright/test";

test.use({ storageState: "playwright/.auth/admin.json" });

test.describe("Admin Dashboard", () => {
  const testEmail = `e2e-test-${Date.now()}@example.com`;

  test("should allow admin to manage users", async ({ page }) => {
    // Navigate directly to admin dashboard (already authenticated as admin)
    await page.goto("/adminDashboard");
    await expect(page).toHaveURL(/.*adminDashboard/);

    // Verify page elements
    await expect(page.getByText("Staff Management")).toBeVisible();
    await expect(page.getByPlaceholder("Search by name or email...")).toBeVisible();
    await expect(page.getByRole("button", { name: /HIRE NEW STAFF/i })).toBeVisible();

    // Open create user modal
    await page.getByRole("button", { name: /HIRE NEW STAFF/i }).click();
    await expect(page.getByText("Create New Staff")).toBeVisible();

    // Fill the form
    await page.getByPlaceholder("First Name").fill("E2E");
    await page.getByPlaceholder("Last Name").fill("TestUser");
    await page.getByTestId("create-user-email").fill(testEmail);
    await page.getByPlaceholder("Password", { exact: true }).fill("TestPass123!");
    await page.getByPlaceholder("Confirm Password", { exact: true }).fill("TestPass123!");
    // Select role "User"
    await page.locator('select').selectOption("user");

    // Create user
    await page.getByRole("button", { name: "Create User" }).click();

    // Wait for modal to close and user to appear in list
    await expect(page.getByText("Create New Staff")).not.toBeVisible();

    // Search for the created user
    await page.getByPlaceholder("Search by name or email...").fill("E2E TestUser");
    await expect(page.getByText("E2E TestUser")).toBeVisible();
    await expect(page.getByText(testEmail)).toBeVisible();

    // Clear search and find user again
    await page.getByPlaceholder("Search by name or email...").fill("");
    await page.getByPlaceholder("Search by name or email...").fill(testEmail);
    await expect(page.getByText(testEmail)).toBeVisible();

    // Clear search
    await page.getByPlaceholder("Search by name or email...").fill("");

    // Delete the test user
    await page.getByTestId(`delete-user-${testEmail}`).click();

    // Confirm delete modal
    await expect(page.getByText("Remove Staff")).toBeVisible();
    await page.getByRole("button", { name: "Delete" }).click();

    // Verify user is gone
    await page.getByPlaceholder("Search by name or email...").fill(testEmail);
    await expect(page.getByText("No users found")).toBeVisible();
  });
});
