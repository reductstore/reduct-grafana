import { test, expect } from '@grafana/plugin-e2e';
import type { Page } from '@playwright/test';

function picker(page: Page, label: string) {
  return page.locator(`label:has-text("${label}") >> xpath=following-sibling::*[1]`);
}

test.describe('ReductStore Query Editor', () => {
  test.beforeEach(async ({ page }) => {
    // Mock listBuckets
    await page.route('**/api/datasources/**/resources/listBuckets', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{ name: 'test-bucket' }]),
      });
    });

    // Mock listEntries
    await page.route('**/api/datasources/**/resources/listEntries', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{ name: 'test-entry' }]),
      });
    });

    // Mock serverInfo
    await page.route('**/api/datasources/**/resources/serverInfo', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          bucket_count: 1,
          usage: 123456,
          version: '1.0.0',
          oldest_record: Date.now() * 1000 - 50000,
          latest_record: Date.now() * 1000,
          license: null,
        }),
      });
    });
  });

  test('smoke: should render query editor', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await expect(page.locator('label:has-text("Bucket")')).toBeVisible();
  });

  test('should load entries when a bucket is selected', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    // Select bucket
    await picker(page, 'Bucket').click();
    await page.getByRole('option').first().click();

    // Select entry
    await picker(page, 'Entry').click();

    // Entries should now be loaded and visible
    await expect(page.getByRole('option')).toBeVisible();
  });

  test('should trigger query when bucket and entry selected', async ({
    panelEditPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    const queryReq = panelEditPage.waitForQueryDataRequest();

    // Select bucket
    await picker(page, 'Bucket').click();
    await page.getByRole('option').first().click();

    // Select entry
    await picker(page, 'Entry').click();
    await page.getByRole('option').first().click();

    // Query should be triggered
    await expect(await queryReq).toBeTruthy();
  });

  test('should trigger query when scope is changed', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    // Select bucket
    await picker(page, 'Bucket').click();
    await page.getByRole('option').first().click();

    // Select entry
    await picker(page, 'Entry').click();
    await page.getByRole('option').first().click();

    const queryReq = panelEditPage.waitForQueryDataRequest();

    // Change Scope
    await picker(page, 'Scope').click();
    await page.getByRole('option', { name: 'Content Only' }).click();

    await expect(await queryReq).toBeTruthy();
  });

  test('should trigger query when JSON editor changes full query', async ({
    panelEditPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    // Select a bucket
    await picker(page, 'Bucket').click();
    await page.getByRole('option', { name: 'test-bucket' }).click();

    // Select an entry
    await picker(page, 'Entry').click();
    await page.getByRole('option', { name: 'test-entry' }).click();

    const newQueryJson = JSON.stringify(
      {
        bucket: 'test-bucket',
        entry: 'test-entry',
        options: { mode: 'LabelOnly' },
      },
      null,
      2
    );

    const editor = page.getByRole('textbox', { name: /editor content/i });
    await editor.fill(newQueryJson);
    await editor.blur();

    // Format the query
    await page.getByRole('button', { name: /format query/i }).click();

    // Run the query
    const runButton = page.getByRole('button', { name: /^run query$/i });
    await expect(runButton).toBeEnabled();

    const queryReq = panelEditPage.waitForQueryDataRequest();

    await runButton.click();

    // Wait for query to complete
    await expect(await queryReq).toBeTruthy();
  });
});
