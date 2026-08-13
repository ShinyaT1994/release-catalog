import { Outlet, Link } from "react-router-dom";

export default function Layout() {
  return (
    <div style={{ minHeight: "100vh", display: "flex", flexDirection: "column" }}>
      <header style={{ background: "#1a56db", color: "#fff", padding: "12px 24px", display: "flex", alignItems: "center", gap: 16 }}>
        <Link to="/products" style={{ color: "#fff", textDecoration: "none", fontWeight: 700, fontSize: 18 }}>
          Release Catalog
        </Link>
        <span style={{ fontSize: 12, opacity: 0.8 }}>Dependency-Track Integration</span>
      </header>
      <main style={{ flex: 1, padding: 24, maxWidth: 1200, margin: "0 auto", width: "100%" }}>
        <Outlet />
      </main>
    </div>
  );
}
