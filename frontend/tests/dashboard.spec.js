import { test, expect } from '@playwright/test';

// Shared login helper
async function login(page) {
  await page.goto('/');
  await page.fill('input[autocomplete="username"]', 'admin');
  await page.fill('input[type="password"]', 'admin');
  await page.click('button.btn-signin');
  await expect(page.locator('h1')).toContainText('Dashboard', { timeout: 10_000 });
}

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('shows four metric cards', async ({ page }) => {
    await expect(page.locator('.card')).toHaveCount(4);
  });

  test('metric cards have label and value', async ({ page }) => {
    const firstCard = page.locator('.card').first();
    await expect(firstCard.locator('.card-label')).toBeVisible();
    await expect(firstCard.locator('.card-value')).toBeVisible();
  });

  test('live badge is visible', async ({ page }) => {
    await expect(page.locator('.live-pill')).toBeVisible();
  });

  test('aggregate history table is present', async ({ page }) => {
    await expect(page.locator('table')).toBeVisible();
    await expect(page.locator('thead')).toContainText('Metric');
  });

  test('sidebar shows system status after a moment', async ({ page }) => {
    await expect(page.locator('.sb-footer')).toBeVisible();
  });
});

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('alerts page loads', async ({ page }) => {
    await page.click('a[href="/alerts"]');
    await expect(page.locator('h1')).toContainText('Alerts');
  });

  test('rules page loads', async ({ page }) => {
    await page.click('a[href="/rules"]');
    await expect(page.locator('h1')).toContainText('Rules');
  });

  test('back to dashboard from rules', async ({ page }) => {
    await page.click('a[href="/rules"]');
    await page.click('a[href="/"]');
    await expect(page.locator('h1')).toContainText('Dashboard');
  });
});

test.describe('Rules management (admin)', () => {
  test.beforeEach(async ({ page }) => { await login(page); });

  test('add rule form opens and closes', async ({ page }) => {
    await page.click('a[href="/rules"]');
    await page.click('button.add');
    await expect(page.locator('.form-panel')).toBeVisible();
    await page.click('button.cancel');
    await expect(page.locator('.form-panel')).not.toBeVisible();
  });

  test('validation requires all fields', async ({ page }) => {
    await page.click('a[href="/rules"]');
    await page.click('button.add');
    await page.click('button.btn-submit');
    await expect(page.locator('.form-err')).toBeVisible();
  });

  test('can create and see a new rule', async ({ page }) => {
    await page.click('a[href="/rules"]');
    await page.click('button.add');
    await page.fill('input[placeholder="temp-warn"]', 'e2e-test-rule');
    await page.fill('input[placeholder="temperature"]', 'temperature');
    await page.fill('input[type="number"]', '90');
    await page.fill('input[placeholder*="Temperature"]', 'E2E test alert');
    await page.click('button.btn-submit');
    await expect(page.locator('.ri-id')).toContainText('e2e-test-rule', { timeout: 5_000 });
  });
});
