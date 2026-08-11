export interface SeriesPoint {
  label: string;
  value: number;
}

export function lineChartOption(
  points: SeriesPoint[],
  color: string,
  yName?: string,
) {
  return {
    grid: { left: 48, right: 12, top: 24, bottom: 28 },
    tooltip: { trigger: "axis" as const },
    xAxis: {
      type: "category" as const,
      data: points.map((p) => p.label),
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
      axisLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
    },
    yAxis: {
      type: "value" as const,
      name: yName,
      splitLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
    },
    series: [
      {
        type: "line" as const,
        smooth: true,
        data: points.map((p) => p.value),
        areaStyle: { opacity: 0.12 },
        lineStyle: { width: 2, color },
        itemStyle: { color },
        showSymbol: false,
      },
    ],
  };
}

export function barChartOption(
  points: SeriesPoint[],
  color: string,
  yName?: string,
) {
  return {
    grid: { left: 48, right: 12, top: 24, bottom: 28 },
    tooltip: { trigger: "axis" as const },
    xAxis: {
      type: "category" as const,
      data: points.map((p) => p.label),
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
      axisLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
    },
    yAxis: {
      type: "value" as const,
      name: yName,
      splitLine: { lineStyle: { color: "var(--nx-border-subtle)" } },
      axisLabel: { color: "var(--nx-text-tertiary)", fontSize: 11 },
    },
    series: [
      {
        type: "bar" as const,
        data: points.map((p) => p.value),
        itemStyle: { color, borderRadius: [2, 2, 0, 0] },
        barMaxWidth: 28,
      },
    ],
  };
}
