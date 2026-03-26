"use client";

import Link from "next/link";
import { Hero } from "@/lib/types";
import { useState } from "react";

export default function HeroSelector({ heroes }: { heroes: Hero[] }) {
  const [search, setSearch] = useState("");

  const filtered = heroes.filter((h) =>
    h.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div>
      <input
        type="text"
        placeholder="Search heroes..."
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        className="w-full max-w-md mb-6 px-4 py-2 rounded-lg bg-[var(--card)] border border-[var(--card-border)] text-[var(--foreground)] placeholder-[var(--muted)] focus:outline-none focus:border-[var(--accent)]"
      />
      <div className="grid grid-cols-3 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8 gap-3">
        {filtered.map((hero) => (
          <Link
            key={hero.id}
            href={`/hero/${hero.id}`}
            className="flex flex-col items-center p-3 rounded-lg bg-[var(--card)] border border-[var(--card-border)] hover:border-[var(--accent)] transition-colors"
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
          </Link>
        ))}
      </div>
    </div>
  );
}
