"use client";

import { useState } from "react";
import { WPAResult, Item } from "@/lib/types";

type SortField = "delta_w" | "initial_w" | "win_rate" | "sample_size";

interface Props {
  results: WPAResult[];
  itemMap: Record<number, Item>;
}

export default function ItemWPATable({ results, itemMap }: Props) {
  const [sortField, setSortField] = useState<SortField>("delta_w");
  const [sortAsc, setSortAsc] = useState(false);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortAsc(!sortAsc);
    } else {
      setSortField(field);
      setSortAsc(false);
    }
  };

  const sorted = [...results].sort((a, b) => {
    let va: number, vb: number;
    switch (sortField) {
      case "delta_w":
        va = a.mean_delta_w;
        vb = b.mean_delta_w;
        break;
      case "initial_w":
        va = a.mean_initial_w;
        vb = b.mean_initial_w;
        break;
      case "win_rate":
        va = a.win_rate;
        vb = b.win_rate;
        break;
      case "sample_size":
        va = a.sample_size;
        vb = b.sample_size;
        break;
    }
    return sortAsc ? va - vb : vb - va;
  });

  const arrow = (field: SortField) =>
    sortField === field ? (sortAsc ? " ↑" : " ↓") : "";

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-[var(--card-border)] text-[var(--muted)] text-left">
            <th className="py-2 px-3">Item</th>
            <th className="py-2 px-3">Category</th>
            <th className="py-2 px-3">Tier</th>
            <th
              className="py-2 px-3 cursor-pointer hover:text-[var(--foreground)]"
              onClick={() => handleSort("delta_w")}
            >
              ΔW̄ (%){arrow("delta_w")}
            </th>
            <th
              className="py-2 px-3 cursor-pointer hover:text-[var(--foreground)]"
              onClick={() => handleSort("initial_w")}
            >
              W̄ (%){arrow("initial_w")}
            </th>
            <th
              className="py-2 px-3 cursor-pointer hover:text-[var(--foreground)]"
              onClick={() => handleSort("win_rate")}
            >
              Win Rate (%){arrow("win_rate")}
            </th>
            <th
              className="py-2 px-3 cursor-pointer hover:text-[var(--foreground)]"
              onClick={() => handleSort("sample_size")}
            >
              K{arrow("sample_size")}
            </th>
          </tr>
        </thead>
        <tbody>
          {sorted.map((r) => {
            const item = itemMap[r.item_id];
            const deltaClass =
              r.mean_delta_w > 0
                ? "text-[var(--positive)]"
                : r.mean_delta_w < 0
                ? "text-[var(--negative)]"
                : "";
            const biasClass =
              r.mean_initial_w > 0.55
                ? "text-yellow-400"
                : r.mean_initial_w < 0.45
                ? "text-blue-400"
                : "";

            return (
              <tr
                key={r.item_id}
                className="border-b border-[var(--card-border)] hover:bg-[var(--card)]"
              >
                <td className="py-2 px-3 flex items-center gap-2">
                  {item?.image_url && (
                    <img
                      src={item.image_url}
                      alt=""
                      className="w-6 h-6 rounded"
                    />
                  )}
                  <span>{item?.name || <span className="text-[var(--muted)]">Unknown Item #{r.item_id}</span>}</span>
                </td>
                <td className="py-2 px-3 capitalize text-[var(--muted)]">
                  {item?.item_slot_type || "-"}
                </td>
                <td className="py-2 px-3 text-[var(--muted)]">
                  T{item?.item_tier || "?"}
                </td>
                <td className={`py-2 px-3 font-mono font-bold ${deltaClass}`}>
                  {(r.mean_delta_w * 100).toFixed(2)}%
                </td>
                <td className={`py-2 px-3 font-mono ${biasClass}`}>
                  {(r.mean_initial_w * 100).toFixed(1)}%
                </td>
                <td className="py-2 px-3 font-mono">
                  {(r.win_rate * 100).toFixed(1)}%
                </td>
                <td className="py-2 px-3 font-mono text-[var(--muted)]">
                  {r.sample_size.toLocaleString()}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
      {sorted.length === 0 && (
        <p className="text-center text-[var(--muted)] py-8">
          No WPA data available for this hero/context. Try lowering the minimum
          sample size or selecting a different context.
        </p>
      )}
    </div>
  );
}
