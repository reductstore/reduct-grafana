export const DIRECTIVES = [
  { name: '#ctx_before', description: 'Include records before match', insertText: '"#ctx_before": 10' },
  { name: '#ctx_after', description: 'Include records after match', insertText: '"#ctx_after": 5' },
  {
    name: '#select_labels',
    description: 'Select specific labels',
    insertText: '"#select_labels": ["label1", "label2"]',
  },
  { name: '#batch_size', description: 'Batch size limit', insertText: '"#batch_size": "5MB"' },
  { name: '#batch_records', description: 'Batch record limit', insertText: '"#batch_records": 1000' },
  { name: '#batch_timeout', description: 'Batch timeout', insertText: '"#batch_timeout": "200ms"' },
  {
    name: '#batch_metadata_size',
    description: 'Batch metadata size limit',
    insertText: '"#batch_metadata_size": "500KB"',
  },
  { name: '#record_timeout', description: 'Record processing timeout', insertText: '"#record_timeout": "100ms"' },
  {
    name: '#ext',
    description: 'Extension parameters',
    insertText: '"#ext": {\n  "extension_name": {\n    "param": "value"\n  }\n}',
  },
];
