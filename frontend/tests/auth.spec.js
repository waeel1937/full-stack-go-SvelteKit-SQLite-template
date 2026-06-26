import { test, expect } from '@playwright/test';

test.describe('Authentication', () => {
  test('login page shows EdgeIIoT branding', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.brand-name')).toContainText('Edge');
    await expect(page.locator('input[autocomplete="username"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.locator('button.btn-signin')).toBeVisible();
  });

  test('invalid credentials shows error message', async ({ page }) => {
    await page.goto('/');
    await page.fill('input[autocomplete="username"]', 'wrong');
    await page.fill('input[type="password"]', 'wrongpassword');
    await page.click('button.btn-signin');
    await expect(page.locator('.login-err')).toBeVisible();
    await expect(page.locator('.login-err')).toContainText('Invalid');
  });

  test('valid admin login reaches dashboard', async ({ page }) => {
    await page.goto('/');
    await page.fill('input[autocomplete="username"]', 'admin');
    await page.fill('input[type="password"]', 'admin');
    await page.click('button.btn-signin');
    await expect(page.locator('h1')).toContainText('Dashboard', { timeout: 10_000 });
    await expect(page.locator('.sidebar')).toBeVisible();
  });

  test('enter key submits login form', async ({ page }) => {
    await page.goto('/');
    await page.fill('input[autocomplete="username"]', 'admin');
    await page.fill('input[type="password"]', 'admin');
    await page.press('input[type="password"]', 'Enter');
    await expect(page.locator('h1')).toContainText('Dashboard', { timeout: 10_000 });
  });

  test('logout returns to login page', async ({ page }) => {
    await page.goto('/');
    await page.fill('input[autocomplete="username"]', 'admin');
    await page.fill('input[type="password"]', 'admin');
    await page.click('button.btn-signin');
    await expect(page.locator('h1')).toContainText('Dashboard', { timeout: 10_000 });
    await page.click('button.btn-out');
    await expect(page.locator('button.btn-signin')).toBeVisible({ timeout: 5_000 });
  });
});
