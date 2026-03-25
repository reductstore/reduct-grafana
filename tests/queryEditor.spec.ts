import { test, expect } from '@grafana/plugin-e2e';
import { Client } from 'reduct-js';

const REDUCTSTORE_URL = 'http://localhost:8383';
const REDUCTSTORE_TOKEN = 'dev-token';
const TEST_BUCKET = 'e2e-query-test';

test.describe('ReductStore Query Editor', () => {
  test.describe.configure({ mode: 'serial' });

  test.beforeAll(async () => {
    const client = new Client(REDUCTSTORE_URL, { apiToken: REDUCTSTORE_TOKEN });
    const bucket = await client.getOrCreateBucket(TEST_BUCKET);
    const record = await bucket.beginWrite('test-entry');
    await record.write(JSON.stringify({ value: 1 }));
  });

  test.afterAll(async () => {
    const client = new Client(REDUCTSTORE_URL, { apiToken: REDUCTSTORE_TOKEN });
    const bucket = await client.getBucket(TEST_BUCKET);
    await bucket.remove();
  });

  test('should render query editor with all fields', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await expect(page.locator('label:has-text("Bucket")')).toBeVisible();
    await expect(page.locator('label:has-text("Entry")')).toBeVisible();
    await expect(page.locator('label:has-text("Scope")')).toBeVisible();
  });

  test('should display server info from backend', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await expect(page.getByText(/Buckets:/)).toBeVisible({ timeout: 10000 });
    await expect(page.getByText(/Usage:/)).toBeVisible();
  });

  test('should load buckets from backend', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    const bucketPicker = page.getByTestId('bucket-picker');
    await bucketPicker.click();

    await expect(page.getByRole('option').first()).toBeVisible({ timeout: 10000 });
  });

  test('should load entries after selecting a bucket', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    const bucketPicker = page.getByTestId('bucket-picker');
    await bucketPicker.click();
    await page.getByRole('option', { name: TEST_BUCKET }).click();

    const entryPicker = page.getByTestId('entry-picker');
    await entryPicker.click();

    await expect(page.getByRole('option', { name: 'test-entry' })).toBeVisible({ timeout: 10000 });
  });

  test('should execute query when bucket and entry are selected', async ({
    panelEditPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    const queryReq = panelEditPage.waitForQueryDataRequest();

    const bucketPicker = page.getByTestId('bucket-picker');
    await bucketPicker.click();
    await page.getByRole('option', { name: TEST_BUCKET }).click();

    const entryPicker = page.getByTestId('entry-picker');
    await entryPicker.click();
    await page.getByRole('option', { name: 'test-entry' }).click();

    await expect(await queryReq).toBeTruthy();
  });
});
