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
  // If backend provides a complete ECharts config with series data, use it directly
  if (config.series && config.series.length > 0) {
    const option: Record<string, unknown> = {};

    if (config.title) {
      option.title = { text: config.title, left: 'center' };
    }

    if (config.tooltip) {
      option.tooltip = config.tooltip;
    }

    if (config.legend) {
      option.legend = config.legend;
    }

    if (config.colors) {
      option.color = config.colors;
    }

    if (config.x_axis) {
      option.xAxis = {
        type: config.x_axis.type,
        data: config.x_axis.data,
        name: config.x_axis.name,
      };
    }

    if (config.y_axis) {
      option.yAxis = {
        type: config.y_axis.type,
        name: config.y_axis.name,
      };
    }

    option.series = config.series.map((s) => ({
      name: s.name,
      type: s.type,
      data: s.data,
      ...(s.label ? { label: s.label } : {}),
      ...(s.color ? { color: s.color } : {}),
    }));

    return option;
  }

  // Fallback: build simple chart from rows data
  if (!columns || !rows || rows.length === 0) return {};

  const xField = columns[0]?.name || 'x';
  const yField = columns[1]?.name || 'y';

  if (config.type === 'pie_chart') {
    const pieData = rows.map((r) => ({
      name: String(r[xField] ?? ''),
      value: Number(r[yField]) || 0,
    }));
    return {
      title: config.title ? { text: config.title, left: 'center' } : undefined,
      tooltip: { trigger: 'item' as const },
      series: [{ type: 'pie' as const, radius: '50%', data: pieData }],
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
        type: config.type === 'line_chart' ? ('line' as const) : ('bar' as const),
        ...(config.type === 'line_chart' ? { smooth: true } : {}),
      },
    ],
  };
}
