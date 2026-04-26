import { test, expect } from '@playwright/test';

test('should load page with title', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(1000);
  await expect(page).toHaveTitle(/PaperBanana/);
});

test('should display header with logo', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(1000);

  await expect(page.locator('span:has-text("PaperBanana")')).toBeVisible();
});

test('should toggle language', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(1000);
  
  const enBtn = page.locator('button:has-text("EN")');
  const zhBtn = page.locator('button:has-text("ZH")');
  
  await enBtn.click();
  await page.waitForTimeout(300);
  await zhBtn.click();
  await page.waitForTimeout(300);
});

test('should open history panel', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(1000);
  
  const historyBtn = page.locator('button[aria-label*="历史"], button:has-text("历史记录")').first();
  await historyBtn.click();
  await page.waitForTimeout(500);
  
  const historyPanel = page.locator('[role="dialog"]');
  await expect(historyPanel.first()).toBeVisible();
});

test('should switch between generate and refine modes', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(1000);
  
  const generateBtn = page.locator('button:has-text("生成")').first();
  const refineBtn = page.locator('button:has-text("精修")').first();
  
  await refineBtn.click();
  await page.waitForTimeout(300);
  await generateBtn.click();
  await page.waitForTimeout(300);
});

test('should display input form in generate mode', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(1000);
  
  const methodTextarea = page.locator('textarea').first();
  await expect(methodTextarea).toBeVisible();
});

test('should display footer', async ({ page }) => {
  await page.goto('/');
  await page.waitForTimeout(1000);
  
  const footer = page.locator('footer, [role="contentinfo"]');
  await expect(footer).toBeVisible();
});
