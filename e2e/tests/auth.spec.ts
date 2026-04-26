import { test, expect } from "@playwright/test";

test.describe("Authentication and Home", () => {
  test("should load the login page by default", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(
      /.*login/);
    await expect(page.locator("h1")).toContainText("FocusCafe");
    await expect(page.getByPlaceholder("Email")).toBeVisible();
    await expect(page.getByPlaceholder("Password")).toBeVisible();
    await expect(page.getByRole("button", { name: "Enter the Café" })).toBeVisible();
  });

  test("should navigate to registration page", async ({ page }) => {
    await page.goto("/login");
    await page.click("text=New here? Sign up");
    await expect(page).toHaveURL(/.*register/);
    await expect(page.locator("h1")).toContainText("FocusCafe");
    await expect(page.getByPlaceholder("First Name")).toBeVisible();
    await expect(page.getByPlaceholder("Last Name")).toBeVisible();
    await expect(page.getByRole("button", { name: "Create my Café" })).toBeVisible();
  });

  test("should navigate back to login from registration", async ({ page }) => {
    await page.goto("/register");
    await page.click("text=Already have an account? Sign in");
    await expect(page).toHaveURL(/.*login/);
  });

  test("should redirect to login when accessing home without auth", async ({ page }) => {
    await page.goto("/home");
    await expect(page).toHaveURL(/.*login/);
  });
});
