import { expect, Page } from "@playwright/test";

function generateUniqueEmail(prefix: string): string {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 1000);
  return `${prefix}.${timestamp}.${random}@example.com`;
}

export async function RegisterAndloginAsUser(page: Page): Promise<String> {
  const uniqueEmail = generateUniqueEmail('user');
  const password = "Password123!";

  await page.goto("/register");
  await page.evaluate(() => localStorage.clear());

  await page.getByPlaceholder('First Name').fill('Test');
  await page.getByPlaceholder('Last Name').fill('User');
  await page.getByPlaceholder('Email').fill(uniqueEmail);
  await page.getByPlaceholder('Password', { exact: true }).fill(password);
  await page.getByPlaceholder('Confirm Password').fill(password);

  const registerResponsePromise = page.waitForResponse(
    (response) => response.url().includes("/api/register") && (response.status() === 200 || response.status() === 201),
    { timeout: 10000 }
  );
  await page.getByRole('button', { name: 'Create my Café' }).click();
  await registerResponsePromise;

  await page.waitForURL(/.*login/, { timeout: 10000 });
  
  // Clear any stale auth state before filling credentials
  await page.evaluate(() => localStorage.clear());
  
  const emailInput = page.getByPlaceholder("Email").locator("visible=true");
  const passwordInput = page.getByPlaceholder("Password", { exact: true }).locator("visible=true");

  // Robust fill-and-verify strategy to counter React state resets during mounting
  await emailInput.fill(uniqueEmail.toString());
  await passwordInput.fill(password);
  
  // Verify values; if they get cleared by a re-render, re-fill them
  const filledEmail = await emailInput.inputValue();
  if (filledEmail !== uniqueEmail) {
    // If React wiped out the input on mount, safely refill the values
    await emailInput.fill(uniqueEmail);
    await passwordInput.fill(password);
  }

  // 6. Make sure the value is locked in place before clicking
  await expect(emailInput).toHaveValue(uniqueEmail, { timeout: 10000 });
  await expect(passwordInput).toHaveValue(password, { timeout: 10000 });

  // Click login and wait for the API response to finish
  const loginResponsePromise = page.waitForResponse(
    (response) => response.url().includes("/api/login") && response.status() === 200,
    { timeout: 15000 }
  );
  await page.getByRole("button", { name: "Enter the Café" }).click();
  await loginResponsePromise;

  // Now wait for React Router to redirect after auth state updates
  await page.waitForURL(/.*home/, { timeout: 10000 });
  
  await page.getByTestId("nav-dashboard").waitFor({ state: "visible", timeout: 10000 });
  return uniqueEmail;
}

export async function CreateAndLoginAsAdmin(page: Page): Promise<String> {
  const uniqueAdminEmail = generateUniqueEmail('admin');
  const password = "AdminPassword123!";

  // 1. Login with admin user to create a new user
  await page.goto("/login");
  await page.evaluate(() => localStorage.clear());
  
  const adminEmailInput = page.getByPlaceholder("Email");
  const adminPasswordInput = page.getByPlaceholder("Password", { exact: true });

  await adminEmailInput.fill("admin@admin.com");
  await adminPasswordInput.fill("admin123");
  
  const initialLoginPromise = page.waitForResponse(
    (res) => res.url().includes("/api/login") && res.status() === 200,
    { timeout: 15000 }
  );
  await page.getByRole("button", { name: "Enter the Café" }).click();
  await initialLoginPromise;
  await page.waitForURL(/.*adminDashboard/, { timeout: 10000 });

  // 2. Open modal HIRE STAFF
  await page.getByRole("button", { name: /HIRE NEW STAFF/i }).click();
  await expect(page.getByText("Create New Staff")).toBeVisible({ timeout: 10000 });

  await page.getByPlaceholder("First Name").fill("E2E");
  await page.getByPlaceholder("Last Name").fill("Admin");
  await page.getByTestId("create-user-email").fill(uniqueAdminEmail);
  await page.getByPlaceholder("Password", { exact: true }).fill(password);
  await page.getByPlaceholder("Confirm Password", { exact: true }).fill(password);
  await page.locator('select').selectOption("admin");

  // 5. Create user
  await page.getByRole("button", { name: "Create User" }).click();

  // 6. Wait for modal to close and user to appear in list
  await expect(page.getByText("Create New Staff")).not.toBeVisible({ timeout: 10000 });

  // LOGOUT
  const logoutButton = page.getByRole("button", { name: /logout|cerrar sesión/i });
  if (await logoutButton.isVisible()) {
    await logoutButton.click();
  } else {
    await page.evaluate(() => localStorage.clear());
    await page.goto("/login");
  }

  await page.waitForURL(/.*login/, { timeout: 10000 });

  // Login with the new admin
  const newEmailInput = page.getByPlaceholder("Email").locator("visible=true");
  const newPasswordInput = page.getByPlaceholder("Password", { exact: true }).locator("visible=true");
  
  await newEmailInput.fill(uniqueAdminEmail);
  await newPasswordInput.fill(password);

  const finalLoginPromise = page.waitForResponse(
    (res) => res.url().includes("/api/login") && res.status() === 200,
    { timeout: 15000 }
  );
  await page.getByRole("button", { name: "Enter the Café" }).click();
  await finalLoginPromise;

  await page.waitForURL(/.*adminDashboard/, { timeout: 10000 });
  
  // Wait for post-login stabilizers
  await page.waitForLoadState('networkidle');

  return uniqueAdminEmail;
}
