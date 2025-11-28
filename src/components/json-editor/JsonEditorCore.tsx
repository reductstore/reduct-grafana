import React, { useCallback, useEffect, useRef } from 'react';
import { CodeEditor, Monaco, MonacoEditor } from '@grafana/ui';
import { ReductQuery } from '../../types';
import { getCompletionProvider } from './reductstore';

type Props = {
  query: ReductQuery;
  onChange: (value: ReductQuery, processQuery: boolean) => void;
  width?: number;
  height?: number;
  children: (formatCode: () => void) => React.ReactNode;
};

export function JsonEditorCore({ onChange, query, width, height, children }: Props) {
  const queryRef = useRef(query);
  const registeredRef = useRef(false);
  const editorRef = useRef<MonacoEditor | null>(null);

  // Keep queryRef always updated
  useEffect(() => {
    queryRef.current = query;
  }, [query]);

  const editorValue =
    typeof query.options?.when === 'string'
      ? query.options.when
      : query.options?.when
      ? JSON.stringify(query.options.when, null, 2)
      : '';

  const handleEditorChange = (value: string) => {
    const trimmedValue = value.trim();
    let whenValue: any;

    if (trimmedValue === '') {
      whenValue = undefined;
    } else {
      try {
        whenValue = JSON.parse(trimmedValue);
      } catch {
        whenValue = value;
      }
    }

    const newQuery: ReductQuery = {
      ...queryRef.current,
      options: {
        ...queryRef.current.options,
        when: whenValue,
      },
    };
    onChange(newQuery, false);
  };

  const formatCode = useCallback(() => {
    if (editorRef.current) {
      editorRef.current.getAction('editor.action.formatDocument')?.run();
    }
  }, []);

  const onEditorDidMount = (editor: MonacoEditor, monaco: Monaco) => {
    if (!registeredRef.current) {
      monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
        validate: false,
        allowComments: false,
        schemas: [],
        enableSchemaRequest: false,
      });
      monaco.languages.json.jsonDefaults.setModeConfiguration({
        documentFormattingEdits: false,
        documentRangeFormattingEdits: false,
        completionItems: false,
        hovers: false,
        documentSymbols: false,
        tokens: true,
        colors: true,
        foldingRanges: true,
        diagnostics: false,
      });

      // Register our custom completion provider
      monaco.languages.registerCompletionItemProvider('json', getCompletionProvider());

      registeredRef.current = true;
    }
    editorRef.current = editor;
  };

  return (
    <>
      <CodeEditor
        width={width}
        height={height ?? 200}
        language="json"
        value={editorValue}
        onChange={handleEditorChange}
        onEditorDidMount={onEditorDidMount}
        showMiniMap={false}
        showLineNumbers={true}
        monacoOptions={{
          suggestOnTriggerCharacters: true,
          quickSuggestions: true,
          wordBasedSuggestions: true,
          formatOnPaste: false,
          formatOnType: false,
        }}
      />
      {children(formatCode)}
    </>
  );
}
