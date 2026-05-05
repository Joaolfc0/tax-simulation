import { Link, Outlet } from "react-router-dom";

export function Layout() {
  return (
    <div className="min-h-full bg-white text-ink">
      <header className="sticky top-0 z-10 border-b border-line bg-white/90 backdrop-blur">
        <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
          <Link
            to="/companies"
            className="flex items-center gap-3 focus-ring rounded-sm"
          >
            <img
              src="/logo.webp"
              alt="Tax Simulation"
              className="h-8 w-8 rounded object-contain"
            />
            <span className="text-sm font-semibold tracking-tight text-ink">
              Tax Simulation
            </span>
          </Link>
          <div className="text-xs text-ink-faint">v0.1</div>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-10">
        <Outlet />
      </main>
    </div>
  );
}
