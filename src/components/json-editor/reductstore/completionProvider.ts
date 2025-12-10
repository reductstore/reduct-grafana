import {
  COMPARISON_OPERATORS,
  LOGICAL_OPERATORS,
  STRING_OPERATORS,
  ARITHMETIC_OPERATORS,
  AGGREGATION_OPERATORS,
  MISC_OPERATORS,
} from './operators';
import { DIRECTIVES } from './directives';
import { EXAMPLES } from './examples';

interface MonacoModel {
  getLineContent(lineNumber: number): string;
}

interface MonacoPosition {
  lineNumber: number;
  column: number;
}

interface MonacoRange {
  startLineNumber: number;
  endLineNumber: number;
  startColumn: number;
  endColumn: number;
}

interface MonacoCompletionItem {
  label: string;
  kind: number;
  insertText: string;
  detail?: string;
  documentation?: string;
  range: MonacoRange;
  sortText?: string;
}

interface MonacoCompletionResult {
  suggestions: MonacoCompletionItem[];
}

export const getCompletionProvider = () => {
  return {
    triggerCharacters: ['{', '"', ':', ',', ' ', '$', '&', '@', '#'],
    provideCompletionItems: (model: MonacoModel, position: MonacoPosition): MonacoCompletionResult => {
      const monaco = (window as any).monaco;
      const CompletionItemKind = monaco.languages.CompletionItemKind;

      // Get context from the current line
      const lineText = model.getLineContent(position.lineNumber);
      const textUntilPosition = lineText.substring(0, position.column - 1);
      const isInsideQuotes = (textUntilPosition.match(/"/g) || []).length % 2 === 1;
      const isEmpty = textUntilPosition.trim() === '';
      const isAfterOpenBrace = textUntilPosition.trim().endsWith('{');
      const isAfterColon = /:(\s*)$/.test(textUntilPosition) && !isInsideQuotes;
      const isAfterEachT = /"\$each_t"\s*:\s*$/.test(textUntilPosition);

      // Check if at the very start of the document
      const isDocumentStart = isEmpty && position.lineNumber === 1;

      // Calculate proper range to replace partial text
      // Find the start of the current word being typed
      let wordStart = position.column - 1;
      while (wordStart > 0) {
        const char = lineText.charAt(wordStart - 1);
        if (!/[$@&#\w]/.test(char)) {
          break;
        }
        wordStart--;
      }

      // Build suggestions based on context
      const suggestions: MonacoCompletionItem[] = [];
      const range: MonacoRange = {
        startLineNumber: position.lineNumber,
        endLineNumber: position.lineNumber,
        startColumn: wordStart + 1,
        endColumn: position.column,
      };

      // 1. When document is completely empty (suggest complete examples)
      if (isDocumentStart) {
        EXAMPLES.forEach((example, index) => {
          suggestions.push({
            label: example.name,
            kind: CompletionItemKind.Snippet,
            insertText: example.insertText,
            detail: example.description,
            documentation: 'Complete query example',
            range,
            sortText: `0${index.toString().padStart(2, '0')}`,
          });
        });
        return { suggestions };
      }

      // 2. When typing inside quotes
      if (isInsideQuotes) {
        // Label references first (highest priority: 000-001)
        suggestions.push(
          {
            label: '&label_name',
            kind: CompletionItemKind.Property,
            insertText: '&label_name',
            detail: 'Label reference',
            documentation: 'Reference to a label in the record',
            range,
            sortText: '000',
          },
          {
            label: '@computed_label',
            kind: CompletionItemKind.Property,
            insertText: '@computed_label',
            detail: 'Computed label',
            documentation: 'Reference to a computed label from extensions',
            range,
            sortText: '001',
          }
        );

        let sortIndex = 0;
        // Comparison operators next (priority 1XX)
        COMPARISON_OPERATORS.forEach((op) => {
          suggestions.push({
            label: op.name,
            kind: CompletionItemKind.Operator,
            insertText: op.insertText,
            detail: op.description,
            range,
            sortText: `1${sortIndex.toString().padStart(2, '0')}`,
          });
          sortIndex++;
        });
        // Logical operators
        LOGICAL_OPERATORS.forEach((op) => {
          suggestions.push({
            label: op.name,
            kind: CompletionItemKind.Operator,
            insertText: op.insertText,
            detail: op.description,
            range,
            sortText: `2${sortIndex.toString().padStart(2, '0')}`,
          });
          sortIndex++;
        });
        // String operators
        STRING_OPERATORS.forEach((op) => {
          suggestions.push({
            label: op.name,
            kind: CompletionItemKind.Operator,
            insertText: op.insertText,
            detail: op.description,
            range,
            sortText: `3${sortIndex.toString().padStart(2, '0')}`,
          });
          sortIndex++;
        });
        // Arithmetic operators
        ARITHMETIC_OPERATORS.forEach((op) => {
          suggestions.push({
            label: op.name,
            kind: CompletionItemKind.Operator,
            insertText: op.insertText,
            detail: op.description,
            range,
            sortText: `4${sortIndex.toString().padStart(2, '0')}`,
          });
          sortIndex++;
        });
        // Aggregation operators
        AGGREGATION_OPERATORS.forEach((op) => {
          suggestions.push({
            label: op.name,
            kind: CompletionItemKind.Operator,
            insertText: op.insertText,
            detail: op.description,
            range,
            sortText: `5${sortIndex.toString().padStart(2, '0')}`,
          });
          sortIndex++;
        });
        // Misc operators
        MISC_OPERATORS.forEach((op) => {
          suggestions.push({
            label: op.name,
            kind: CompletionItemKind.Operator,
            insertText: op.insertText,
            detail: op.description,
            range,
            sortText: `6${sortIndex.toString().padStart(2, '0')}`,
          });
          sortIndex++;
        });
        return { suggestions };
      }

      // 3. When after colon (expecting values)
      if (isAfterColon) {
        if (isAfterEachT) {
          suggestions.push({
            label: '$__interval',
            kind: CompletionItemKind.Value,
            insertText: '"$__interval"',
            detail: 'Grafana interval macro',
            documentation: 'Replaced by Grafana with an auto interval for the current time range',
            range,
            sortText: '000',
          });
        }

        // Suggest highest priority (100-104)
        suggestions.push(
          {
            label: 'String value',
            kind: CompletionItemKind.Value,
            insertText: '"value"',
            detail: 'String value',
            range,
            sortText: '100',
          },
          {
            label: 'Numeric value',
            kind: CompletionItemKind.Value,
            insertText: '100',
            detail: 'Numeric value',
            range,
            sortText: '101',
          },
          {
            label: 'Boolean true',
            kind: CompletionItemKind.Value,
            // Double quotes needed for Grafana
            insertText: '"True"',
            detail: 'Boolean value',
            range,
            sortText: '102',
          },
          {
            label: 'Boolean false',
            kind: CompletionItemKind.Value,
            // Double quotes needed for Grafana
            insertText: '"False"',
            detail: 'Boolean value',
            range,
            sortText: '103',
          }
        );
        return { suggestions };
      }

      // 4. When inside object but not in quotes (expecting keys)
      if (isAfterOpenBrace || (!isEmpty && !isInsideQuotes)) {
        // Label references (highest priority: 100-101)
        suggestions.push(
          {
            label: '&label_name',
            kind: CompletionItemKind.Property,
            insertText: '"&label_name": { "$eq": "value" }',
            detail: 'Label reference',
            documentation: 'Reference to a label in the record',
            range,
            sortText: '100',
          },
          {
            label: '@computed_label',
            kind: CompletionItemKind.Property,
            insertText: '"@computed_label": { "$gt": 0 }',
            detail: 'Computed label',
            documentation: 'Reference to a computed label from extensions',
            range,
            sortText: '101',
          }
        );

        // Logical operators (priority 2XX)
        LOGICAL_OPERATORS.forEach((op, index) => {
          suggestions.push({
            label: op.name,
            kind: CompletionItemKind.Operator,
            insertText: `"${op.name}": `,
            detail: op.description,
            range,
            sortText: `2${index.toString().padStart(2, '0')}`,
          });
        });

        // Aggregation operators (priority 3XX)
        AGGREGATION_OPERATORS.forEach((op, index) => {
          suggestions.push({
            label: op.name,
            kind: CompletionItemKind.Operator,
            insertText: `"${op.name}": `,
            detail: op.description,
            range,
            sortText: `3${index.toString().padStart(2, '0')}`,
          });
        });

        // Misc operators (priority 4XX)
        MISC_OPERATORS.forEach((op, index) => {
          suggestions.push({
            label: op.name,
            kind: CompletionItemKind.Operator,
            insertText: `"${op.name}": `,
            detail: op.description,
            range,
            sortText: `4${index.toString().padStart(2, '0')}`,
          });
        });

        // Directives (priority 9XX)
        DIRECTIVES.forEach((directive, index) => {
          suggestions.push({
            label: directive.name,
            kind: CompletionItemKind.Property,
            insertText: directive.insertText,
            detail: directive.description,
            documentation: 'Query directive',
            range,
            sortText: `9${index.toString().padStart(2, '0')}`,
          });
        });
      }

      return { suggestions };
    },
  };
};
