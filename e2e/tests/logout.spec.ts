import { test, expect } from "@playwright/test";
import { RegisterAndloginAsUser } from "../lib/auth-helper";

test.describe("Logout", () => {
  test("should logout from Home and redirect to login", async ({ page }) => {
    // 1. Login and land on Home
    await RegisterAndloginAsUser(page);
    await expect(page).toHaveURL(/.*home/, { timeout: 10000 });

    // 2. Verify token exists in localStorage
    const tokenBefore = await page.evaluate(() => localStorage.getItem("token"));
    expect(tokenBefore).not.toBeNull();

    // 3. Click Logout button in header
    await page.getByRole("button", { name: "Logout" }).click();

    // 4. Should redirect to login page
    await expect(page).toHaveURL(/.*login/, { timeout: 10000 });

    // 5. Verify token is removed from localStorage
    const tokenAfter = await page.evaluate(() => localStorage.getItem("token"));
    expect(tokenAfter).toBeNull();

    // 6. Verify login form is visible
    await expect(page.getByPlaceholder("Email")).toBeVisible({ timeout: 10000 });
    await expect(page.getByPlaceholder("Password", { exact: true })).toBeVisible({ timeout: 10000 });
  });

  test("should logout from Dashboard and redirect to login", async ({ page }) => {
    // 1. Login and navigate to Dashboard
    await RegisterAndloginAsUser(page);
    await expect(page).toHaveURL(/.*home/, { timeout: 10000 });
    await page.getByTestId("nav-dashboard").click();
    await expect(page).toHaveURL(/.*dashboard/, { timeout: 10000 });

    // 2. Click Logout button
    await page.getByRole("button", { name: "Logout" }).click();

    // 3. Should redirect to login
    await expect(page).toHaveURL(/.*login/, { timeout: 10000 });

    // 4. Verify token cleared
    const token = await page.evaluate(() => localStorage.getItem("token"));
    expect(token).toBeNull();
  });
});
