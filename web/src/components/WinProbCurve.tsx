"use client";

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from "recharts";

interface DataPoint {
  time_min: number;
  win_prob: number;
}

interface Props {
  data: DataPoint[];
}

export default function WinProbCurve({ data }: Props) {
  if (data.length === 0) {
    return (
      <p className="text-sm text-[var(--muted)]">
        No win probability curve data available.
      </p>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={250}>
      <LineChart data={data} margin={{ top: 10, right: 20, bottom: 20, left: 20 }}>
        <XAxis
          dataKey="time_min"
          tick={{ fill: "#71717a", fontSize: 11 }}
          label={{
            value: "Game Time (min)",
            position: "bottom",
            fill: "#71717a",
            fontSize: 11,
          }}
        />
        <YAxis
          domain={[0, 100]}
          tick={{ fill: "#71717a", fontSize: 11 }}
          tickFormatter={(v) => `${v}%`}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: "#1a1a24",
            border: "1px solid #2a2a38",
            borderRadius: 8,
            color: "#e2e2e8",
          }}
          formatter={(value) => [`${Number(value).toFixed(1)}%`, "Win Prob"]}
          labelFormatter={(label) => `${label} min`}
        />
        <ReferenceLine y={50} stroke="#2a2a38" strokeDasharray="3 3" />
        <Line
          type="monotone"
          dataKey="win_prob"
          stroke="#6366f1"
          strokeWidth={2}
          dot={false}
        />
      </LineChart>
    </ResponsiveContainer>
  );
}
