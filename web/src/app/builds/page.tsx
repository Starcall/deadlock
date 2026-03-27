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
      <div className="mb-6 max-w-2xl space-y-3 text-sm text-[var(--muted)]">
        <p>
          End-game item builds per hero, discovered automatically using K-means clustering.
        </p>
        <details className="cursor-pointer">
          <summary className="text-[var(--foreground)] font-medium hover:text-[var(--accent)]">
            How it works
          </summary>
          <div className="mt-2 space-y-2 pl-2 border-l-2 border-[var(--card-border)]">
            <p>
              For each hero, we collect every player&apos;s final item set (items not sold by end of game)
              across all matches in the current patch window.
            </p>
            <p>
              Each player&apos;s build is converted into a binary vector &mdash; 1 if they have a given item, 0 otherwise.
              We then run <strong className="text-[var(--foreground)]">K-means clustering</strong> on these vectors
              to group players with similar item choices together.
            </p>
            <p>
              The number of clusters (3&ndash;12) is chosen automatically using the{" "}
              <strong className="text-[var(--foreground)]">silhouette score</strong>, which measures
              how well each player fits their assigned cluster vs. the next closest one. A higher score
              means more distinct, well-separated build archetypes.
            </p>
            <p>
              Each cluster&apos;s <strong className="text-[var(--foreground)]">template</strong> is the set
              of items that appear in at least 30% of that cluster&apos;s players (up to 8 items).
              These are the core items that define the build.
            </p>
            <p>
              <strong className="text-[var(--foreground)]">Win rate</strong> is computed from all players
              assigned to the cluster, not just those with an exact item match.
              Clusters with fewer than 30 games are excluded.
            </p>
            <p>
              <strong className="text-[var(--foreground)]">Coverage</strong> shows what percentage of games
              fall into a cluster large enough to report. Since every player is assigned to exactly one cluster,
              coverage is typically high (80&ndash;100%).
            </p>
          </div>
        </details>
      </div>
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
