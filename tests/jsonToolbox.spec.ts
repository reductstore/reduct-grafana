import { test, expect } from '@grafana/plugin-e2e';
import { picker, clickOption } from './helpers/selectHelpers';

test.describe('ReductStore JSON Toolbox', () => {
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

    await page.route('**/api/datasources/**/resources/validateCondition', async (route) => {
      const body = await route.request().postDataJSON();
      const isValid = typeof body.condition === 'object' && body.condition !== null && !Array.isArray(body.condition);

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          valid: isValid,
          error: isValid ? undefined : 'Malformed JSON',
        }),
      });
    });
  });

  test('shows missing bucket/entry validation messages', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);
    await expect(page.getByText(/select bucket and entry/i)).toBeVisible();
  });

  test('valid JSON is sent as parsed object to backend validator', async ({
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

    const editor = page.getByRole('textbox', { name: /editor content/i });
    const requestPromise = page.waitForRequest('**/validateCondition');
    await editor.fill('{ "&sensor": { "$eq": "ok" } }');

    const req = await requestPromise;
    const body = req.postDataJSON();
    expect(body.condition).toEqual({ '&sensor': { $eq: 'ok' } });
  });

  test('sends invalid JSON to backend', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await picker(page, 'Bucket').click();
    await clickOption(page, 'bucket-picker-option-test-bucket', 'test-bucket');

    await picker(page, 'Entry').click();
    await clickOption(page, 'entry-picker-option-test-entry', 'test-entry');

    const editor = page.getByRole('textbox', { name: /editor content/i });
    const requestPromise = page.waitForRequest('**/validateCondition');

    await editor.fill('not json at all');

    const req = await requestPromise;
    const postData = req.postDataJSON();
    expect(postData.condition).toBe('not json at all');
  });

  test('formats JSON with Monaco', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await picker(page, 'Bucket').click();
    await clickOption(page, 'bucket-picker-option-test-bucket', 'test-bucket');

    await picker(page, 'Entry').click();
    await clickOption(page, 'entry-picker-option-test-entry', 'test-entry');

    const editor = page.getByRole('textbox', { name: /editor content/i });
    await editor.fill('{"a":1}');

    const content = await editor.inputValue();
    expect(content).toMatch(/\n/);
  });

  test('expands and collapses the JSON editor', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await picker(page, 'Bucket').click();
    await clickOption(page, 'bucket-picker-option-test-bucket', 'test-bucket');

    await picker(page, 'Entry').click();
    await clickOption(page, 'entry-picker-option-test-entry', 'test-entry');

    await page.getByRole('button', { name: /expand editor/i }).click();

    await expect(page.getByRole('heading', { name: /json condition editor/i })).toBeVisible();

    await page.getByRole('button', { name: /collapse editor/i }).click();

    await expect(page.getByText(/editing in expanded json condition editor/i)).not.toBeVisible();
  });
});
