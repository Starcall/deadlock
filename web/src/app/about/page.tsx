export default function AboutPage() {
  return (
    <div className="max-w-3xl">
      <h1 className="text-3xl font-bold mb-6">Methodology</h1>

      <div className="space-y-8 text-[var(--foreground)]">
        <section>
          <h2 className="text-xl font-semibold mb-3">
            Why Not Just Use Win Rate?
          </h2>
          <p className="text-[var(--muted)] leading-relaxed">
            Raw win rates for items are misleading because of{" "}
            <strong className="text-[var(--foreground)]">selection bias</strong>.
            Expensive late-game items have inflated win rates because they are
            primarily bought by players who are already winning. Conversely,
            comeback items may have low win rates despite being impactful because
            they are bought from behind. Win Probability Added (WPA) corrects
            for this by measuring the <em>change</em> in win probability, not
            the absolute level.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">Win Probability Model</h2>
          <p className="text-[var(--muted)] leading-relaxed mb-3">
            We train a logistic regression model to estimate{" "}
            <code className="bg-[var(--card)] px-1.5 py-0.5 rounded text-sm">
              P(win | game_state)
            </code>{" "}
            — the probability that a team wins given the current game state.
          </p>
          <p className="text-[var(--muted)] leading-relaxed mb-3">
            The model uses 67 features including:
          </p>
          <ul className="list-disc list-inside text-[var(--muted)] space-y-1 ml-4">
            <li>Team net worth difference</li>
            <li>Team kills/deaths/assists differences</li>
            <li>Team damage dealt difference</li>
            <li>Per-player net worth, kills, deaths, assists, and level (12 players x 5 features)</li>
            <li>Normalized game time</li>
            <li>Average rank badge</li>
          </ul>
          <p className="text-[var(--muted)] leading-relaxed mt-3">
            <strong className="text-[var(--foreground)]">
              Items are deliberately excluded
            </strong>{" "}
            from the feature set. Since items are the actions being evaluated,
            including them would create circular reasoning — the model would
            learn that &quot;having good items means winning&quot; rather than
            measuring the causal impact of buying an item.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">
            Win Probability Added (WPA)
          </h2>
          <p className="text-[var(--muted)] leading-relaxed mb-3">
            For each item purchase event, we compute:
          </p>
          <div className="bg-[var(--card)] border border-[var(--card-border)] rounded-lg p-4 font-mono text-sm mb-3">
            <p>
              W(before) = model prediction just before the purchase
            </p>
            <p>
              W(after) = model prediction shortly after the purchase
            </p>
            <p className="mt-2 text-[var(--accent)]">
              deltaW = W(after) - W(before)
            </p>
          </div>
          <p className="text-[var(--muted)] leading-relaxed">
            We then aggregate deltaW across all purchases of the same item by the
            same hero, segmented by rank bracket and game phase.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">Key Metrics</h2>
          <div className="space-y-4">
            <div className="bg-[var(--card)] border border-[var(--card-border)] rounded-lg p-4">
              <h3 className="font-semibold text-[var(--positive)]">
                Mean WPA (deltaW)
              </h3>
              <p className="text-[var(--muted)] text-sm mt-1">
                The average change in win probability after buying this item.
                Positive = the item tends to improve your chances. Negative =
                the item tends to coincide with declining game state.
              </p>
            </div>
            <div className="bg-[var(--card)] border border-[var(--card-border)] rounded-lg p-4">
              <h3 className="font-semibold text-yellow-400">
                Mean Initial Win Probability (W)
              </h3>
              <p className="text-[var(--muted)] text-sm mt-1">
                The average win probability when the item is purchased. This
                reveals selection bias — a high W means the item is mostly
                bought while ahead, a low W means it is bought from behind.
              </p>
            </div>
            <div className="bg-[var(--card)] border border-[var(--card-border)] rounded-lg p-4">
              <h3 className="font-semibold">Win Rate</h3>
              <p className="text-[var(--muted)] text-sm mt-1">
                The raw win rate of games where this item was purchased. Compare
                with W to see how biased the raw win rate is.
              </p>
            </div>
            <div className="bg-[var(--card)] border border-[var(--card-border)] rounded-lg p-4">
              <h3 className="font-semibold">K (Sample Size)</h3>
              <p className="text-[var(--muted)] text-sm mt-1">
                The number of purchase events observed. Larger K means more
                confidence in the estimate. Results with very small K should be
                interpreted cautiously.
              </p>
            </div>
          </div>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">Model Calibration</h2>
          <p className="text-[var(--muted)] leading-relaxed">
            The model uses Platt scaling for probability calibration. This
            ensures that when the model predicts a 70% win probability, the team
            actually wins approximately 70% of the time. We measure calibration
            using Expected Calibration Error (ECE) and reliability diagrams.
          </p>
        </section>

        <section>
          <h2 className="text-xl font-semibold mb-3">References</h2>
          <p className="text-[var(--muted)] leading-relaxed">
            This methodology is adapted from xPetu&apos;s thesis on win
            probability estimation in esports, which applies the WPA framework
            (originally from baseball analytics) to competitive gaming.
          </p>
        </section>
      </div>
    </div>
  );
}
