import { useQuery } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { api } from "../api/client";
import StatusBadge from "../components/StatusBadge";
import GraphView from "../components/GraphView";

export default function BranchDetail() {
  const { branchId } = useParams<{ branchId: string }>();
  const { data: branch, isLoading } = useQuery({ queryKey: ["branch", branchId], queryFn: () => api.getBranch(branchId!) });
  const { data: currentState } = useQuery({ queryKey: ["currentState", branchId], queryFn: () => api.getCurrentState(branchId!) });
  const { data: releases } = useQuery({ queryKey: ["releases", branchId], queryFn: () => api.listReleases(branchId!) });
  const { data: graph } = useQuery({
    queryKey: ["graph", branchId],
    queryFn: () => api.getBranchGraph(branchId!),
    enabled: !!currentState?.rootDtProjectUuid,
  });

  if (isLoading || !branch) return <p>Loading...</p>;

  return (
    <div>
      <nav style={{ fontSize: 14, color: "#6b7280", marginBottom: 16 }}>
        <Link to="/products" style={{ color: "#1a56db" }}>Products</Link>
        {" / "}
        <Link to={`/products/${branch.productId}`} style={{ color: "#1a56db" }}>Product</Link>
        {" / "}
        {branch.displayName || branch.name}
      </nav>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
        <h1 style={{ fontSize: 24, fontWeight: 700 }}>{branch.displayName || branch.name}</h1>
        <StatusBadge status={branch.status} />
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16, marginBottom: 24 }}>
        <InfoCard label="Type" value={branch.type} />
        <InfoCard label="Name" value={branch.name} />
        <InfoCard label="Status" value={branch.status} />
        <InfoCard label="Created" value={new Date(branch.createdAt).toLocaleDateString()} />
      </div>

      {/* Current State */}
      <section style={{ marginBottom: 32 }}>
        <h2 style={{ fontSize: 18, fontWeight: 600, marginBottom: 12 }}>Current State</h2>
        <div style={cardStyle}>
          {currentState?.rootDtProjectUuid ? (
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 8, fontSize: 14 }}>
              <div><strong>Root DT Project:</strong> <code>{currentState.rootDtProjectUuid}</code></div>
              <div><strong>BOM Serial:</strong> <code>{currentState.rootBomSerialNumber || "—"}</code></div>
              <div><strong>BOM Version:</strong> {currentState.rootBomVersion ?? "—"}</div>
              <div><strong>SHA-256:</strong> <code style={{ fontSize: 11 }}>{currentState.rootBomSha256 ? currentState.rootBomSha256.slice(0, 16) + "..." : "—"}</code></div>
              <div><strong>Source Revision:</strong> <code>{currentState.sourceRevision || "—"}</code></div>
              <div><strong>Updated:</strong> {new Date(currentState.updatedAt).toLocaleString()}</div>
            </div>
          ) : (
            <p style={{ color: "#6b7280" }}>No root project configured.</p>
          )}
        </div>
      </section>

      {/* Release Graph */}
      {graph && graph.nodes.length > 0 && (
        <section style={{ marginBottom: 32 }}>
          <h2 style={{ fontSize: 18, fontWeight: 600, marginBottom: 12 }}>Release Graph</h2>
          <GraphView graph={graph} />
        </section>
      )}

      {/* Release History */}
      {branch.type === "RELEASE" && (
        <section>
          <h2 style={{ fontSize: 18, fontWeight: 600, marginBottom: 12 }}>Release History</h2>
          {(!releases || releases.length === 0) ? (
            <p style={{ color: "#6b7280" }}>No releases yet.</p>
          ) : (
            <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 14 }}>
              <thead>
                <tr style={{ borderBottom: "2px solid #e5e7eb", textAlign: "left" }}>
                  <th style={{ padding: 8 }}>Version</th>
                  <th style={{ padding: 8 }}>Status</th>
                  <th style={{ padding: 8 }}>Released</th>
                  <th style={{ padding: 8 }}>Root Project</th>
                </tr>
              </thead>
              <tbody>
                {releases.map(r => (
                  <tr key={r.id} style={{ borderBottom: "1px solid #f3f4f6" }}>
                    <td style={{ padding: 8 }}>
                      <Link to={`/releases/${r.id}`} style={{ color: "#1a56db", fontWeight: 600 }}>{r.version}</Link>
                    </td>
                    <td style={{ padding: 8 }}><StatusBadge status={r.status} /></td>
                    <td style={{ padding: 8 }}>{r.releasedAt ? new Date(r.releasedAt).toLocaleDateString() : "—"}</td>
                    <td style={{ padding: 8 }}><code style={{ fontSize: 11 }}>{r.rootDtProjectUuid?.slice(0, 8) || "—"}</code></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </section>
      )}
    </div>
  );
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div style={cardStyle}>
      <div style={{ fontSize: 12, color: "#6b7280" }}>{label}</div>
      <div style={{ fontSize: 14, fontWeight: 600 }}>{value}</div>
    </div>
  );
}

const cardStyle: React.CSSProperties = { background: "#fff", border: "1px solid #e5e7eb", borderRadius: 8, padding: 12 };
