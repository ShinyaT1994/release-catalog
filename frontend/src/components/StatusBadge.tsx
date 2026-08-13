const colors: Record<string, { bg: string; text: string }> = {
  active: { bg: "#dcfce7", text: "#166534" },
  maintenance: { bg: "#fef9c3", text: "#854d0e" },
  security_only: { bg: "#fee2e2", text: "#991b1b" },
  end_of_support: { bg: "#f3f4f6", text: "#6b7280" },
  closed: { bg: "#e5e7eb", text: "#374151" },
  released: { bg: "#dbeafe", text: "#1e40af" },
  draft: { bg: "#f3f4f6", text: "#6b7280" },
  resolved: { bg: "#dcfce7", text: "#166534" },
  missing_project: { bg: "#fee2e2", text: "#991b1b" },
  missing_bom: { bg: "#fef9c3", text: "#854d0e" },
};

export default function StatusBadge({ status }: { status: string }) {
  const c = colors[status] || { bg: "#f3f4f6", text: "#374151" };
  return (
    <span
      style={{
        display: "inline-block",
        padding: "2px 8px",
        borderRadius: 4,
        fontSize: 12,
        fontWeight: 600,
        background: c.bg,
        color: c.text,
      }}
    >
      {status.replace(/_/g, " ")}
    </span>
  );
}
