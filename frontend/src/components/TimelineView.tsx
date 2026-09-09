import { Link } from "react-router-dom";
import type { TimelineEntry } from "../api/client";
import StatusBadge from "./StatusBadge";

// TimelineView renders release timeline entries in chronological order (as
// returned by the API), grouped visually by a vertical rail.
export default function TimelineView({ entries }: { entries: TimelineEntry[] }) {
  return (
    <div style={{ background: "#fff", border: "1px solid #e5e7eb", borderRadius: 8, padding: 16 }}>
      <div style={{ position: "relative", paddingLeft: 20 }}>
        <div style={railStyle} />
        {entries.map((e) => (
          <div key={e.versionId} style={{ position: "relative", padding: "8px 0" }}>
            <div style={{ ...dotStyle, background: e.status === "finalized" ? "#22c55e" : "#f59e0b" }} />
            <div style={{ display: "flex", alignItems: "center", gap: 8, flexWrap: "wrap" }}>
              <span style={{ fontSize: 12, color: "#6b7280", minWidth: 92 }}>
                {e.releaseDate ? new Date(e.releaseDate).toLocaleDateString() : "(no date)"}
              </span>
              <span style={{
                fontSize: 11, fontWeight: 600, padding: "1px 6px", borderRadius: 4,
                background: e.branchType === "MAIN" ? "#e0e7ff" : "#fce7f3",
                color: e.branchType === "MAIN" ? "#3730a3" : "#9d174d",
              }}>
                {e.branchName}
              </span>
              <Link to={`/versions/${e.versionId}`} style={{ color: "#1a56db", fontWeight: 600, fontSize: 14 }}>
                {e.versionString}
              </Link>
              <StatusBadge status={e.status} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

const railStyle: React.CSSProperties = {
  position: "absolute", left: 3, top: 0, bottom: 0, width: 2, background: "#e5e7eb",
};
const dotStyle: React.CSSProperties = {
  position: "absolute", left: -20, top: 14, width: 8, height: 8, borderRadius: "50%", border: "2px solid #fff",
};
