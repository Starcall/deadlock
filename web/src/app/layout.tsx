import type { Metadata } from "next";
import localFont from "next/font/local";
import "./globals.css";
import Link from "next/link";

const geistSans = localFont({
  src: "./fonts/GeistVF.woff",
  variable: "--font-geist-sans",
  weight: "100 900",
});
const geistMono = localFont({
  src: "./fonts/GeistMonoVF.woff",
  variable: "--font-geist-mono",
  weight: "100 900",
});

export const metadata: Metadata = {
  title: "Deadlock WPA Analytics",
  description:
    "Win Probability Added analytics for Deadlock — debiased item effectiveness metrics",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased min-h-screen`}
      >
        <nav className="border-b border-[var(--card-border)] bg-[var(--card)]">
          <div className="max-w-7xl mx-auto px-4 py-3 flex items-center gap-6">
            <Link
              href="/"
              className="text-lg font-bold text-[var(--accent)]"
            >
              Deadlock WPA
            </Link>
            <Link
              href="/"
              className="text-sm text-[var(--muted)] hover:text-[var(--foreground)]"
            >
              Heroes
            </Link>
            <Link
              href="/builds"
              className="text-sm text-[var(--muted)] hover:text-[var(--foreground)]"
            >
              Builds
            </Link>
            <Link
              href="/about"
              className="text-sm text-[var(--muted)] hover:text-[var(--foreground)]"
            >
              Methodology
            </Link>
          </div>
        </nav>
        <main className="max-w-7xl mx-auto px-4 py-6">{children}</main>
      </body>
    </html>
  );
}
