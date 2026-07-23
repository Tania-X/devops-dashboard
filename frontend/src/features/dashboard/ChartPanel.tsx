import { useEffect, useRef, useState } from 'react';
import { Card, Spin } from 'antd';
import * as echarts from 'echarts';
import type { ChartPanelConfig } from './dashboard-config';
import { getDevOpsDashboardAPI } from '../../api/client';

interface ChartPanelProps {
  config: ChartPanelConfig;
}

/**
 * 降采样：将 N 个连续点合并为一个平均值点
 * 后端 10s 打一个点，6 小时 = 2160 点，按 1 分钟聚合后约 360 点
 */
function downsample(data: number[], labels: string[], pointsPerGroup: number) {
  const result: number[] = [];
  const resultLabels: string[] = [];

  for (let i = 0; i < data.length; i += pointsPerGroup) {
    const group = data.slice(i, i + pointsPerGroup);
    const avg = group.reduce((a, b) => a + b, 0) / group.length;
    result.push(avg);
    resultLabels.push(labels[i]);
  }

  return { data: result, labels: resultLabels };
}

/** 格式化 "HH:MM:SS" → "HH:MM" */
function formatLabel(label: string) {
  return label.length > 5 ? label.substring(0, 5) : label;
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

    const fetchAndRender = () => {
      promise
        .then((res) => {
          const data = res.data;

          // 降采样：每 6 个点（60秒）合并为一个平均值点
          const dsCpu = downsample(
            (data as unknown as Record<string, number[]>)['cpuData'] as number[],
            (data as unknown as Record<string, string[]>)['timeLabels'] as string[],
            6,
          );
          const dsMemory = downsample(
            (data as unknown as Record<string, number[]>)['memoryData'] as number[],
            (data as unknown as Record<string, string[]>)['timeLabels'] as string[],
            6,
          );

          const option: echarts.EChartsOption = {
            backgroundColor: 'transparent',
            grid: { top: 40, right: 30, bottom: 30, left: 50 },
            tooltip: {
              trigger: 'axis',
              backgroundColor: 'rgba(0,0,0,0.8)',
              borderColor: '#333',
              textStyle: { color: '#ffffff' },
            },
            legend: {
              data: config.series.map((s) => s.name),
              textStyle: { color: '#aaaaaa' },
              top: 0,
              left: 0,
            },
            xAxis: {
              type: 'category',
              data: dsCpu.labels,
              boundaryGap: false,
              axisLine: { lineStyle: { color: '#333' } },
              axisLabel: {
                color: '#aaaaaa',
                fontSize: 11,
                // 根据数据量自动控制显示密度，约每页 12 个标签
                interval: Math.max(0, Math.floor(dsCpu.labels.length / 12 - 1)),
                formatter: formatLabel,
              },
              splitLine: { show: false },
            },
            yAxis: {
              type: 'value',
              min: 0,
              max: 100,
              axisLine: { lineStyle: { color: '#333' } },
              axisLabel: { color: '#aaaaaa', fontSize: 11, formatter: '{value}%' },
              splitLine: { lineStyle: { color: '#333', type: 'dashed' } },
            },
            series: config.series.map((s) => {
              const serieData =
                s.dataKey === 'cpuData' ? dsCpu.data : s.dataKey === 'memoryData' ? dsMemory.data : [];
              return {
                name: s.name,
                type: (config.chartType === 'area' ? 'line' : config.chartType) as 'line' | 'bar',
                data: serieData,
                smooth: true,
                symbol: 'none',
                sampling: 'lttb',
                lineStyle: { color: s.color, width: 2 },
                itemStyle: { color: s.color },
                areaStyle:
                  config.chartType === 'area'
                    ? { opacity: 0.1, color: s.color }
                    : undefined,
              };
            }),
          };

          chartInstance.current?.setOption(option, { notMerge: true });
        })
        .catch((err) => {
          console.error('[ChartPanel] API 请求失败:', err);
        })
        .finally(() => {
          setLoading(false);
        });
    };

    fetchAndRender();

    // 自动刷新：每 15 秒重新拉取数据
    const interval = setInterval(fetchAndRender, 15000);

    const handleResize = () => chartInstance.current?.resize();
    window.addEventListener('resize', handleResize);

    return () => {
      clearInterval(interval);
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
      <div style={{ height: config.height, position: 'relative' }}>
        {/* echarts 独占一个容器，React 不管理其内部 DOM */}
        <div ref={chartRef} style={{ height: '100%', width: '100%' }} />

        {/* loading 遮罩层作为兄弟节点，避免 React DOM diff 冲突 */}
        {loading && (
          <div
            style={{
              position: 'absolute',
              inset: 0,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              background: '#1f1f1f',
              zIndex: 1,
            }}
          >
            <Spin />
          </div>
        )}
      </div>
    </Card>
  );
}
