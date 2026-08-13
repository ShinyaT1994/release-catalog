import { useQuery } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { api } from "../api/client";
import StatusBadge from "../components/StatusBadge";
import GraphView from "../components/GraphView";

export default function ReleaseDetail() {
  const { releaseId } = useParams<{ releaseId: string }>();
  const { data: release, isLoading } = useQuery({ queryKey: ["release", releaseId], queryFn: () => api.getRelease(releaseId!) });
  const { data: graph } = useQuery({
    queryKey: ["releaseGraph", releaseId],
    queryFn: () => api.getReleaseGraph(releaseId!),
    enabled: !!release?.rootDtProjectUuid,
  });

  if (isLoading || !release) return <p>Loading...</p>;

  return (
    <div>
      <nav style={{ fontSize: 14, color: "#6b7280", marginBottom: 16 }}>
        <Link to="/products" style={{ color: "#1a56db" }}>Products</Link>
        {" / "}
        <Link to={`/branches/${release.branchLineId}`} style={{ color: "#1a56db" }}>Branch</Link>
        {" / "}
        Release {release.version}
      </nav>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16 }}>
        <h1 style={{ fontSize: 24, fontWeight: 700 }}>Release {release.version}</h1>
        <StatusBadge status={release.status} />
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))", gap: 12, marginBottom: 24 }}>
        <InfoCard label="Snapshot Type" value={release.snapshotType} />
        <InfoCard label="Created" value={new Date(release.createdAt).toLocaleString()} />
        <InfoCard label="Released" value={release.releasedAt ? new Date(release.releasedAt).toLocaleString() : "—"} />
        <InfoCard label="Root DT Project" value={release.rootDtProjectUuid || "—"} />
        <InfoCard label="BOM Serial" value={release.rootBomSerialNumber || "—"} />
        <InfoCard label="BOM Version" value={String(release.rootBomVersion ?? "—")} />
        <InfoCard label="SHA-256" value={release.rootBomSha256 ? release.rootBomSha256.slice(0, 16) + "..." : "—"} />
        <InfoCard label="Source Revision" value={release.sourceRevision || "—"} />
      </div>

      {graph && graph.nodes.length > 0 && (
        <section>
          <h2 style={{ fontSize: 18, fontWeight: 600, marginBottom: 12 }}>Release Graph</h2>
          <GraphView graph={graph} />
        </section>
      )}
    </div>
  );
}

function InfoCard({ label, value }: { label: string; value: string }) {
  return (
    <div style={{ background: "#fff", border: "1px solid #e5e7eb", borderRadius: 8, padding: 12 }}>
      <div style={{ fontSize: 12, color: "#6b7280" }}>{label}</div>
      <div style={{ fontSize: 13, fontWeight: 600, wordBreak: "break-all" }}>{value}</div>
    </div>
  );
}
