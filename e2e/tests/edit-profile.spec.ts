import { test, expect } from "@playwright/test";

test.use({ storageState: "playwright/.auth/user.json" });

test.describe("Edit Profile", () => {
  test("should update first and last name and reflect changes on Dashboard", async ({ page }) => {
    // Navigate to Edit Profile (already authenticated)
    await page.goto("/edit-profile");
    await expect(page).toHaveURL(/.*edit-profile/);

    // Verify current name is pre-filled
    const firstNameInput = page.getByPlaceholder("Your first name");
    const lastNameInput = page.getByPlaceholder("Your last name");
    await expect(firstNameInput).toHaveValue("Test");
    await expect(lastNameInput).toHaveValue("User");

    // Change names
    await firstNameInput.fill("E2EChanged");
    await lastNameInput.fill("NameChanged");

    // Save changes
    await page.getByRole("button", { name: "Save Changes" }).click();

    // Verify redirect to /dashboard
    await page.waitForURL(/.*dashboard/);
    await expect(page).toHaveURL(/.*dashboard/);

    // Verify updated name appears on Dashboard
    await expect(page.getByText("E2EChanged NameChanged")).toBeVisible();

    // Cleanup: revert names back to original
    await page.goto("/edit-profile");
    await page.getByPlaceholder("Your first name").fill("Test");
    await page.getByPlaceholder("Your last name").fill("User");
    await page.getByRole("button", { name: "Save Changes" }).click();
    await page.waitForURL(/.*dashboard/);

    // Verify original name is restored
    await expect(page.getByText("Test User")).toBeVisible();
  });
});
