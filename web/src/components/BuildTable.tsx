"use client";

import { BuildTemplate, Item } from "@/lib/types";

interface Props {
  builds: BuildTemplate[];
  itemMap: Record<number, Item>;
  totalPlayers: number;
}

export default function BuildTable({ builds, itemMap, totalPlayers }: Props) {
  if (builds.length === 0) {
    return (
      <p className="text-[var(--muted)]">
        No build data available. Run compute first.
      </p>
    );
  }

  return (
    <div className="space-y-3">
      {builds.map((build) => {
        const winPct = (build.win_rate * 100).toFixed(1);
        const popularity = totalPlayers > 0
          ? ((build.fuzzy_count / totalPlayers) * 100).toFixed(1)
          : "0";

        return (
          <div
            key={build.build_rank}
            className="p-4 rounded-lg bg-[var(--card)] border border-[var(--card-border)]"
          >
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-3">
                <span className="text-lg font-bold text-[var(--muted)] w-8">
                  #{build.build_rank}
                </span>
                <div className="flex items-center gap-1.5 flex-wrap">
                  {build.item_ids.map((itemId, idx) => {
                    const item = itemMap[itemId];
                    return (
                      <div
                        key={`${itemId}-${idx}`}
                        className="relative group"
                      >
                        {item?.image_url ? (
                          <img
                            src={item.image_url}
                            alt={item?.name || `Item ${itemId}`}
                            className="w-8 h-8 rounded border border-[var(--card-border)]"
                          />
                        ) : (
                          <div className="w-8 h-8 rounded border border-[var(--card-border)] bg-[var(--card-border)] flex items-center justify-center text-[8px] text-[var(--muted)]">
                            {item?.name?.[0] || "?"}
                          </div>
                        )}
                        <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-1 px-2 py-1 rounded bg-black/90 text-[10px] text-white whitespace-nowrap opacity-0 group-hover:opacity-100 pointer-events-none z-10">
                          {item?.name || `Item ${itemId}`}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
              <div className="flex items-center gap-4 text-sm font-mono">
                <span className={build.win_rate >= 0.5 ? "text-[var(--positive)]" : "text-[var(--negative)]"}>
                  {winPct}% WR
                </span>
              </div>
            </div>
            <div className="flex items-center gap-4 text-xs text-[var(--muted)]">
              <span>{build.fuzzy_count} games ({popularity}% of total)</span>
              <span>{build.exact_count} exact matches</span>
              <span>{build.wins}W / {build.losses}L</span>
              {/* Popularity bar */}
              <div className="flex-1 h-1.5 bg-[var(--card-border)] rounded-full overflow-hidden max-w-32">
                <div
                  className="h-full bg-[var(--accent)] rounded-full"
                  style={{ width: `${Math.min(parseFloat(popularity) * 2, 100)}%` }}
                />
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
