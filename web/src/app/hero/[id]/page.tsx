"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams } from "next/navigation";
import { getHeroWPA, getItems, getHeroes } from "@/lib/api";
import { WPAResult, Item, Hero } from "@/lib/types";
import ItemWPATable from "@/components/ItemWPATable";
import WPABarChart from "@/components/WPABarChart";
import ContextFilters from "@/components/ContextFilters";

export default function HeroDashboard() {
  const params = useParams();
  const heroId = Number(params.id);

  const [hero, setHero] = useState<Hero | null>(null);
  const [results, setResults] = useState<WPAResult[]>([]);
  const [itemMap, setItemMap] = useState<Record<number, Item>>({});
  const [rank, setRank] = useState("all");
  const [time, setTime] = useState("all");
  const [minSampleSize, setMinSampleSize] = useState(30);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [wpa, items, heroes] = await Promise.all([
        getHeroWPA(heroId, rank, time, minSampleSize),
        getItems(),
        getHeroes(),
      ]);
      setResults(wpa);
      const map: Record<number, Item> = {};
      for (const item of items) {
        map[item.id] = item;
      }
      setItemMap(map);
      setHero(heroes.find((h) => h.id === heroId) || null);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, [heroId, rank, time, minSampleSize]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return (
    <div>
      <div className="mb-6">
        <div className="flex items-center gap-4 mb-4">
          {hero?.image_url && (
            <img
              src={hero.image_url}
              alt={hero.name}
              className="w-16 h-16 rounded-full"
            />
          )}
          <div>
            <h1 className="text-2xl font-bold">
              {hero?.name || `Hero ${heroId}`}
            </h1>
            <p className="text-[var(--muted)] text-sm">
              Item Win Probability Added (WPA) analysis
            </p>
          </div>
        </div>

        <ContextFilters
          rank={rank}
          onRankChange={setRank}
          time={time}
          onTimeChange={setTime}
          minSampleSize={minSampleSize}
          onMinSampleSizeChange={setMinSampleSize}
        />
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-16">
          <div className="animate-spin w-8 h-8 border-2 border-[var(--accent)] border-t-transparent rounded-full" />
        </div>
      ) : error ? (
        <div className="bg-red-900/20 border border-red-800 rounded-lg p-4 text-red-300">
          {error}
        </div>
      ) : (
        <div className="space-y-8">
          <section className="bg-[var(--card)] border border-[var(--card-border)] rounded-lg p-4">
            <h2 className="text-lg font-semibold mb-4">
              Top Items by Win Probability Added
            </h2>
            <WPABarChart results={results} itemMap={itemMap} limit={20} />
          </section>

          <section className="bg-[var(--card)] border border-[var(--card-border)] rounded-lg p-4">
            <h2 className="text-lg font-semibold mb-4">
              All Items — WPA Details
            </h2>
            <div className="text-xs text-[var(--muted)] mb-3 space-y-1">
              <p>
                <strong>ΔW̄</strong> = Mean WPA (change in win probability after
                purchase). Green = positive impact, Red = negative.
              </p>
              <p>
                <strong>W̄</strong> = Mean initial win probability when item is
                purchased. Yellow = bought while ahead, Blue = bought while
                behind.
              </p>
              <p>
                <strong>K</strong> = Sample size (number of observed purchases).
              </p>
              <p>
                <strong>P</strong> = P-value (t-test, H₀: ΔW̄ = 0). Green = p &lt; 0.01, Yellow = p &lt; 0.05, Gray = not significant.
              </p>
            </div>
            <ItemWPATable results={results} itemMap={itemMap} />
          </section>
        </div>
      )}
    </div>
  );
}
