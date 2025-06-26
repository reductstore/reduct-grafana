import React from 'react';
import { QueryEditorProps } from '@grafana/data';
import { DataSource } from '../datasource';
import { MyDataSourceOptions, MyQuery } from '../types';
import { Stack } from '@grafana/ui';

type Props = QueryEditorProps<DataSource, MyQuery, MyDataSourceOptions>;


export function QueryEditor({ query, onChange, onRunQuery }: Props) {
  
  return (
    <Stack direction="column" gap={2}>
    </Stack>
  );
}
