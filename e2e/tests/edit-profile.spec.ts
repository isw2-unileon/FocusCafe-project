import { test, expect } from "@playwright/test";
import { loginAsUser } from "../lib/auth-helper";

test.describe("Edit Profile", () => {
  test("should update first and last name and reflect changes on Dashboard", async ({ page }) => {
    // 1. Login as user
    await loginAsUser(page);

    // 2. Navigate to Edit Profile
    await page.goto("/edit-profile");
    await expect(page).toHaveURL(/.*edit-profile/);

    // 3. Capture original names for cleanup
    const firstNameInput = page.getByPlaceholder("Your first name");
    const lastNameInput = page.getByPlaceholder("Your last name");
    const originalFirst = await firstNameInput.inputValue();
    const originalLast = await lastNameInput.inputValue();

    try {
      // 4. Change names
      await firstNameInput.fill("E2EChanged");
      await lastNameInput.fill("NameChanged");

      // 5. Save changes
      await page.getByRole("button", { name: "Save Changes" }).click();

      // 6. Verify redirect to /dashboard
      await page.waitForURL(/.*dashboard/);
      await expect(page).toHaveURL(/.*dashboard/);

      // 7. Verify updated name appears on Dashboard
      await expect(page.getByText("E2EChanged NameChanged")).toBeVisible();
    } finally {
      // 8. ALWAYS revert names back to original (even if test fails mid-way)
      await page.goto("/edit-profile");
      await expect(page).toHaveURL(/.*edit-profile/);
      await page.getByPlaceholder("Your first name").fill(originalFirst);
      await page.getByPlaceholder("Your last name").fill(originalLast);
      await page.getByRole("button", { name: "Save Changes" }).click();
      await page.waitForURL(/.*dashboard/);
    }

    // 9. Verify original name is restored
    await expect(page.getByText("Test User")).toBeVisible();
  });
});
