import { useEffect, useRef, useState } from 'react';
import { Card, Spin } from 'antd';
import * as echarts from 'echarts';
import type { ChartPanelConfig } from './dashboard-config';
import { getDevOpsDashboardAPI } from '../../api/client';

interface ChartPanelProps {
  config: ChartPanelConfig;
}

export default function ChartPanel({ config }: ChartPanelProps) {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstance = useRef<echarts.ECharts | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!chartRef.current) return;
    chartInstance.current = echarts.init(chartRef.current, 'dark');

    const api = getDevOpsDashboardAPI();
    let promise;

    switch (config.api) {
      case 'getDashboardTrend':
        promise = api.getDashboardTrend();
        break;
      default:
        setLoading(false);
        return;
    }

    promise.then((res) => {
      const data = res.data;
      const option: echarts.EChartsOption = {
        backgroundColor: 'transparent',
        grid: { top: 40, right: 20, bottom: 20, left: 40, containLabel: true },
        tooltip: {
          trigger: 'axis',
          backgroundColor: 'rgba(0,0,0,0.8)',
          borderColor: '#333',
        },
        legend: {
          data: config.series.map((s) => s.name),
          textStyle: { color: '#aaaaaa' },
        },
        xAxis: {
          type: 'category',
          data: data.timeLabels,
          axisLine: { lineStyle: { color: '#333' } },
          axisLabel: { color: '#aaaaaa' },
        },
        yAxis: {
          type: 'value',
          axisLine: { lineStyle: { color: '#333' } },
          axisLabel: { color: '#aaaaaa', formatter: '{value}%' },
          splitLine: { lineStyle: { color: '#333' } },
        },
        series: config.series.map((s) => ({
          name: s.name,
          type: (config.chartType === 'area' ? 'line' : config.chartType) as 'line' | 'bar',
          data: (data as unknown as Record<string, unknown>)[s.dataKey] as number[],
          smooth: true,
          lineStyle: { color: s.color, width: 2 },
          itemStyle: { color: s.color },
          areaStyle:
            config.chartType === 'area'
              ? { opacity: 0.1, color: s.color }
              : undefined,
        })),
      };

      chartInstance.current?.setOption(option);
      setLoading(false);
    });

    const handleResize = () => chartInstance.current?.resize();
    window.addEventListener('resize', handleResize);

    return () => {
      window.removeEventListener('resize', handleResize);
      chartInstance.current?.dispose();
    };
  }, [config]);

  return (
    <Card
      title={config.title}
      style={{
        background: '#1f1f1f',
        border: 'none',
        borderRadius: 4,
      }}
      headStyle={{
        color: '#ffffff',
        borderBottom: '1px solid #333333',
        fontSize: 16,
        fontWeight: 500,
      }}
    >
      {loading ? (
        <div style={{ height: config.height, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Spin />
        </div>
      ) : (
        <div ref={chartRef} style={{ height: config.height }} />
      )}
    </Card>
  );
}
