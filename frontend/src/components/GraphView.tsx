import type { ReleaseGraph, GraphNode } from "../api/client";
import StatusBadge from "./StatusBadge";

interface Props {
  graph: ReleaseGraph;
}

export default function GraphView({ graph }: Props) {
  const nodeMap = new Map(graph.nodes.map(n => [n.id, n]));
  const rootNode = nodeMap.get(graph.rootNodeId);

  // Build adjacency list
  const children = new Map<string, { nodeId: string; bomRef?: string }[]>();
  for (const edge of graph.edges) {
    if (!children.has(edge.sourceNodeId)) children.set(edge.sourceNodeId, []);
    children.get(edge.sourceNodeId)!.push({ nodeId: edge.targetNodeId, bomRef: edge.bomRef });
  }

  return (
    <div style={{ background: "#fff", border: "1px solid #e5e7eb", borderRadius: 8, padding: 16 }}>
      {/* Metadata */}
      <div style={{ display: "flex", gap: 16, fontSize: 12, color: "#6b7280", marginBottom: 16 }}>
        <span>Nodes: {graph.metadata.totalNodes}</span>
        <span>Edges: {graph.metadata.totalEdges}</span>
        {graph.metadata.unresolvedLinks > 0 && <span style={{ color: "#dc2626" }}>Unresolved: {graph.metadata.unresolvedLinks}</span>}
        {graph.metadata.cyclesDetected > 0 && <span style={{ color: "#d97706" }}>Cycles: {graph.metadata.cyclesDetected}</span>}
      </div>

      {/* Tree visualization */}
      {rootNode && <TreeNode node={rootNode} children={children} nodeMap={nodeMap} visited={new Set()} depth={0} />}
    </div>
  );
}

function TreeNode({
  node,
  children,
  nodeMap,
  visited,
  depth,
}: {
  node: GraphNode;
  children: Map<string, { nodeId: string; bomRef?: string }[]>;
  nodeMap: Map<string, GraphNode>;
  visited: Set<string>;
  depth: number;
}) {
  if (visited.has(node.id)) {
    return (
      <div style={{ marginLeft: depth * 24, padding: "4px 0", fontSize: 13 }}>
        <span style={{ color: "#d97706" }}>&#8635; {node.projectName} {node.projectVersion} (cycle)</span>
      </div>
    );
  }
  visited.add(node.id);

  const childEdges = children.get(node.id) || [];

  return (
    <div style={{ marginLeft: depth * 24 }}>
      <div style={{ display: "flex", alignItems: "center", gap: 8, padding: "6px 0" }}>
        <div style={{
          width: 8, height: 8, borderRadius: "50%",
          background: node.resolutionStatus === "resolved" ? "#22c55e" : "#ef4444"
        }} />
        <span style={{ fontWeight: 600, fontSize: 14 }}>{node.projectName}</span>
        <span style={{ fontSize: 12, color: "#6b7280" }}>{node.projectVersion}</span>
        {node.resolutionStatus !== "resolved" && <StatusBadge status={node.resolutionStatus} />}
      </div>
      {childEdges.map((edge, i) => {
        const childNode = nodeMap.get(edge.nodeId);
        if (!childNode) return null;
        return (
          <TreeNode
            key={`${edge.nodeId}-${i}`}
            node={childNode}
            children={children}
            nodeMap={nodeMap}
            visited={visited}
            depth={depth + 1}
          />
        );
      })}
    </div>
  );
}
