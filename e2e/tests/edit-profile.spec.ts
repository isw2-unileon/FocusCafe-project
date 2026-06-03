import { test, expect } from "@playwright/test";
import { RegisterAndloginAsUser } from "../lib/auth-helper";

test.describe("Edit Profile", () => {
  test("should update first and last name through full navigation flow", async ({ page }) => {
    // 1. Login
    await RegisterAndloginAsUser(page);
    await expect(page).toHaveURL(/.*home/, { timeout: 10000 });

    // 2. Navigate to Dashboard via avatar icon in header
    await page.getByTestId("nav-dashboard").click();
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

    // 3. Click "Edit Profile" button
    await page.getByRole("button", { name: "Edit Profile" }).click();
    await expect(page).toHaveURL(/.*edit-profile/, { timeout: 10000 });

    // 4. Capture original names for cleanup
    const firstNameInput = page.getByPlaceholder("Your first name");
    const lastNameInput = page.getByPlaceholder("Your last name");
    await expect(firstNameInput).toBeVisible({ timeout: 10000 });
    const originalFirst = await firstNameInput.inputValue();
    const originalLast = await lastNameInput.inputValue();

    try {
      // 5. Change names
      await firstNameInput.fill("E2EChanged");
      await lastNameInput.fill("NameChanged");

      // 6. Save changes
      await page.getByRole("button", { name: "Save Changes" }).click();

      // 7. Verify redirect to /dashboard
      await page.waitForURL(/.*dashboard/, { timeout: 10000 });
      await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

      // 8. Verify updated name appears on Dashboard
      await expect(page.getByTestId("profile-full-name")).toHaveText("E2EChanged NameChanged", { timeout: 10000 });
    } finally {
      // 9. ALWAYS revert names back to original (even if test fails mid-way)
      // Navigate back to Edit Profile
      await page.getByRole("button", { name: "Edit Profile" }).click();
      await expect(page).toHaveURL(/.*edit-profile/, { timeout: 10000 });
      await page.getByPlaceholder("Your first name").fill(originalFirst);
      await page.getByPlaceholder("Your last name").fill(originalLast);
      await page.getByRole("button", { name: "Save Changes" }).click();
      await page.waitForURL(/.*dashboard/, { timeout: 10000 });
    }

    // 10. Verify original name is restored
    await expect(page.getByTestId("profile-full-name")).toHaveText(`${originalFirst} ${originalLast}`, { timeout: 10000 });
  });

  test("should navigate back from Edit Profile to Dashboard", async ({ page }) => {
    // 1. Login
    await RegisterAndloginAsUser(page);

    // 2. Go to Dashboard
    await page.getByTestId("nav-dashboard").click();
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

    // 3. Go to Edit Profile
    await page.getByRole("button", { name: "Edit Profile" }).click();
    await expect(page).toHaveURL(/.*edit-profile/, { timeout: 10000 });

    // 4. Click back arrow
    await page.locator("button").filter({ has: page.locator("svg") }).first().click();

    // 5. Should be on Dashboard
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });
  });
});
