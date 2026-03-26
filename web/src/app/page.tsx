import { getHeroes, getStatus } from "@/lib/api";
import HeroSelector from "@/components/HeroSelector";

export default async function HomePage() {
  let heroes: Awaited<ReturnType<typeof getHeroes>> = [];
  let status: Awaited<ReturnType<typeof getStatus>> | null = null;

  try {
    [heroes, status] = await Promise.all([getHeroes(), getStatus()]);
  } catch {
    // API not available - show empty state
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold mb-2">Deadlock WPA Analytics</h1>
        <p className="text-[var(--muted)]">
          Win Probability Added — debiased item effectiveness metrics for every
          hero. Select a hero to see which items genuinely increase win
          probability.
        </p>
        {status && (
          <div className="mt-3 flex gap-4 text-xs text-[var(--muted)]">
            <span>{status.match_count.toLocaleString()} matches analyzed</span>
            {status.model_accuracy > 0 && (
              <span>
                Model accuracy: {(status.model_accuracy * 100).toFixed(1)}%
              </span>
            )}
            {status.latest_match && (
              <span>
                Latest:{" "}
                {new Date(status.latest_match).toLocaleDateString()}
              </span>
            )}
          </div>
        )}
      </div>

      {heroes.length > 0 ? (
        <HeroSelector heroes={heroes} />
      ) : (
        <div className="text-center py-16">
          <p className="text-[var(--muted)] text-lg mb-2">
            No hero data available
          </p>
          <p className="text-[var(--muted)] text-sm">
            Start the API server and run the ingest pipeline to populate data.
          </p>
          <pre className="mt-4 text-xs bg-[var(--card)] p-4 rounded-lg inline-block text-left">
            {`go run ./cmd/ingest --count=1000
go run ./cmd/compute
go run ./cmd/server`}
          </pre>
        </div>
      )}
    </div>
  );
}
