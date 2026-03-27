"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { Hero, BuildCoverageEntry } from "@/lib/types";
import { getHeroes, getBuildCoverage } from "@/lib/api";

export default function BuildsPage() {
  const [heroes, setHeroes] = useState<Hero[]>([]);
  const [coverageMap, setCoverageMap] = useState<Record<number, BuildCoverageEntry>>({});
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const [h, c] = await Promise.all([getHeroes(), getBuildCoverage()]);
        setHeroes(h);
        const map: Record<number, BuildCoverageEntry> = {};
        for (const entry of c) {
          map[entry.hero_id] = entry;
        }
        setCoverageMap(map);
      } catch {
        // API not ready
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  const filtered = heroes.filter((h) =>
    h.name.toLowerCase().includes(search.toLowerCase())
  );

  if (loading) {
    return <div className="text-[var(--muted)]">Loading...</div>;
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-2">Build Win Rates</h1>
      <p className="text-[var(--muted)] mb-6 max-w-2xl">
        End-game item builds per hero, clustered by item similarity.
        Coverage shows what percentage of games fall into a sufficiently large cluster.
      </p>
      <input
        type="text"
        placeholder="Search heroes..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="w-full max-w-md mb-6 px-4 py-2 rounded-lg bg-[var(--card)] border border-[var(--card-border)] text-[var(--foreground)] placeholder-[var(--muted)] focus:outline-none focus:border-[var(--accent)]"
      />
      <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8 gap-3">
        {filtered.map((hero) => {
          const cov = coverageMap[hero.id];
          return (
            <Link
              key={hero.id}
              href={`/builds/${hero.id}`}
              className="flex flex-col items-center p-3 rounded-lg bg-[var(--card)] border border-[var(--card-border)] hover:border-[var(--accent)] transition-colors relative"
            >
              {hero.image_url ? (
                <img
                  src={hero.image_url}
                  alt={hero.name}
                  className="w-12 h-12 rounded-full mb-2"
                />
              ) : (
                <div className="w-12 h-12 rounded-full mb-2 bg-[var(--card-border)] flex items-center justify-center text-xs text-[var(--muted)]">
                  {hero.name[0]}
                </div>
              )}
              <span className="text-xs text-center truncate w-full">
                {hero.name}
              </span>
              {cov && (
                <span className={`text-[10px] mt-1 font-mono ${
                  cov.coverage >= 0.5 ? "text-[var(--positive)]" :
                  cov.coverage >= 0.3 ? "text-yellow-400" :
                  "text-[var(--muted)]"
                }`}>
                  {(cov.coverage * 100).toFixed(0)}% cov
                </span>
              )}
            </Link>
          );
        })}
      </div>
    </div>
  );
}
