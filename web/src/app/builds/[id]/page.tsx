"use client";

import { useParams } from "next/navigation";
import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { Hero, Item, HeroBuildData } from "@/lib/types";
import { getHeroes, getItems, getHeroBuilds } from "@/lib/api";
import BuildTable from "@/components/BuildTable";

export default function HeroBuildPage() {
  const params = useParams();
  const heroId = Number(params.id);

  const [hero, setHero] = useState<Hero | null>(null);
  const [buildData, setBuildData] = useState<HeroBuildData | null>(null);
  const [itemMap, setItemMap] = useState<Record<number, Item>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [builds, items, heroes] = await Promise.all([
        getHeroBuilds(heroId),
        getItems(),
        getHeroes(),
      ]);
      setBuildData(builds);
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
  }, [heroId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return <div className="text-[var(--muted)]">Loading...</div>;
  }

  if (error) {
    return <div className="text-[var(--negative)]">{error}</div>;
  }

  return (
    <div>
      <Link
        href="/builds"
        className="text-sm text-[var(--muted)] hover:text-[var(--foreground)] mb-4 inline-block"
      >
        &larr; All Heroes
      </Link>

      <div className="flex items-center gap-4 mb-6">
        {hero?.image_url ? (
          <img
            src={hero.image_url}
            alt={hero.name}
            className="w-16 h-16 rounded-full"
          />
        ) : (
          <div className="w-16 h-16 rounded-full bg-[var(--card-border)]" />
        )}
        <div>
          <h1 className="text-2xl font-bold">{hero?.name || `Hero ${heroId}`}</h1>
          <p className="text-[var(--muted)] text-sm">Build Win Rates</p>
        </div>
      </div>

      {buildData && buildData.total_players > 0 && (
        <div className="flex gap-6 mb-6 p-4 rounded-lg bg-[var(--card)] border border-[var(--card-border)]">
          <div>
            <div className="text-2xl font-bold font-mono">
              {(buildData.coverage * 100).toFixed(1)}%
            </div>
            <div className="text-xs text-[var(--muted)]">Coverage</div>
          </div>
          <div>
            <div className="text-2xl font-bold font-mono">
              {buildData.total_players.toLocaleString()}
            </div>
            <div className="text-xs text-[var(--muted)]">Total Games</div>
          </div>
          <div>
            <div className="text-2xl font-bold font-mono">
              {buildData.builds.length}
            </div>
            <div className="text-xs text-[var(--muted)]">Build Templates</div>
          </div>
        </div>
      )}

      <BuildTable
        builds={buildData?.builds || []}
        itemMap={itemMap}
        totalPlayers={buildData?.total_players || 0}
      />
    </div>
  );
}
