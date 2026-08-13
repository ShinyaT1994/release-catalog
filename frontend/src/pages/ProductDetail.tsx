import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useParams, Link } from "react-router-dom";
import { useState } from "react";
import { api } from "../api/client";
import StatusBadge from "../components/StatusBadge";

export default function ProductDetail() {
  const { productId } = useParams<{ productId: string }>();
  const queryClient = useQueryClient();
  const { data: product, isLoading } = useQuery({ queryKey: ["product", productId], queryFn: () => api.getProduct(productId!) });
  const { data: branches } = useQuery({ queryKey: ["branches", productId], queryFn: () => api.listBranches(productId!) });
  const [showForm, setShowForm] = useState(false);
  const [rlName, setRlName] = useState("");
  const [rlDisplayName, setRlDisplayName] = useState("");

  const createRL = useMutation({
    mutationFn: () => api.createReleaseLine(productId!, { name: rlName, displayName: rlDisplayName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["branches", productId] });
      setShowForm(false);
      setRlName("");
      setRlDisplayName("");
    },
  });

  if (isLoading || !product) return <p>Loading...</p>;

  const mainBranch = branches?.find(b => b.type === "MAIN");
  const releaseBranches = branches?.filter(b => b.type === "RELEASE") || [];

  return (
    <div>
      <nav style={{ fontSize: 14, color: "#6b7280", marginBottom: 16 }}>
        <Link to="/products" style={{ color: "#1a56db" }}>Products</Link> / {product.displayName || product.name}
      </nav>

      <h1 style={{ fontSize: 24, fontWeight: 700, marginBottom: 8 }}>{product.displayName || product.name}</h1>
      {product.description && <p style={{ color: "#4b5563", marginBottom: 24 }}>{product.description}</p>}

      {/* Main Branch */}
      {mainBranch && (
        <section style={{ marginBottom: 32 }}>
          <h2 style={{ fontSize: 18, fontWeight: 600, marginBottom: 12 }}>Main Branch</h2>
          <Link to={`/branches/${mainBranch.id}`} style={{ textDecoration: "none", color: "inherit" }}>
            <div style={cardStyle}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                <span style={{ fontWeight: 600 }}>main</span>
                <StatusBadge status={mainBranch.status} />
              </div>
            </div>
          </Link>
        </section>
      )}

      {/* Release Lines */}
      <section>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
          <h2 style={{ fontSize: 18, fontWeight: 600 }}>Release Lines</h2>
          <button onClick={() => setShowForm(!showForm)} style={btnStyle}>+ New Release Line</button>
        </div>

        {showForm && (
          <div style={{ background: "#f9fafb", border: "1px solid #e5e7eb", borderRadius: 8, padding: 16, marginBottom: 12 }}>
            <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
              <input placeholder="Name (e.g. release/2.x)" value={rlName} onChange={e => setRlName(e.target.value)} style={inputStyle} />
              <input placeholder="Display Name" value={rlDisplayName} onChange={e => setRlDisplayName(e.target.value)} style={inputStyle} />
              <button onClick={() => createRL.mutate()} style={btnStyle} disabled={!rlName}>Create</button>
            </div>
          </div>
        )}

        {releaseBranches.length === 0 ? (
          <p style={{ color: "#6b7280" }}>No release lines yet.</p>
        ) : (
          <div style={{ display: "grid", gap: 8 }}>
            {releaseBranches.map(b => (
              <Link key={b.id} to={`/branches/${b.id}`} style={{ textDecoration: "none", color: "inherit" }}>
                <div style={cardStyle}>
                  <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                    <div>
                      <span style={{ fontWeight: 600 }}>{b.displayName || b.name}</span>
                      <span style={{ fontSize: 12, color: "#6b7280", marginLeft: 8 }}>{b.name}</span>
                    </div>
                    <StatusBadge status={b.status} />
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}

const cardStyle: React.CSSProperties = { background: "#fff", border: "1px solid #e5e7eb", borderRadius: 8, padding: 16 };
const btnStyle: React.CSSProperties = { background: "#1a56db", color: "#fff", border: "none", borderRadius: 6, padding: "8px 16px", cursor: "pointer", fontWeight: 600, fontSize: 14 };
const inputStyle: React.CSSProperties = { border: "1px solid #d1d5db", borderRadius: 6, padding: "8px 12px", fontSize: 14 };
