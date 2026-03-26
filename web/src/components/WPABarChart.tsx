"use client";

import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  Cell,
  ReferenceLine,
} from "recharts";
import { WPAResult, Item } from "@/lib/types";

interface Props {
  results: WPAResult[];
  itemMap: Record<number, Item>;
  limit?: number;
}

export default function WPABarChart({
  results,
  itemMap,
  limit = 20,
}: Props) {
  const sorted = [...results]
    .sort((a, b) => b.mean_delta_w - a.mean_delta_w)
    .slice(0, limit);

  const data = sorted.map((r) => ({
    name: itemMap[r.item_id]?.name || `Item ${r.item_id}`,
    deltaW: +(r.mean_delta_w * 100).toFixed(2),
    positive: r.mean_delta_w >= 0,
  }));

  return (
    <ResponsiveContainer width="100%" height={Math.max(300, data.length * 28)}>
      <BarChart data={data} layout="vertical" margin={{ left: 120, right: 20 }}>
        <XAxis
          type="number"
          tick={{ fill: "#71717a", fontSize: 11 }}
          tickFormatter={(v) => `${v}%`}
        />
        <YAxis
          type="category"
          dataKey="name"
          tick={{ fill: "#e2e2e8", fontSize: 11 }}
          width={110}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: "#1a1a24",
            border: "1px solid #2a2a38",
            borderRadius: 8,
            color: "#e2e2e8",
          }}
          formatter={(value) => [`${value}%`, "ΔW̄"]}
        />
        <ReferenceLine x={0} stroke="#2a2a38" />
        <Bar dataKey="deltaW" radius={[0, 4, 4, 0]}>
          {data.map((entry, index) => (
            <Cell
              key={index}
              fill={entry.positive ? "#22c55e" : "#ef4444"}
              fillOpacity={0.8}
            />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}
