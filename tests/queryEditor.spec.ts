import { test, expect } from '@grafana/plugin-e2e';
import { picker, clickOption } from './helpers/selectHelpers';

test.describe('ReductStore Query Editor', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/datasources/**/resources/listBuckets', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{ name: 'test-bucket' }]),
      });
    });

    await page.route('**/api/datasources/**/resources/listEntries', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([{ name: 'test-entry' }]),
      });
    });

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

    await picker(page, 'Bucket').click();
    await clickOption(page, 'bucket-picker-option-test-bucket', 'test-bucket');

    await picker(page, 'Entry').click();
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

    await picker(page, 'Bucket').click();
    await clickOption(page, 'bucket-picker-option-test-bucket', 'test-bucket');

    await picker(page, 'Entry').click();
    await clickOption(page, 'entry-picker-option-test-entry', 'test-entry');

    await expect(await queryReq).toBeTruthy();
  });

  test('should trigger query when scope is changed', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await picker(page, 'Bucket').click();
    await clickOption(page, 'bucket-picker-option-test-bucket', 'test-bucket');

    await picker(page, 'Entry').click();
    await clickOption(page, 'entry-picker-option-test-entry', 'test-entry');

    const queryReq = panelEditPage.waitForQueryDataRequest();

    await picker(page, 'Scope').click();
    await clickOption(page, 'scope-picker-option-content-only', 'Content Only');

    await expect(await queryReq).toBeTruthy();
  });

  test('should trigger query when JSON editor changes full query', async ({
    panelEditPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await picker(page, 'Bucket').click();
    await clickOption(page, 'bucket-picker-option-test-bucket', 'test-bucket');

    await picker(page, 'Entry').click();
    await clickOption(page, 'entry-picker-option-test-entry', 'test-entry');

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

    await page.getByRole('button', { name: /format query/i }).click();

    const runButton = page.getByRole('button', { name: /^run query$/i });
    await expect(runButton).toBeEnabled();

    const queryReq = panelEditPage.waitForQueryDataRequest();

    await runButton.click();

    await expect(await queryReq).toBeTruthy();
  });
});
