import { test, expect } from '@grafana/plugin-e2e';
import { getCompletionProvider } from '../src/components/json-editor/reductstore';

test.describe('ReductStore Monaco Completion Provider', () => {
  test.beforeEach(() => {
    // Mock Monaco globally
    (global as any).window = {};
    (window as any).monaco = {
      languages: {
        CompletionItemKind: {
          Snippet: 17,
          Property: 9,
          Operator: 25,
          Value: 12,
        },
      },
    };
  });

  function mockModel(lines: string[]) {
    return {
      getLineContent: (line: number) => lines[line - 1] || '',
    };
  }

  test('suggests EXAMPLES when document is empty', () => {
    const provider = getCompletionProvider();
    const model = mockModel(['']);

    const result = provider.provideCompletionItems(model, {
      lineNumber: 1,
      column: 1,
    });

    expect(result.suggestions.length).toBeGreaterThan(1);

    // All example labels must appear
    const labels = result.suggestions.map((s) => s.label.toLowerCase());
    expect(labels).toContain('simple label comparison');
    expect(labels).toContain('numeric range filter');
  });

  test('suggests operators inside quotes', () => {
    const provider = getCompletionProvider();
    const model = mockModel(['  "&temp": "$']);

    const result = provider.provideCompletionItems(model, {
      lineNumber: 1,
      column: 14,
    });

    const labels = result.suggestions.map((s) => s.label);

    expect(labels).toContain('$eq');
    expect(labels).toContain('$contains');
    expect(labels).toContain('&label_name');
    expect(labels).toContain('@computed_label');
  });

  test('suggests values after colon', () => {
    const provider = getCompletionProvider();
    const model = mockModel(['  "&sensor": ']);

    const result = provider.provideCompletionItems(model, {
      lineNumber: 1,
      column: 13,
    });

    const labels = result.suggestions.map((s) => s.label);

    expect(labels).toContain('String value');
    expect(labels).toContain('Numeric value');
    expect(labels).toContain('Boolean true');
  });

  test('suggests keys inside object braces', () => {
    const provider = getCompletionProvider();
    const model = mockModel(['  { ']);

    const result = provider.provideCompletionItems(model, {
      lineNumber: 1,
      column: 4,
    });

    const labels = result.suggestions.map((s) => s.label);

    expect(labels).toContain('&label_name');
    expect(labels).toContain('@computed_label');

    // Logical operators too
    expect(labels).toContain('$and');
    expect(labels).toContain('$or');
    expect(labels).toContain('$not');

    // Directives
    expect(labels).toContain('#ctx_before');
  });

  test('suggests $__interval value after $each_t key', () => {
    const provider = getCompletionProvider();
    const model = mockModel(['{ "$each_t": ']);

    const result = provider.provideCompletionItems(model, {
      lineNumber: 1,
      column: 13,
    });

    const labels = result.suggestions.map((s) => s.label);
    expect(labels).toContain('$__interval');
  });
});
