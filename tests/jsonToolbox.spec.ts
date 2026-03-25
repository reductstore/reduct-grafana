import { test, expect } from '@grafana/plugin-e2e';

test.describe('ReductStore JSON Toolbox', () => {
  test('should show validation prompt when bucket/entry not selected', async ({
    panelEditPage,
    readProvisionedDataSource,
    page,
  }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await expect(page.getByText(/select bucket and entry/i)).toBeVisible();
  });

  test('should show expand/collapse buttons', async ({ panelEditPage, readProvisionedDataSource, page }) => {
    const ds = await readProvisionedDataSource({ fileName: 'datasources.yml' });
    await panelEditPage.datasource.set(ds.name);

    await expect(page.getByRole('button', { name: /expand editor/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /format query/i })).toBeVisible();
  });
});
