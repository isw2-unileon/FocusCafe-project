import { test, expect } from "@playwright/test";

test.use({ storageState: "playwright/.auth/user.json" });

test.describe("Edit Profile", () => {
  test("should update first and last name and reflect changes on Dashboard", async ({ page }) => {
    // Navigate to Edit Profile (already authenticated)
    await page.goto("/edit-profile");
    await expect(page).toHaveURL(/.*edit-profile/);

    // Capture original names for cleanup
    const firstNameInput = page.getByPlaceholder("Your first name");
    const lastNameInput = page.getByPlaceholder("Your last name");
    const originalFirst = await firstNameInput.inputValue();
    const originalLast = await lastNameInput.inputValue();

    try {
      // 1. Change names
      await firstNameInput.fill("E2EChanged");
      await lastNameInput.fill("NameChanged");

      // 2. Save changes
      await page.getByRole("button", { name: "Save Changes" }).click();

      // 3. Verify redirect to /dashboard
      await page.waitForURL(/.*dashboard/);
      await expect(page).toHaveURL(/.*dashboard/);

      // 4. Verify updated name appears on Dashboard
      await expect(page.getByText("E2EChanged NameChanged")).toBeVisible();
    } finally {
      // 5. ALWAYS revert names back to original (even if test fails mid-way)
      await page.goto("/edit-profile");
      await page.getByPlaceholder("Your first name").fill(originalFirst);
      await page.getByPlaceholder("Your last name").fill(originalLast);
      await page.getByRole("button", { name: "Save Changes" }).click();
      await page.waitForURL(/.*dashboard/);
    }

    // 6. Verify original name is restored
    await expect(page.getByText("Test User")).toBeVisible();
  });
});
