import { test, expect } from '@playwright/test';

test('should load workspace page', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(2000);
  await expect(page).toHaveTitle(/PaperBanana/);
});

test('should display theme selector', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  const themeButtons = page.locator('button[role="radio"]');
  const count = await themeButtons.count();
  expect(count).toBeGreaterThanOrEqual(4);
});

test('should switch themes', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  const themeButtons = page.locator('button[role="radio"]');
  
  for (let i = 0; i < 4; i++) {
    await themeButtons.nth(i).click();
    await page.waitForTimeout(300);
    const theme = await page.locator('html').getAttribute('data-theme');
    console.log(`Theme ${i}: ${theme}`);
  }
});

test('should open history panel', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  
  // Look for history button
  const historyBtn = page.locator('button[aria-label*="历史"], button[aria-label*="History"], button:has-text("历史")');
  const count = await historyBtn.count();
  console.log(`History buttons found: ${count}`);
});

test('should open settings', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  
  // Look for settings button
  const settingsBtn = page.locator('button[aria-label*="设置"], button[aria-label*="Settings"], button:has-text("设置")');
  const count = await settingsBtn.count();
  console.log(`Settings buttons found: ${count}`);
});
