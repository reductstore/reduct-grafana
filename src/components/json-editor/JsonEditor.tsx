import { JsonEditorCore } from './JsonEditorCore';
import React, { useState } from 'react';
import { ReductQuery } from '../../types';
import { DataSource } from '../../datasource';
import { JsonToolbox } from './JsonToolbox';
import { Modal, useStyles2, useTheme2 } from '@grafana/ui';
import { css } from '@emotion/css';
import AutoSizer from 'react-virtualized-auto-sizer';
import { useMeasure } from 'react-use';

interface JsonEditorProps {
  query: ReductQuery;
  onChange: (q: ReductQuery, processQuery: boolean) => void;
  datasource: DataSource;
}

export function JsonEditor({
  query,
  onChange,
  datasource,
}: JsonEditorProps) {
  const theme = useTheme2();
  const styles = useStyles2(getStyles);

  const [isExpanded, setIsExpanded] = useState(false);
  const [toolboxRef, toolboxMeasure] = useMeasure<HTMLDivElement>();

  const renderQueryEditor = (width?: number, height?: number) => {
    return (
      <div className={styles.queryEditorContainer} style={{ width, height }}>
        <div className={styles.editorWrapper}>
          <JsonEditorCore
            query={query}
            width={width}
            height={height ? height - toolboxMeasure.height : undefined}
            onChange={onChange}
          >
            {(formatCode) => (
              <div ref={toolboxRef}>
                <JsonToolbox
                  query={query}
                  formatCode={formatCode}
                  onExpand={setIsExpanded}
                  isExpanded={isExpanded}
                  datasourceId={datasource.id}
                />
              </div>
            )}
          </JsonEditorCore>
        </div>
      </div>
    );
  };

  const renderEditor = (isModal = false) => {
    return (
      <div className={isModal ? styles.modalEditor : styles.editorContainer}>
        <AutoSizer>
          {({ width, height }: { width: number; height: number }) => {
            return renderQueryEditor(width, height);
          }}
        </AutoSizer>
      </div>
    );
  };

  const renderPlaceholder = () => {
    return (
      <div
        style={{
          width: '100%',
          height: '200px',
          background: theme.colors.background.primary,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        Editing in expanded JSON condition editor
      </div>
    );
  };

  return (
    <>
      {isExpanded ? renderPlaceholder() : renderEditor()}
      {isExpanded && (
        <Modal
          title={`JSON Condition Editor - Query ${query.refId}`}
          closeOnBackdropClick={false}
          closeOnEscape={false}
          className={styles.modal}
          contentClassName={styles.modalContent}
          isOpen={isExpanded}
          onDismiss={() => {
            setIsExpanded(false);
          }}
        >
          {renderEditor(true)}
        </Modal>
      )}
    </>
  );
}

function getStyles() {
  return {
    modal: css`
      width: 95vw;
      height: 95vh;
    `,
    modalContent: css`
      height: 100%;
      padding-top: 0;
    `,
    modalEditor: css`
      height: 100%;
      width: 100%;
    `,
    editorContainer: css`
      height: 200px;
      width: 100%;
      min-height: 200px;
    `,
    queryEditorContainer: css`
      display: flex;
      flex-direction: column;
      height: 100%;
    `,
    editorWrapper: css`
      flex: 1;
      display: flex;
      flex-direction: column;
    `,
  };
}
