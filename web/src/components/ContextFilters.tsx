"use client";

interface Props {
  context: string;
  onContextChange: (ctx: string) => void;
  minSampleSize: number;
  onMinSampleSizeChange: (n: number) => void;
}

const RANK_CONTEXTS = [
  { key: "all", label: "All Ranks" },
  { key: "rank:low", label: "Low Rank" },
  { key: "rank:mid", label: "Mid Rank" },
  { key: "rank:high", label: "High Rank" },
];

const TIME_CONTEXTS = [
  { key: "before:5m", label: "< 5 min" },
  { key: "before:8m", label: "< 8 min" },
  { key: "before:10m", label: "< 10 min" },
  { key: "before:15m", label: "< 15 min" },
  { key: "before:20m", label: "< 20 min" },
  { key: "phase:early", label: "Early (0-10m)" },
  { key: "phase:mid", label: "Mid (10-25m)" },
  { key: "phase:late", label: "Late (25m+)" },
];

export default function ContextFilters({
  context,
  onContextChange,
  minSampleSize,
  onMinSampleSizeChange,
}: Props) {
  const renderButtons = (items: { key: string; label: string }[]) =>
    items.map((c) => (
      <button
        key={c.key}
        onClick={() => onContextChange(c.key)}
        className={`px-3 py-1.5 text-xs rounded-full border transition-colors ${
          context === c.key
            ? "bg-[var(--accent)] border-[var(--accent)] text-white"
            : "bg-[var(--card)] border-[var(--card-border)] text-[var(--muted)] hover:border-[var(--accent)]"
        }`}
      >
        {c.label}
      </button>
    ));

  return (
    <div className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center gap-4">
        <span className="text-xs text-[var(--muted)] w-12">Rank</span>
        <div className="flex flex-wrap gap-2">{renderButtons(RANK_CONTEXTS)}</div>
      </div>
      <div className="flex flex-wrap items-center gap-4">
        <span className="text-xs text-[var(--muted)] w-12">Time</span>
        <div className="flex flex-wrap gap-2">{renderButtons(TIME_CONTEXTS)}</div>
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
