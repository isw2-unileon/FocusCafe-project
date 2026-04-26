import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  const validEmail = `testuser.${Date.now()}@example.com`;
  const validPassword = 'Password123!';
  test('should register a new user successfully and then login', async ({ page }) => {
    // Navigate to registration page
    await page.goto('/register');

    // Fill registration form
    await page.getByPlaceholder('First Name').fill('Test');
    await page.getByPlaceholder('Last Name').fill('User');
    await page.getByPlaceholder('Email').fill(validEmail);
    await page.getByPlaceholder('Password', {exact: true}).fill(validPassword);
    await page.getByPlaceholder('Confirm Password').fill(validPassword);

    // Click register button
    await page.getByRole('button', { name: 'Create my Café' }).click();

    // Verify redirection to login page after successful registration
    await page.waitForURL(/.*login/);
    await expect(page).toHaveURL(/.*login/);

    // Fill login form (using mock or existing credentials if applicable)
    await page.getByPlaceholder('Email').fill(validEmail);
    await page.getByPlaceholder('Password', {exact: true}).fill(validPassword);

    // Click login button
    await page.getByRole('button', { name: 'Enter the Café' }).click();
    // Verify redirection to home/dashboard
    // According to App.tsx, successful login should lead to a protected route like /home
    await page.waitForURL(/.*home/);
    await expect(page).toHaveURL(/.*home/);
  });

  test('should show error when registration passwords do not match', async ({ page }) => {
    await page.goto('/register');

    await page.getByPlaceholder('First Name').fill('Test');
    await page.getByPlaceholder('Last Name').fill('User');
    await page.getByPlaceholder('Email').fill('mismatch@example.com');
    await page.getByPlaceholder('Password', {exact: true}).fill(validPassword);
    await page.getByPlaceholder('Confirm Password').fill('DifferentPassword123!');

    await page.getByRole('button', { name: 'Create my Café' }).click();

    // Verify error message
    await expect(page.getByText('Las contraseñas no coinciden.')).toBeVisible();
  });

  test('should show error with invalid login credentials', async ({ page }) => {
    await page.goto('/login');

    await page.getByPlaceholder('Email').fill('invalid@example.com');
    await page.getByPlaceholder('Password', {exact: true}).fill('WrongPassword!');

    await page.getByRole('button', { name: 'Enter the Café' }).click();

    // Verify error message (assuming the service returns an error that is displayed)
    // The exact error message depends on the auth service implementation
    const errorElement = page.locator('p.text-red-500');
    await expect(errorElement).toBeVisible();
  });
});
