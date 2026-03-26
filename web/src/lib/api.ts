import { Hero, Item, WPAResult, ModelStats, StatusInfo } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function fetchJSON<T>(path: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, { next: { revalidate: 60 } });
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`);
  }
  return res.json();
}

export async function getHeroes(): Promise<Hero[]> {
  return fetchJSON<Hero[]>("/api/heroes");
}

export async function getItems(): Promise<Item[]> {
  return fetchJSON<Item[]>("/api/items");
}

export async function getHeroWPA(
  heroId: number,
  context: string = "all",
  minSampleSize: number = 30
): Promise<WPAResult[]> {
  return fetchJSON<WPAResult[]>(
    `/api/wpa/hero/${heroId}?context=${context}&min_sample_size=${minSampleSize}`
  );
}

export async function getHeroItemWPA(
  heroId: number,
  itemId: number
): Promise<WPAResult[]> {
  return fetchJSON<WPAResult[]>(`/api/wpa/hero/${heroId}/item/${itemId}`);
}

export async function getModelStats(): Promise<ModelStats> {
  return fetchJSON<ModelStats>("/api/model/stats");
}

export async function getStatus(): Promise<StatusInfo> {
  return fetchJSON<StatusInfo>("/api/status");
}
