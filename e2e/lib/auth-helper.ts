import { Page } from "@playwright/test";

export async function loginAsUser(page: Page) {
  await page.goto("/login");
  await page.getByPlaceholder("Email").fill("user@user.com");
  await page.getByPlaceholder("Password", { exact: true }).fill("user123");
  await page.getByRole("button", { name: "Enter the Café" }).click();
  await page.waitForURL(/.*home/);
}

export async function loginAsAdmin(page: Page) {
  await page.goto("/login");
  await page.getByPlaceholder("Email").fill("admin@admin.com");
  await page.getByPlaceholder("Password", { exact: true }).fill("admin123");
  await page.getByRole("button", { name: "Enter the Café" }).click();
  await page.waitForURL(/.*adminDashboard/);
}
