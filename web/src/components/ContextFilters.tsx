"use client";

interface Props {
  rank: string;
  onRankChange: (rank: string) => void;
  time: string;
  onTimeChange: (time: string) => void;
  minSampleSize: number;
  onMinSampleSizeChange: (n: number) => void;
}

const RANKS = [
  { key: "all", label: "All Ranks" },
  { key: "rank:low", label: "Low" },
  { key: "rank:mid", label: "Mid" },
  { key: "rank:high", label: "High" },
];

const TIMES = [
  { key: "all", label: "All Time" },
  { key: "before:5m", label: "< 5m" },
  { key: "before:8m", label: "< 8m" },
  { key: "before:10m", label: "< 10m" },
  { key: "before:15m", label: "< 15m" },
  { key: "before:20m", label: "< 20m" },
  { key: "phase:early", label: "Early" },
  { key: "phase:mid", label: "Mid" },
  { key: "phase:late", label: "Late" },
];

export default function ContextFilters({
  rank,
  onRankChange,
  time,
  onTimeChange,
  minSampleSize,
  onMinSampleSizeChange,
}: Props) {
  const btn = (active: boolean) =>
    `px-3 py-1.5 text-xs rounded-full border transition-colors ${
      active
        ? "bg-[var(--accent)] border-[var(--accent)] text-white"
        : "bg-[var(--card)] border-[var(--card-border)] text-[var(--muted)] hover:border-[var(--accent)]"
    }`;

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-4">
        <span className="text-xs text-[var(--muted)] w-12">Rank</span>
        <div className="flex flex-wrap gap-2">
          {RANKS.map((r) => (
            <button key={r.key} onClick={() => onRankChange(r.key)} className={btn(rank === r.key)}>
              {r.label}
            </button>
          ))}
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-4">
        <span className="text-xs text-[var(--muted)] w-12">Time</span>
        <div className="flex flex-wrap gap-2">
          {TIMES.map((t) => (
            <button key={t.key} onClick={() => onTimeChange(t.key)} className={btn(time === t.key)}>
              {t.label}
            </button>
          ))}
        </div>
      </div>
      <div className="flex items-center gap-2 text-xs text-[var(--muted)]">
        <label>Min samples:</label>
        <input
          type="number"
          min={1}
          value={minSampleSize}
          onChange={(e) => onMinSampleSizeChange(Number(e.target.value) || 1)}
          className="w-20 px-2 py-1 rounded bg-[var(--card)] border border-[var(--card-border)] text-[var(--foreground)] focus:outline-none focus:border-[var(--accent)]"
        />
      </div>
    </div>
  );
}
