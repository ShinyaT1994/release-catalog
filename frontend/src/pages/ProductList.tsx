import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Link } from "react-router-dom";
import { api } from "../api/client";
import type { Product } from "../api/client";

export default function ProductList() {
  const queryClient = useQueryClient();
  const { data: products, isLoading, error } = useQuery({ queryKey: ["products"], queryFn: api.listProducts });
  const [showForm, setShowForm] = useState(false);
  const [name, setName] = useState("");
  const [displayName, setDisplayName] = useState("");

  const createMutation = useMutation({
    mutationFn: (data: { name: string; displayName: string }) => api.createProduct(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["products"] });
      setShowForm(false);
      setName("");
      setDisplayName("");
    },
  });

  if (isLoading) return <p>Loading...</p>;
  if (error) return <p style={{ color: "red" }}>Error loading products</p>;

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16 }}>
        <h1 style={{ fontSize: 24, fontWeight: 700 }}>Products</h1>
        <button onClick={() => setShowForm(!showForm)} style={btnStyle}>+ New Product</button>
      </div>

      {showForm && (
        <div style={{ background: "#f9fafb", border: "1px solid #e5e7eb", borderRadius: 8, padding: 16, marginBottom: 16 }}>
          <div style={{ display: "flex", gap: 8, flexWrap: "wrap" }}>
            <input placeholder="Name (unique)" value={name} onChange={e => setName(e.target.value)} style={inputStyle} />
            <input placeholder="Display Name" value={displayName} onChange={e => setDisplayName(e.target.value)} style={inputStyle} />
            <button onClick={() => createMutation.mutate({ name, displayName })} style={btnStyle} disabled={!name}>
              Create
            </button>
          </div>
        </div>
      )}

      {(!products || products.length === 0) ? (
        <p style={{ color: "#6b7280" }}>No products yet. Create one to get started.</p>
      ) : (
        <div style={{ display: "grid", gap: 12 }}>
          {products.map((p: Product) => (
            <Link
              key={p.id}
              to={`/products/${p.id}`}
              style={{ textDecoration: "none", color: "inherit", display: "block", background: "#fff", border: "1px solid #e5e7eb", borderRadius: 8, padding: 16 }}
            >
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
                <div>
                  <h3 style={{ fontSize: 16, fontWeight: 600 }}>{p.displayName || p.name}</h3>
                  <p style={{ fontSize: 12, color: "#6b7280" }}>{p.name}</p>
                </div>
                <span style={{ fontSize: 12, color: "#9ca3af" }}>{new Date(p.createdAt).toLocaleDateString()}</span>
              </div>
              {p.description && <p style={{ fontSize: 14, color: "#4b5563", marginTop: 8 }}>{p.description}</p>}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

const btnStyle: React.CSSProperties = { background: "#1a56db", color: "#fff", border: "none", borderRadius: 6, padding: "8px 16px", cursor: "pointer", fontWeight: 600, fontSize: 14 };
const inputStyle: React.CSSProperties = { border: "1px solid #d1d5db", borderRadius: 6, padding: "8px 12px", fontSize: 14 };
