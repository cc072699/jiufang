// ============================================================
// ChartRenderer - ECharts 图表渲染 + 表格切换
// ============================================================

import { useState } from 'react';
import { Button, Space } from 'antd';
import { TableOutlined, BarChartOutlined } from '@ant-design/icons';
import ReactECharts from 'echarts-for-react';
import type { ChartConfig, ColumnDef } from '../api/types';
import DataTable from './DataTable';

interface Props {
  chartConfig: ChartConfig;
  columns?: ColumnDef[];
  rows?: Record<string, unknown>[];
}

export default function ChartRenderer({ chartConfig, columns, rows }: Props) {
  const [viewMode, setViewMode] = useState<'chart' | 'table'>('chart');

  const option = buildEChartsOption(chartConfig, columns, rows);

  return (
    <div>
      <Space style={{ marginBottom: 12 }}>
        <Button
          size="small"
          type={viewMode === 'chart' ? 'primary' : 'default'}
          icon={<BarChartOutlined />}
          onClick={() => setViewMode('chart')}
        >
          图表
        </Button>
        <Button
          size="small"
          type={viewMode === 'table' ? 'primary' : 'default'}
          icon={<TableOutlined />}
          onClick={() => setViewMode('table')}
        >
          表格
        </Button>
      </Space>
      {viewMode === 'chart' ? (
        <ReactECharts option={option} style={{ height: 400 }} />
      ) : (
        columns && rows && <DataTable columns={columns} rows={rows} />
      )}
    </div>
  );
}

function buildEChartsOption(
  config: ChartConfig,
  columns?: ColumnDef[],
  rows?: Record<string, unknown>[],
) {
  if (!columns || !rows) return {};

  const xField = config.x_field || columns[0]?.name || 'x';
  const yField = config.y_field || columns[1]?.name || 'y';

  // Pie chart uses a different ECharts structure
  if (config.chart_type === 'pie') {
    const pieData = rows.map((r) => ({
      name: String(r[xField] ?? ''),
      value: Number(r[yField]) || 0,
    }));
    return {
      title: config.title ? { text: config.title, left: 'center' } : undefined,
      tooltip: { trigger: 'item' as const },
      series: [
        {
          type: 'pie' as const,
          radius: '50%',
          data: pieData,
        },
      ],
    };
  }

  const xData = rows.map((r) => String(r[xField] ?? ''));
  const yData = rows.map((r) => Number(r[yField]) || 0);

  return {
    title: config.title ? { text: config.title, left: 'center' } : undefined,
    tooltip: { trigger: 'axis' as const },
    xAxis: { type: 'category' as const, data: xData },
    yAxis: { type: 'value' as const },
    series: [
      {
        data: yData,
        type: config.chart_type || 'bar',
        ...(config.chart_type === 'line' ? { smooth: true } : {}),
      },
    ],
  };
}
