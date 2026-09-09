import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";
import type { DTProject } from "../api/client";

interface Props {
  title?: string;
  currentUuid?: string;
  onSelect: (uuid: string) => void;
  onClose: () => void;
  saving?: boolean;
  error?: boolean;
}

// DTProjectSelector is a generic Dependency-Track project picker. It returns the
// chosen project UUID via onSelect; the caller decides how to persist it.
export default function DTProjectSelector({ title, currentUuid, onSelect, onClose, saving, error }: Props) {
  const [search, setSearch] = useState("");
  const [selected, setSelected] = useState<string | undefined>(currentUuid);
  const [manualUuid, setManualUuid] = useState("");

  const { data: projects, isLoading, error: loadError } = useQuery({
    queryKey: ["dtProjects", search],
    queryFn: () => api.searchDTProjects(search || undefined),
  });

  const handleSave = () => {
    const uuid = manualUuid.trim() || selected;
    if (uuid) onSelect(uuid);
  };

  return (
    <div style={overlayStyle} onClick={onClose}>
      <div style={modalStyle} onClick={(e) => e.stopPropagation()}>
        <h3 style={{ fontSize: 18, fontWeight: 700, marginBottom: 16 }}>{title || "Select DT Project"}</h3>

        <input
          type="text"
          placeholder="Search projects by name..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={inputStyle}
        />

        <div style={{ maxHeight: 240, overflowY: "auto", margin: "12px 0", border: "1px solid #e5e7eb", borderRadius: 6 }}>
          {isLoading && <p style={{ padding: 12, color: "#6b7280" }}>Loading...</p>}
          {loadError && <p style={{ padding: 12, color: "#dc2626" }}>Failed to load DT projects. Check Dependency-Track connection.</p>}
          {projects && projects.length === 0 && <p style={{ padding: 12, color: "#6b7280" }}>No projects found.</p>}
          {projects?.map((p: DTProject) => (
            <div
              key={p.uuid}
              onClick={() => { setSelected(p.uuid); setManualUuid(""); }}
              style={{
                padding: 10,
                cursor: "pointer",
                background: selected === p.uuid && !manualUuid ? "#eff6ff" : "#fff",
                borderBottom: "1px solid #f3f4f6",
              }}
            >
              <div style={{ fontWeight: 600, fontSize: 14 }}>{p.name} <span style={{ color: "#6b7280", fontWeight: 400 }}>v{p.version}</span></div>
              <code style={{ fontSize: 11, color: "#6b7280" }}>{p.uuid}</code>
            </div>
          ))}
        </div>

        <div style={{ marginBottom: 16 }}>
          <label style={{ fontSize: 12, color: "#6b7280" }}>Or enter UUID manually:</label>
          <input
            type="text"
            placeholder="00000000-0000-0000-0000-000000000000"
            value={manualUuid}
            onChange={(e) => { setManualUuid(e.target.value); setSelected(undefined); }}
            style={inputStyle}
          />
        </div>

        {error && (
          <p style={{ color: "#dc2626", fontSize: 13, marginBottom: 12 }}>Failed to save. Please try again.</p>
        )}

        <div style={{ display: "flex", justifyContent: "flex-end", gap: 8 }}>
          <button onClick={onClose} style={btnSecondary}>Cancel</button>
          <button
            onClick={handleSave}
            disabled={(!selected && !manualUuid.trim()) || saving}
            style={{ ...btnPrimary, opacity: (!selected && !manualUuid.trim()) || saving ? 0.5 : 1 }}
          >
            {saving ? "Saving..." : "Save"}
          </button>
        </div>
      </div>
    </div>
  );
}

const overlayStyle: React.CSSProperties = {
  position: "fixed", top: 0, left: 0, right: 0, bottom: 0,
  background: "rgba(0,0,0,0.4)", display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
};
const modalStyle: React.CSSProperties = {
  background: "#fff", borderRadius: 8, padding: 24, width: 480, maxWidth: "90vw", boxShadow: "0 10px 40px rgba(0,0,0,0.2)",
};
const inputStyle: React.CSSProperties = {
  width: "100%", padding: 8, border: "1px solid #d1d5db", borderRadius: 6, fontSize: 14, boxSizing: "border-box",
};
const btnPrimary: React.CSSProperties = {
  padding: "8px 16px", background: "#1a56db", color: "#fff", border: "none", borderRadius: 6, cursor: "pointer", fontSize: 14, fontWeight: 600,
};
const btnSecondary: React.CSSProperties = {
  padding: "8px 16px", background: "#fff", color: "#374151", border: "1px solid #d1d5db", borderRadius: 6, cursor: "pointer", fontSize: 14,
};
