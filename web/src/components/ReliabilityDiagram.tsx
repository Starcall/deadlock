"use client";

import {
  ScatterChart,
  Scatter,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from "recharts";

interface ReliabilityBin {
  bin_start: number;
  bin_end: number;
  mean_predicted: number;
  mean_actual: number;
  count: number;
}

interface Props {
  bins: ReliabilityBin[];
}

export default function ReliabilityDiagram({ bins }: Props) {
  const data = bins
    .filter((b) => b.count > 0)
    .map((b) => ({
      predicted: +(b.mean_predicted * 100).toFixed(1),
      actual: +(b.mean_actual * 100).toFixed(1),
      count: b.count,
    }));

  if (data.length === 0) {
    return (
      <p className="text-sm text-[var(--muted)]">
        No reliability data available.
      </p>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={300}>
      <ScatterChart margin={{ top: 10, right: 20, bottom: 20, left: 20 }}>
        <XAxis
          type="number"
          dataKey="predicted"
          domain={[0, 100]}
          tick={{ fill: "#71717a", fontSize: 11 }}
          label={{
            value: "Predicted (%)",
            position: "bottom",
            fill: "#71717a",
            fontSize: 11,
          }}
        />
        <YAxis
          type="number"
          dataKey="actual"
          domain={[0, 100]}
          tick={{ fill: "#71717a", fontSize: 11 }}
          label={{
            value: "Actual (%)",
            angle: -90,
            position: "left",
            fill: "#71717a",
            fontSize: 11,
          }}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: "#1a1a24",
            border: "1px solid #2a2a38",
            borderRadius: 8,
            color: "#e2e2e8",
          }}
          formatter={(value, name) => [
            `${value}%`,
            name === "predicted" ? "Predicted" : "Actual",
          ]}
        />
        <ReferenceLine
          segment={[
            { x: 0, y: 0 },
            { x: 100, y: 100 },
          ]}
          stroke="#2a2a38"
          strokeDasharray="3 3"
        />
        <Scatter data={data} fill="#6366f1" />
      </ScatterChart>
    </ResponsiveContainer>
  );
}
