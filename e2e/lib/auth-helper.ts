import { Page } from "@playwright/test";

export async function loginAsUser(page: Page) {
  await page.goto("/login");
  
  // Clear any stale auth state before filling credentials
  await page.evaluate(() => localStorage.clear());
  
  await page.getByPlaceholder("Email").fill("user@user.com");
  await page.getByPlaceholder("Password", { exact: true }).fill("user123");

  // Click login and wait for the API response to finish
  const loginResponsePromise = page.waitForResponse(
    (response) => response.url().includes("/api/login") && response.status() === 200,
    { timeout: 10000 }
  );
  await page.getByRole("button", { name: "Enter the Café" }).click();
  await loginResponsePromise;

  // Now wait for React Router to redirect after auth state updates
  await page.waitForURL(/.*home/, { timeout: 10000 });
  
  // Small delay to let backend stabilize between rapid test logins
  await page.waitForTimeout(300);
}

export async function loginAsAdmin(page: Page) {
  await page.goto("/login");
  
  // Clear any stale auth state before filling credentials
  await page.evaluate(() => localStorage.clear());
  
  await page.getByPlaceholder("Email").fill("admin@admin.com");
  await page.getByPlaceholder("Password", { exact: true }).fill("admin123");

  // Click login and wait for the API response to finish
  const loginResponsePromise = page.waitForResponse(
    (response) => response.url().includes("/api/login") && response.status() === 200,
    { timeout: 10000 }
  );
  await page.getByRole("button", { name: "Enter the Café" }).click();
  await loginResponsePromise;

  // Now wait for React Router to redirect after auth state updates
  await page.waitForURL(/.*adminDashboard/, { timeout: 10000 });
  
  // Small delay to let backend stabilize between rapid test logins
  await page.waitForTimeout(300);
}
