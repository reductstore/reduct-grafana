import React from 'react';
import { useTheme2 } from '@grafana/ui';
import { Controlled as CodeMirror } from 'react-codemirror2';
import { css } from '@emotion/css';
import 'codemirror/lib/codemirror.css';
import 'codemirror/theme/dracula.css';
import 'codemirror/mode/javascript/javascript';
import 'codemirror/addon/edit/matchbrackets';
import { parseJson, stringifyJson } from '../utils/json';

interface CodeEditorProps {
  value: string;
  onChange: (value: string) => void;
  onBlur?: (value: string) => void;
  onError?: (error: string) => void;
  placeholder?: string;
}

export function CodeEditor({ value, onChange, onBlur, onError, placeholder = '{}' }: CodeEditorProps) {
  const theme = useTheme2();

  const getCodeMirrorStyles = () => css`
    .CodeMirror {
      max-height: 200px;
    }
  `;

  const handleBlur = (editor: any) => {
    const editorValue = editor.getValue().trim();

    if (editorValue === '') {
      const fallback = placeholder;
      editor.setValue(fallback);
      if (onBlur) {
        onBlur(fallback);
      }
      if (onError) {
        onError('');
      }
      return;
    }

    try {
      const parsed = parseJson(editorValue);
      const pretty = stringifyJson(parsed);
      editor.setValue(pretty);
      if (onBlur) {
        onBlur(pretty);
      }
      if (onError) {
        onError('');
      }
    } catch (err: any) {
      if (onError) {
        onError(err.message);
      }
    }
  };

  return (
    <div className={getCodeMirrorStyles()}>
      <CodeMirror
        value={value}
        options={{
          mode: { name: 'javascript', json: true },
          theme: theme.isDark ? 'dracula' : 'default',
          lineNumbers: true,
          lineWrapping: true,
          viewportMargin: Infinity,
          matchBrackets: true,
          readOnly: false,
          indentUnit: 2,
          tabSize: 2,
        }}
        onBeforeChange={(_, __, newValue: string) => onChange(newValue)}
        onBlur={handleBlur}
      />
    </div>
  );
}
