import React, { useCallback, useEffect, useRef } from 'react';
import { CodeEditor, Monaco, MonacoEditor, useTheme2 } from '@grafana/ui';
import { ReductQuery } from '../../types';
import { getCompletionProvider } from './reductstore';

type Props = {
  query: ReductQuery;
  onChange: (value: ReductQuery, processQuery: boolean) => void;
  width?: number;
  height?: number;
  children: (formatCode: () => void) => React.ReactNode;
};

// Global flag to track if completion provider has been registered
let isCompletionProviderRegistered = false;

export function JsonEditorCore({ onChange, query, width, height, children }: Props) {
  const queryRef = useRef(query);
  const editorRef = useRef<MonacoEditor | null>(null);
  const monacoRef = useRef<Monaco | null>(null);
  const theme = useTheme2();

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
    const trimmed = value.trim();
    let whenValue: any;

    if (trimmed === '') {
      whenValue = undefined;
    } else {
      try {
        whenValue = JSON.parse(trimmed);
      } catch {
        whenValue = value;
      }
    }

    const newQuery: ReductQuery = {
      ...queryRef.current,
      options: { ...queryRef.current.options, when: whenValue },
    };

    onChange(newQuery, false);
  };

  const formatCode = useCallback(() => {
    editorRef.current?.getAction('editor.action.formatDocument')?.run();
  }, []);

  const onEditorDidMount = (editor: MonacoEditor, monaco: Monaco) => {
    editorRef.current = editor;
    monacoRef.current = monaco;

    if (!isCompletionProviderRegistered) {
      // Disable default JSON completions and validation
      monaco.languages.json.jsonDefaults.setDiagnosticsOptions({ validate: false });
      monaco.languages.json.jsonDefaults.setModeConfiguration({
        completionItems: false,
        hovers: false,
        documentSymbols: false,
        colors: false,
        foldingRanges: false,
        diagnostics: false,
        selectionRanges: false,
      });

      // Register our custom completion provider
      monaco.languages.registerCompletionItemProvider('json', getCompletionProvider());
      isCompletionProviderRegistered = true;
    }

    monaco.editor.setTheme(theme.isDark ? 'grafana-dark' : 'grafana-light');
  };

  return (
    <>
      <CodeEditor
        key={theme.isDark ? 'dark' : 'light'}
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
