import { test, expect } from '@playwright/test';

test.describe('PaperBanana UI Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('http://localhost:5173');
    await page.waitForTimeout(2000);
  });

  test('should load workspace page', async ({ page }) => {
    await expect(page).toHaveTitle(/PaperBanana/);
    await page.screenshot({ path: 'test-results/workspace.png', fullPage: true });
  });

  test('should display theme selector', async ({ page }) => {
    const themeButtons = page.locator('button[role="radio"]');
    const count = await themeButtons.count();
    expect(count).toBeGreaterThanOrEqual(4);
  });

  test('should switch themes', async ({ page }) => {
    const themeButtons = page.locator('button[role="radio"]');
    
    // Click each theme and take screenshot
    for (let i = 0; i < 4; i++) {
      await themeButtons.nth(i).click();
      await page.waitForTimeout(500);
      const theme = await page.locator('html').getAttribute('data-theme');
      console.log(`Theme ${i}: ${theme}`);
    }
  });
});
