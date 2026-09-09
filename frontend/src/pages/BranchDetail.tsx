import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, Link, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import StatusBadge from "../components/StatusBadge";
import TimelineView from "../components/TimelineView";

export default function BranchDetail() {
  const { branchId } = useParams<{ branchId: string }>();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { data: branch, isLoading } = useQuery({ queryKey: ["branch", branchId], queryFn: () => api.getBranch(branchId!) });
  const { data: versions } = useQuery({ queryKey: ["versions", branchId], queryFn: () => api.listVersions(branchId!) });
  const { data: timeline } = useQuery({ queryKey: ["branchTimeline", branchId], queryFn: () => api.getBranchTimeline(branchId!) });

  const [newVersionString, setNewVersionString] = useState("");
  const createVersion = useMutation({
    mutationFn: () => api.createVersion(branchId!, { versionString: newVersionString }),
    onSuccess: (v) => {
      queryClient.invalidateQueries({ queryKey: ["versions", branchId] });
      queryClient.invalidateQueries({ queryKey: ["branchTimeline", branchId] });
      setNewVersionString("");
      navigate(`/versions/${v.id}`);
    },
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

      {/* Timeline */}
      {timeline && timeline.entries.length > 0 && (
        <section style={{ marginBottom: 32 }}>
          <h2 style={{ fontSize: 18, fontWeight: 600, marginBottom: 12 }}>Timeline</h2>
          <TimelineView entries={timeline.entries} />
        </section>
      )}

      {/* Versions */}
      <section>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
          <h2 style={{ fontSize: 18, fontWeight: 600 }}>Versions</h2>
          <div style={{ display: "flex", gap: 8 }}>
            <input
              placeholder="Version string (e.g. 1.0)"
              value={newVersionString}
              onChange={e => setNewVersionString(e.target.value)}
              style={inputStyle}
            />
            <button onClick={() => createVersion.mutate()} style={btnStyle} disabled={!newVersionString || createVersion.isPending}>
              + New Version
            </button>
          </div>
        </div>

        {(!versions || versions.length === 0) ? (
          <p style={{ color: "#6b7280" }}>No versions yet. Preset one with a version string.</p>
        ) : (
          <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 14 }}>
            <thead>
              <tr style={{ borderBottom: "2px solid #e5e7eb", textAlign: "left" }}>
                <th style={{ padding: 8 }}>Version</th>
                <th style={{ padding: 8 }}>Status</th>
                <th style={{ padding: 8 }}>Release Date</th>
                <th style={{ padding: 8 }}>Customer</th>
                <th style={{ padding: 8 }}>Projects</th>
              </tr>
            </thead>
            <tbody>
              {versions.map(v => (
                <tr key={v.id} style={{ borderBottom: "1px solid #f3f4f6" }}>
                  <td style={{ padding: 8 }}>
                    <Link to={`/versions/${v.id}`} style={{ color: "#1a56db", fontWeight: 600 }}>{v.versionString}</Link>
                  </td>
                  <td style={{ padding: 8 }}><StatusBadge status={v.status} /></td>
                  <td style={{ padding: 8 }}>{v.releaseDate ? new Date(v.releaseDate).toLocaleDateString() : "—"}</td>
                  <td style={{ padding: 8 }}>{v.customer || "—"}</td>
                  <td style={{ padding: 8 }}>{v.projects?.length ?? 0}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
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
const btnStyle: React.CSSProperties = { background: "#1a56db", color: "#fff", border: "none", borderRadius: 6, padding: "8px 16px", cursor: "pointer", fontWeight: 600, fontSize: 14 };
const inputStyle: React.CSSProperties = { border: "1px solid #d1d5db", borderRadius: 6, padding: "8px 12px", fontSize: 14 };
