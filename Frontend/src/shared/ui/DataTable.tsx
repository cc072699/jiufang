// ============================================================
// DataTable - 查询结果表格（含空值展示规则）
// ============================================================

import { Table, Tooltip } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import type { ColumnDef } from '../api/types';
import { formatNullValue, isMonetaryField, isNullValue } from '../utils/nullDisplay';

interface Props {
  columns: ColumnDef[];
  rows: Record<string, unknown>[];
  maxRows?: number;
}

export default function DataTable({ columns, rows, maxRows = 500 }: Props) {
  const displayRows = rows.slice(0, maxRows);

  const antColumns: ColumnsType<Record<string, unknown>> = columns.map((col) => ({
    title: col.name,
    dataIndex: col.name,
    key: col.name,
    render: (val: unknown) => {
      const formatted = formatNullValue(val, col.type);
      const isMonetary = isMonetaryField(col.name);
      const isEmpty = isNullValue(val);

      if (isMonetary && isEmpty) {
        return (
          <Tooltip title="该字段无数据">
            <span style={{ color: '#999' }}>{formatted}</span>
          </Tooltip>
        );
      }

      if (isEmpty && col.type !== 'number') {
        return <span style={{ color: '#d9d9d9' }}>—</span>;
      }

      return String(formatted);
    },
  }));

  return (
    <div>
      <Table
        columns={antColumns}
        dataSource={displayRows}
        rowKey={(row) => columns.map((c) => JSON.stringify(row[c.name] ?? null)).join('|')}
        size="small"
        scroll={{ x: 'max-content' }}
        pagination={rows.length > 50 ? { pageSize: 50, showSizeChanger: true } : false}
      />
      {rows.length > maxRows && (
        <div style={{ textAlign: 'center', padding: 8, color: '#999' }}>
          仅展示前 {maxRows} 行，完整数据请导出
        </div>
      )}
    </div>
  );
}
