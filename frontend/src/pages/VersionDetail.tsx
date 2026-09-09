import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { api } from "../api/client";
import type { Role, VersionDTProject, VersionStatus } from "../api/client";
import StatusBadge from "../components/StatusBadge";
import GraphView from "../components/GraphView";
import DTProjectSelector from "../components/DTProjectSelector";

export default function VersionDetail() {
  const { versionId } = useParams<{ versionId: string }>();
  const queryClient = useQueryClient();
  const { data: version, isLoading } = useQuery({
    queryKey: ["version", versionId],
    queryFn: () => api.getVersion(versionId!),
  });
  const { data: branch } = useQuery({
    queryKey: ["branch", version?.branchLineId],
    queryFn: () => api.getBranch(version!.branchLineId),
    enabled: !!version?.branchLineId,
  });

  const [picker, setPicker] = useState<null | { role: Role }>(null);

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ["version", versionId] });

  const setRoot = useMutation({
    mutationFn: (uuid: string) => api.setVersionRoot(versionId!, { dtProjectUuid: uuid }),
    onSuccess: () => { invalidate(); setPicker(null); },
  });
  const addProject = useMutation({
    mutationFn: (vars: { role: "PROFILE" | "SUB"; uuid: string }) =>
      api.addVersionProject(versionId!, { role: vars.role, dtProjectUuid: vars.uuid }),
    onSuccess: () => { invalidate(); setPicker(null); },
  });
  const removeProject = useMutation({
    mutationFn: (bindingId: number) => api.removeVersionProject(versionId!, bindingId),
    onSuccess: invalidate,
  });

  if (isLoading || !version) return <p>Loading...</p>;

  const projects = version.projects || [];
  const root = projects.find(p => p.role === "ROOT");
  const profiles = projects.filter(p => p.role === "PROFILE");
  const subs = projects.filter(p => p.role === "SUB");
  const isMain = branch?.type === "MAIN";

  const handlePick = (uuid: string) => {
    if (!picker) return;
    if (picker.role === "ROOT") setRoot.mutate(uuid);
    else addProject.mutate({ role: picker.role, uuid });
  };

  return (
    <div>
      <nav style={{ fontSize: 14, color: "#6b7280", marginBottom: 16 }}>
        <Link to="/products" style={{ color: "#1a56db" }}>Products</Link>
        {" / "}
        <Link to={`/branches/${version.branchLineId}`} style={{ color: "#1a56db" }}>
          {branch?.displayName || branch?.name || "Branch"}
        </Link>
        {" / "}
        {version.versionString}
      </nav>

      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16 }}>
        <h1 style={{ fontSize: 24, fontWeight: 700 }}>Version {version.versionString}</h1>
        <StatusBadge status={version.status} />
      </div>

      <VersionEditor versionId={version.id} initial={version} onSaved={invalidate} />

      {/* Role-tagged DT projects */}
      <section style={{ marginTop: 32 }}>
        <h2 style={{ fontSize: 18, fontWeight: 600, marginBottom: 12 }}>DT Projects (roles)</h2>

        {/* ROOT */}
        <RoleBlock
          title="ROOT"
          hint="Exactly one root project."
          onAdd={() => setPicker({ role: "ROOT" })}
          addLabel={root ? "Change ROOT" : "Set ROOT"}
        >
          {root ? (
            <ProjectRow versionId={version.id} p={root} onRemove={() => removeProject.mutate(root.id)} removable />
          ) : (
            <Empty text="No ROOT project set." />
          )}
        </RoleBlock>

        {/* PROFILE (release only) */}
        {!isMain && (
          <RoleBlock
            title="PROFILE"
            hint="Environmental differences (release branches only)."
            onAdd={() => setPicker({ role: "PROFILE" })}
            addLabel="+ Add PROFILE"
          >
            {profiles.length === 0 ? <Empty text="No PROFILE projects." /> :
              profiles.map(p => <ProjectRow key={p.id} versionId={version.id} p={p} onRemove={() => removeProject.mutate(p.id)} removable />)}
          </RoleBlock>
        )}

        {/* SUB */}
        <RoleBlock
          title="SUB"
          hint="Additional sub-projects."
          onAdd={() => setPicker({ role: "SUB" })}
          addLabel="+ Add SUB"
        >
          {subs.length === 0 ? <Empty text="No SUB projects." /> :
            subs.map(p => <ProjectRow key={p.id} versionId={version.id} p={p} onRemove={() => removeProject.mutate(p.id)} removable />)}
        </RoleBlock>
      </section>

      {picker && (
        <DTProjectSelector
          title={`Select ${picker.role} project`}
          currentUuid={picker.role === "ROOT" ? root?.dtProjectUuid : undefined}
          onSelect={handlePick}
          onClose={() => setPicker(null)}
          saving={setRoot.isPending || addProject.isPending}
          error={setRoot.isError || addProject.isError}
        />
      )}
    </div>
  );
}

function VersionEditor({ versionId, initial, onSaved }: {
  versionId: string;
  initial: { versionString: string; status: VersionStatus; location?: string; customer?: string; releaseDate?: string };
  onSaved: () => void;
}) {
  const [versionString, setVersionString] = useState(initial.versionString);
  const [status, setStatus] = useState<VersionStatus>(initial.status);
  const [location, setLocation] = useState(initial.location || "");
  const [customer, setCustomer] = useState(initial.customer || "");
  const [releaseDate, setReleaseDate] = useState(initial.releaseDate ? initial.releaseDate.slice(0, 10) : "");

  const save = useMutation({
    mutationFn: () => api.updateVersion(versionId, {
      versionString,
      status,
      location,
      customer,
      releaseDate: releaseDate ? new Date(releaseDate).toISOString() : "",
    }),
    onSuccess: onSaved,
  });

  return (
    <section style={cardStyle}>
      <h2 style={{ fontSize: 16, fontWeight: 600, marginBottom: 12 }}>Details</h2>
      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 12 }}>
        <Field label="Version string">
          <input value={versionString} onChange={e => setVersionString(e.target.value)} style={inputStyle} />
        </Field>
        <Field label="Status">
          <select value={status} onChange={e => setStatus(e.target.value as VersionStatus)} style={inputStyle}>
            <option value="incomplete">incomplete</option>
            <option value="finalized">finalized</option>
          </select>
        </Field>
        <Field label="Location">
          <input value={location} onChange={e => setLocation(e.target.value)} style={inputStyle} />
        </Field>
        <Field label="Customer">
          <input value={customer} onChange={e => setCustomer(e.target.value)} style={inputStyle} />
        </Field>
        <Field label="Release date">
          <input type="date" value={releaseDate} onChange={e => setReleaseDate(e.target.value)} style={inputStyle} />
        </Field>
      </div>
      <div style={{ marginTop: 12, display: "flex", gap: 8, alignItems: "center" }}>
        <button onClick={() => save.mutate()} style={btnPrimary} disabled={save.isPending || !versionString}>
          {save.isPending ? "Saving..." : "Save"}
        </button>
        {save.isError && <span style={{ color: "#dc2626", fontSize: 13 }}>Save failed.</span>}
        {save.isSuccess && <span style={{ color: "#166534", fontSize: 13 }}>Saved.</span>}
      </div>
    </section>
  );
}

function RoleBlock({ title, hint, addLabel, onAdd, children }: {
  title: string; hint: string; addLabel: string; onAdd: () => void; children: React.ReactNode;
}) {
  return (
    <div style={{ ...cardStyle, marginBottom: 12 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
        <div>
          <span style={{ fontWeight: 700 }}>{title}</span>
          <span style={{ fontSize: 12, color: "#6b7280", marginLeft: 8 }}>{hint}</span>
        </div>
        <button onClick={onAdd} style={editBtnStyle}>{addLabel}</button>
      </div>
      {children}
    </div>
  );
}

function ProjectRow({ versionId, p, onRemove, removable }: {
  versionId: string; p: VersionDTProject; onRemove: () => void; removable?: boolean;
}) {
  const [showGraph, setShowGraph] = useState(false);
  const { data: graph, isLoading, isError, error } = useQuery({
    queryKey: ["projectGraph", versionId, p.role, p.id],
    queryFn: () => api.getProjectGraph(versionId, p.role),
    enabled: showGraph && !!p.dtProjectUuid,
  });
  const dtUnavailable = isError && (error as { error?: string })?.error === "DEPENDENCY_TRACK_UNAVAILABLE";

  return (
    <div style={{ borderTop: "1px solid #f3f4f6", paddingTop: 8, marginTop: 8 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 8, fontSize: 13 }}>
        <code style={{ wordBreak: "break-all" }}>{p.dtProjectUuid || "—"}</code>
        <div style={{ display: "flex", gap: 8, flexShrink: 0 }}>
          {p.dtProjectUuid && (
            <button onClick={() => setShowGraph(v => !v)} style={editBtnStyleSecondary}>
              {showGraph ? "Hide SBOM" : "View SBOM"}
            </button>
          )}
          {removable && <button onClick={onRemove} style={dangerBtnStyle}>Remove</button>}
        </div>
      </div>
      {showGraph && (
        <div style={{ marginTop: 8 }}>
          {isLoading && <p style={{ color: "#6b7280", fontSize: 13 }}>Resolving SBOM…</p>}
          {dtUnavailable && <p style={{ color: "#dc2626", fontSize: 13 }}>Dependency-Track unavailable (502).</p>}
          {isError && !dtUnavailable && <p style={{ color: "#dc2626", fontSize: 13 }}>Failed to resolve SBOM graph.</p>}
          {graph && graph.nodes && graph.nodes.length > 0 && <GraphView graph={graph} />}
          {graph && (!graph.nodes || graph.nodes.length === 0) && <p style={{ color: "#6b7280", fontSize: 13 }}>Empty graph.</p>}
        </div>
      )}
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label style={{ display: "flex", flexDirection: "column", gap: 4 }}>
      <span style={{ fontSize: 12, color: "#6b7280" }}>{label}</span>
      {children}
    </label>
  );
}

function Empty({ text }: { text: string }) {
  return <p style={{ color: "#6b7280", fontSize: 13 }}>{text}</p>;
}

const cardStyle: React.CSSProperties = { background: "#fff", border: "1px solid #e5e7eb", borderRadius: 8, padding: 16 };
const inputStyle: React.CSSProperties = { border: "1px solid #d1d5db", borderRadius: 6, padding: "8px 12px", fontSize: 14, width: "100%", boxSizing: "border-box" };
const btnPrimary: React.CSSProperties = { background: "#1a56db", color: "#fff", border: "none", borderRadius: 6, padding: "8px 16px", cursor: "pointer", fontWeight: 600, fontSize: 14 };
const editBtnStyle: React.CSSProperties = { padding: "6px 12px", background: "#1a56db", color: "#fff", border: "none", borderRadius: 6, cursor: "pointer", fontSize: 13, fontWeight: 600 };
const editBtnStyleSecondary: React.CSSProperties = { padding: "6px 12px", background: "#fff", color: "#1a56db", border: "1px solid #1a56db", borderRadius: 6, cursor: "pointer", fontSize: 13, fontWeight: 600 };
const dangerBtnStyle: React.CSSProperties = { padding: "6px 12px", background: "#fff", color: "#dc2626", border: "1px solid #fca5a5", borderRadius: 6, cursor: "pointer", fontSize: 13 };
