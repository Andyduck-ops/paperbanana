import { test, expect } from '@playwright/test';

test('should load page with title', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  await expect(page).toHaveTitle(/PaperBanana/);
});

test('should display header with logo and controls', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  
  await expect(page.locator('h1:has-text("PaperBanana")')).toBeVisible();
  const themeButtons = page.locator('button[role="radio"]');
  expect(await themeButtons.count()).toBe(4);
});

test('should switch between all 4 themes', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  
  const themeButtons = page.locator('button[role="radio"]');
  const themes = ['swiss', 'bauhaus', 'neo-minimal', 'art-deco'];
  
  for (let i = 0; i < 4; i++) {
    await themeButtons.nth(i).click();
    await page.waitForTimeout(300);
    const theme = await page.locator('html').getAttribute('data-theme');
    expect(theme).toBe(themes[i]);
  }
});

test('should toggle language', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  
  const enBtn = page.locator('button:has-text("EN")');
  const zhBtn = page.locator('button:has-text("ZH")');
  
  await enBtn.click();
  await page.waitForTimeout(300);
  await zhBtn.click();
  await page.waitForTimeout(300);
});

test('should open and close history panel', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  
  const historyBtn = page.locator('button[aria-label*="历史"], button:has-text("历史记录")').first();
  await historyBtn.click();
  await page.waitForTimeout(500);
  
  const historyPanel = page.locator('dialog, [role="dialog"]');
  expect(await historyPanel.count()).toBeGreaterThan(0);
  
  const closeBtn = page.locator('button[aria-label*="关闭"], button:has-text("关闭")');
  if (await closeBtn.count() > 0) {
    await closeBtn.first().click();
    await page.waitForTimeout(300);
  }
});

test('should switch between generate and refine modes', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  
  const generateBtn = page.locator('button:has-text("生成")').first();
  const refineBtn = page.locator('button:has-text("精修")').first();
  
  await refineBtn.click();
  await page.waitForTimeout(300);
  await generateBtn.click();
  await page.waitForTimeout(300);
});

test('should display input form in generate mode', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  
  const methodTextarea = page.locator('textarea').first();
  await expect(methodTextarea).toBeVisible();
});

test('should display footer', async ({ page }) => {
  await page.goto('http://localhost:5173');
  await page.waitForTimeout(1000);
  
  const footer = page.locator('footer, [role="contentinfo"]');
  await expect(footer).toBeVisible();
});
