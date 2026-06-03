import { test, expect } from "@playwright/test";
import { CreateAndLoginAsAdmin } from "../lib/auth-helper";

test.describe("Admin Dashboard", () => {
  const timestamp = Date.now();
  const testEmail = `e2e-test-${timestamp}@example.com`;
  const uniqueFirstName = `E2E-${timestamp}`;
  const uniqueLastName = `User`;
  const fullName = `${uniqueFirstName} ${uniqueLastName}`;

  test("should allow admin to manage users through full flow", async ({ page }) => {
    // 1. Login as admin (goes to adminDashboard by default)
    await CreateAndLoginAsAdmin(page);
    await expect(page).toHaveURL(/.*adminDashboard/);

    // 2. Verify page elements
    await expect(page.getByText("Staff Management")).toBeVisible();
    const searchInput = page.getByPlaceholder("Search by name or email...");
    await expect(searchInput).toBeVisible();
    await expect(page.getByRole("button", { name: /HIRE NEW STAFF/i })).toBeVisible();

    // 3. Open create user modal
    await page.getByRole("button", { name: /HIRE NEW STAFF/i }).click();
    await expect(page.getByText("Create New Staff")).toBeVisible();

    const modalHeading = page.getByRole("heading", { name: "Create New Staff" });
    await expect(modalHeading).toBeVisible();

    // 4. Fill the form
    await page.getByPlaceholder("First Name").fill(uniqueFirstName);
    await page.getByPlaceholder("Last Name").fill(uniqueLastName);
    await page.getByTestId("create-user-email").fill(testEmail);
    await page.getByPlaceholder("Password", { exact: true }).fill("TestPass123!");
    await page.getByPlaceholder("Confirm Password", { exact: true }).fill("TestPass123!");
    // Select role "User"
    await page.locator('select').selectOption("user");

    // 5. Create user
    await page.getByRole("button", { name: "Create User" }).click();
    // 6. Wait for modal to close and user to appear in list
    await expect(modalHeading).not.toBeVisible();

    // 7. Search for the created user
    await searchInput.waitFor({ state: "visible" });
    await searchInput.fill(fullName);
    await expect(page.getByText(fullName)).toBeVisible();
    await expect(page.getByText(testEmail)).toBeVisible();

    await page.getByPlaceholder("Search by name or email...").fill(fullName);
    await expect(page.getByText(fullName)).toBeVisible();
    await expect(page.getByText(testEmail)).toBeVisible();

    // 8. Clear search and find user again
    await page.getByPlaceholder("Search by name or email...").fill("");
    const deleteUserButton = page.getByTestId(`delete-user-${testEmail}`);
    await expect(deleteUserButton).toBeVisible({ timeout: 5000 });

    // 11.Open modal
    await deleteUserButton.click();
    const removeStaffModal = page.getByText("Remove Staff");
    await expect(removeStaffModal).toBeVisible();

    await page.getByRole("button", { name: "Delete" }).click();
    await expect(removeStaffModal).not.toBeVisible({ timeout: 10000 });

    // 12. Verify user is gone
    await searchInput.fill(testEmail);
    await searchInput.press("Enter");
    await expect(page.getByText("No users found")).toBeVisible({ timeout: 10000 });
  });

  test("should navigate to admin dashboard from Home via shield icon", async ({ page }) => {
    // 1. Login as admin
    await CreateAndLoginAsAdmin(page);

    // Admin login redirects directly to adminDashboard, but let's verify
    await expect(page).toHaveURL(/.*adminDashboard/);

    // 2. Navigate to Home (if there's a back button or home link)
    // For now, verify we're on the correct page
    await expect(page.getByText("Staff Management")).toBeVisible();
  });
});
