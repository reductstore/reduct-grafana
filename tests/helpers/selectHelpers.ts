import type { Page } from '@playwright/test';

export function picker(page: Page, label: string) {
  return page.locator(`label:has-text("${label}") >> xpath=following-sibling::*[1]`);
}

export async function clickOption(page: Page, testId: string, label: string) {
  const byTestId = page.getByTestId(testId);

  if ((await byTestId.count()) > 0) {
    await byTestId.first().click();
    return;
  }

  await page.getByRole('option', { name: label }).click();
}
