import { Link } from "react-router-dom";
import type { LineageEdge, LineageGraph, LineageNode } from "../api/client";
import StatusBadge from "./StatusBadge";

// Layout constants (px). Nodes are placed on a fixed grid so we can draw real
// SVG connector lines between them using deterministic coordinates.
const NODE_W = 120;
const NODE_H = 64;
const COL_GAP = 56; // horizontal space between node columns (room for arrows)
const ROW_GAP = 48; // vertical space between branch swimlanes
const STACK_GAP = 8; // extra gap when several nodes share a lane+column
const PAD_X = 16;
const PAD_Y = 20;
const LANE_GUTTER = 148; // left gutter for swimlane labels
const LANE_LABEL_H = 18;
const AXIS_H = 44; // date-tick strip under the graph (releaseDate axis only)

interface Placed {
  node: LineageNode;
  col: number;
  row: number;
  stack: number; // vertical index when several nodes share a cell
  x: number;
  y: number;
}

interface ColumnPlan {
  colOf: Map<string, number>;
  colCount: number;
  labels: string[]; // tick label per column; empty when unused
}

// LineageGraphView renders the version-lineage graph as branch swimlanes with
// real connecting lines. Parent edges chain versions within a lane; fork edges
// connect a source version to the first version of a forked release line.
//
// Horizontal placement depends on the selected axis:
//   - version: each MAIN version owns a block. Forked release versions are
//     shifted one column right of that MAIN node so the fork line can drop from
//     the MAIN bottom and enter the release node from the left. The next MAIN
//     version starts after the longest release-revision chain of the block.
//   - releaseDate: unique calendar dates are columns, so nodes sharing a date
//     line up in one vertical column. A date scale is drawn under the graph.
export default function LineageGraphView({ graph }: { graph: LineageGraph }) {
  const { laneOrder, lanes } = groupLanes(graph.nodes);
  const isDateAxis = graph.axis === "releaseDate";
  const plan = isDateAxis
    ? assignDateColumns(graph.nodes)
    : assignVersionColumns(graph.nodes, graph.edges, laneOrder, lanes);

  const parentOf = parentMap(graph.edges);

  // Occupancy per (row, col) so same-date nodes on one lane can stack.
  const occupancy = new Map<string, number>();
  const cellIndex = new Map<string, number>();
  const placedMeta: { node: LineageNode; col: number; row: number; stack: number }[] = [];
  laneOrder.forEach((bid, row) => {
    const ordered = orderLane(lanes.get(bid)!, parentOf);
    for (const node of ordered) {
      const col = plan.colOf.get(node.versionId);
      if (col === undefined) continue;
      const key = cellKey(row, col);
      const stack = occupancy.get(key) ?? 0;
      occupancy.set(key, stack + 1);
      cellIndex.set(node.versionId, stack);
      placedMeta.push({ node, col, row, stack });
    }
  });

  const rowHeight: number[] = laneOrder.map((_, row) => {
    let maxStack = 1;
    for (let c = 0; c < plan.colCount; c++) {
      maxStack = Math.max(maxStack, occupancy.get(cellKey(row, c)) ?? 0);
    }
    const stacks = Math.max(1, maxStack);
    return LANE_LABEL_H + stacks * NODE_H + Math.max(0, stacks - 1) * STACK_GAP;
  });
  const rowY: number[] = [];
  let yCursor = PAD_Y;
  for (const h of rowHeight) {
    rowY.push(yCursor);
    yCursor += h + ROW_GAP;
  }

  const originX = LANE_GUTTER + PAD_X;
  const placed = new Map<string, Placed>();
  for (const m of placedMeta) {
    placed.set(m.node.versionId, {
      ...m,
      x: originX + m.col * (NODE_W + COL_GAP),
      y: rowY[m.row] + LANE_LABEL_H + m.stack * (NODE_H + STACK_GAP),
    });
  }

  const width = originX + PAD_X + Math.max(1, plan.colCount) * NODE_W + Math.max(0, plan.colCount - 1) * COL_GAP;
  const plotBottom = yCursor - ROW_GAP + PAD_Y;
  const height = plotBottom + (isDateAxis ? AXIS_H : 0);
  const axisY = plotBottom - PAD_Y / 2;

  const rightMid = (p: Placed) => ({ x: p.x + NODE_W, y: p.y + NODE_H / 2 });
  const leftMid = (p: Placed) => ({ x: p.x, y: p.y + NODE_H / 2 });
  const topMid = (p: Placed) => ({ x: p.x + NODE_W / 2, y: p.y });
  const bottomMid = (p: Placed) => ({ x: p.x + NODE_W / 2, y: p.y + NODE_H });
  const placedList = Array.from(placed.values());

  const sameColForkIndex = new Map<string, number>();
  const sameColForkCount = new Map<string, number>();
  for (const e of graph.edges) {
    const s = placed.get(e.sourceVersionId);
    const t = placed.get(e.targetVersionId);
    if (!s || !t || s.col !== t.col || e.kind !== "fork") continue;
    const k = e.sourceVersionId;
    sameColForkIndex.set(`${k}->${e.targetVersionId}`, sameColForkCount.get(k) ?? 0);
    sameColForkCount.set(k, (sameColForkCount.get(k) ?? 0) + 1);
  }

  const edgePaths: { d: string; kind: "parent" | "fork"; key: string }[] = [];
  for (const e of graph.edges) {
    const s = placed.get(e.sourceVersionId);
    const t = placed.get(e.targetVersionId);
    if (!s || !t) continue;
    const key = `${e.kind}:${e.sourceVersionId}->${e.targetVersionId}`;

    if (e.kind === "parent" && s.row === t.row) {
      const from = rightMid(s);
      const to = leftMid(t);
      const midX = (from.x + to.x) / 2;
      edgePaths.push({
        key,
        kind: "parent",
        d: `M ${from.x} ${from.y} C ${midX} ${from.y}, ${midX} ${to.y}, ${to.x} ${to.y}`,
      });
    } else if (!isDateAxis && e.kind === "fork" && t.x > s.x) {
      // Version axis: drop from the MAIN bottom, then turn right into the
      // release node's left side (shared vertical spine when several lines fork).
      const from = bottomMid(s);
      const to = leftMid(t);
      edgePaths.push({
        key,
        kind: "fork",
        d: `M ${from.x} ${from.y} L ${from.x} ${to.y} L ${to.x} ${to.y}`,
      });
    } else if (s.col === t.col) {
      edgePaths.push({
        key,
        kind: e.kind,
        d: sameColumnPath(s, t, placedList, {
          index: sameColForkIndex.get(`${e.sourceVersionId}->${e.targetVersionId}`) ?? 0,
          count: sameColForkCount.get(e.sourceVersionId) ?? 1,
        }),
      });
    } else {
      const goingDown = t.row > s.row;
      const from = goingDown ? bottomMid(s) : topMid(s);
      const to = goingDown ? topMid(t) : bottomMid(t);
      const midY = (from.y + to.y) / 2;
      edgePaths.push({
        key,
        kind: e.kind,
        d: `M ${from.x} ${from.y} C ${from.x} ${midY}, ${to.x} ${midY}, ${to.x} ${to.y}`,
      });
    }
  }

  return (
    <div style={{ background: "#fff", border: "1px solid #e5e7eb", borderRadius: 8, padding: 16, overflowX: "auto" }}>
      <div style={{ fontSize: 12, color: "#6b7280", marginBottom: 12 }}>
        Axis: {graph.axis} · {graph.nodes.length} versions · {graph.edges.length} edges
        <span style={{ marginLeft: 16 }}>
          <LegendSwatch color="#9ca3af" /> parent
          <span style={{ marginLeft: 12 }} />
          <LegendSwatch color="#db2777" dashed /> fork
        </span>
      </div>

      <div style={{ position: "relative", width, height, minWidth: "100%" }}>
        <svg
          width={width}
          height={height}
          style={{ position: "absolute", top: 0, left: 0, pointerEvents: "none" }}
        >
          <defs>
            <marker id="arrow-parent" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">
              <path d="M0,0 L6,3 L0,6 Z" fill="#9ca3af" />
            </marker>
            <marker id="arrow-fork" markerWidth="8" markerHeight="8" refX="6" refY="3" orient="auto" markerUnits="strokeWidth">
              <path d="M0,0 L6,3 L0,6 Z" fill="#db2777" />
            </marker>
          </defs>

          {isDateAxis &&
            plan.labels.map((_, col) => {
              const x = originX + col * (NODE_W + COL_GAP) + NODE_W / 2;
              return (
                <line
                  key={`guide-${col}`}
                  x1={x}
                  y1={PAD_Y}
                  x2={x}
                  y2={axisY}
                  stroke="#f3f4f6"
                  strokeWidth={1}
                />
              );
            })}

          {edgePaths.map(ep => (
            <path
              key={ep.key}
              d={ep.d}
              fill="none"
              stroke={ep.kind === "fork" ? "#db2777" : "#9ca3af"}
              strokeWidth={ep.kind === "fork" ? 2 : 1.5}
              strokeDasharray={ep.kind === "fork" ? "5 4" : undefined}
              markerEnd={ep.kind === "fork" ? "url(#arrow-fork)" : "url(#arrow-parent)"}
            />
          ))}

          {isDateAxis && plan.colCount > 0 && (
            <DateAxis
              originX={originX}
              axisY={axisY}
              colCount={plan.colCount}
              labels={plan.labels}
            />
          )}
        </svg>

        {laneOrder.map((bid, row) => {
          const lane = lanes.get(bid)![0];
          return (
            <div
              key={`label-${bid}`}
              style={{
                position: "absolute",
                left: 8,
                top: rowY[row] + LANE_LABEL_H + NODE_H / 2 - 8,
                width: LANE_GUTTER - 12,
                fontSize: 11,
                fontWeight: 700,
                color: lane.branchType === "MAIN" ? "#3730a3" : "#9d174d",
                overflow: "hidden",
                textOverflow: "ellipsis",
                whiteSpace: "nowrap",
              }}
              title={`${lane.branchName} (${lane.branchType})`}
            >
              {lane.branchName} ({lane.branchType})
            </div>
          );
        })}

        {Array.from(placed.values()).map(p => (
          <div
            key={p.node.versionId}
            style={{
              position: "absolute",
              left: p.x,
              top: p.y,
              width: NODE_W,
              height: NODE_H,
              boxSizing: "border-box",
              border: "1px solid #e5e7eb",
              borderRadius: 6,
              padding: "6px 8px",
              background: "#f9fafb",
              display: "flex",
              flexDirection: "column",
              justifyContent: "center",
            }}
          >
            <Link to={`/versions/${p.node.versionId}`} style={{ color: "#1a56db", fontWeight: 600, fontSize: 13 }}>
              {p.node.versionString}
            </Link>
            <div style={{ marginTop: 4 }}>
              <StatusBadge status={p.node.status} />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

function DateAxis({
  originX,
  axisY,
  colCount,
  labels,
}: {
  originX: number;
  axisY: number;
  colCount: number;
  labels: string[];
}) {
  const firstX = originX + NODE_W / 2;
  const lastX = originX + (colCount - 1) * (NODE_W + COL_GAP) + NODE_W / 2;
  return (
    <g>
      <line x1={firstX} y1={axisY} x2={Math.max(firstX, lastX)} y2={axisY} stroke="#9ca3af" strokeWidth={1.5} />
      {labels.map((label, col) => {
        const x = originX + col * (NODE_W + COL_GAP) + NODE_W / 2;
        return (
          <g key={`tick-${col}`}>
            <line x1={x} y1={axisY - 5} x2={x} y2={axisY + 5} stroke="#6b7280" strokeWidth={1.5} />
            <text
              x={x}
              y={axisY + 18}
              textAnchor="middle"
              fill="#4b5563"
              fontSize={11}
              fontFamily="ui-sans-serif, system-ui, sans-serif"
            >
              {label}
            </text>
          </g>
        );
      })}
    </g>
  );
}

function LegendSwatch({ color, dashed }: { color: string; dashed?: boolean }) {
  return (
    <svg width="26" height="8" style={{ verticalAlign: "middle", marginRight: 4 }}>
      <line
        x1="0"
        y1="4"
        x2="26"
        y2="4"
        stroke={color}
        strokeWidth="2"
        strokeDasharray={dashed ? "5 4" : undefined}
      />
    </svg>
  );
}

function cellKey(row: number, col: number) {
  return `${row}:${col}`;
}

function columnHasNodeBetween(s: Placed, t: Placed, all: Placed[]): boolean {
  const lo = Math.min(s.y, t.y) + NODE_H;
  const hi = Math.max(s.y, t.y);
  return all.some(p => {
    if (p.node.versionId === s.node.versionId || p.node.versionId === t.node.versionId) return false;
    if (p.col !== s.col) return false;
    return p.y + NODE_H > lo && p.y < hi;
  });
}

// Vertical fork/parent in one column. If another node sits between source and
// target, route around the column so the line does not cut through a card.
function sameColumnPath(
  s: Placed,
  t: Placed,
  all: Placed[],
  fork: { index: number; count: number },
): string {
  const goingDown = t.y >= s.y;
  const from = goingDown
    ? { x: s.x + NODE_W / 2, y: s.y + NODE_H }
    : { x: s.x + NODE_W / 2, y: s.y };
  const to = goingDown
    ? { x: t.x + NODE_W / 2, y: t.y }
    : { x: t.x + NODE_W / 2, y: t.y + NODE_H };

  const blocked = columnHasNodeBetween(s, t, all);
  const stagger = fork.count > 1 ? (fork.index - (fork.count - 1) / 2) * 12 : 0;

  if (!blocked) {
    return `M ${from.x + stagger} ${from.y} L ${to.x + stagger} ${to.y}`;
  }

  const goLeft = fork.index % 2 === 0 && s.col > 0;
  const detourX = goLeft ? s.x - 18 : s.x + NODE_W + 18;
  const y1 = from.y + (goingDown ? 8 : -8);
  const y2 = to.y + (goingDown ? -8 : 8);
  return `M ${from.x} ${from.y} L ${from.x} ${y1} L ${detourX} ${y1} L ${detourX} ${y2} L ${to.x} ${y2} L ${to.x} ${to.y}`;
}

function groupLanes(nodes: LineageNode[]) {
  const laneOrder: string[] = [];
  const lanes = new Map<string, LineageNode[]>();
  for (const n of nodes) {
    if (!lanes.has(n.branchLineId)) {
      lanes.set(n.branchLineId, []);
      laneOrder.push(n.branchLineId);
    }
    lanes.get(n.branchLineId)!.push(n);
  }
  laneOrder.sort((a, b) => {
    const ta = lanes.get(a)![0].branchType;
    const tb = lanes.get(b)![0].branchType;
    if (ta !== tb) return ta === "MAIN" ? -1 : 1;
    return 0;
  });
  return { laneOrder, lanes };
}

function parentMap(edges: LineageEdge[]) {
  const parentOf = new Map<string, string>();
  for (const e of edges) {
    if (e.kind === "parent") parentOf.set(e.targetVersionId, e.sourceVersionId);
  }
  return parentOf;
}

function forkMap(edges: LineageEdge[]) {
  const forkOf = new Map<string, string>();
  for (const e of edges) {
    if (e.kind === "fork") forkOf.set(e.targetVersionId, e.sourceVersionId);
  }
  return forkOf;
}

function orderLane(nodes: LineageNode[], parentOf: Map<string, string>): LineageNode[] {
  const ids = new Set(nodes.map(n => n.versionId));
  const byId = new Map(nodes.map(n => [n.versionId, n]));
  const children = new Map<string, string[]>();
  for (const [child, parent] of parentOf) {
    if (!ids.has(child) || !ids.has(parent)) continue;
    if (!children.has(parent)) children.set(parent, []);
    children.get(parent)!.push(child);
  }
  const roots = nodes.filter(n => !parentOf.has(n.versionId) || !ids.has(parentOf.get(n.versionId)!));
  const out: LineageNode[] = [];
  const seen = new Set<string>();
  const walk = (id: string) => {
    if (seen.has(id)) return;
    seen.add(id);
    const n = byId.get(id);
    if (n) out.push(n);
    for (const c of children.get(id) || []) walk(c);
  };
  for (const r of roots) walk(r.versionId);
  for (const n of nodes) if (!seen.has(n.versionId)) out.push(n);
  return out;
}

function walkToMain(
  startId: string,
  nodeById: Map<string, LineageNode>,
  parentOf: Map<string, string>,
  forkOf: Map<string, string>,
): string | null {
  const seen = new Set<string>();
  let id: string | undefined = startId;
  while (id && !seen.has(id)) {
    seen.add(id);
    const n = nodeById.get(id);
    if (n?.branchType === "MAIN") return id;
    id = forkOf.get(id) || parentOf.get(id);
  }
  return null;
}

function mainChain(nodes: LineageNode[], parentOf: Map<string, string>): LineageNode[] {
  const mains = nodes.filter(n => n.branchType === "MAIN");
  const ids = new Set(mains.map(n => n.versionId));
  const byId = new Map(mains.map(n => [n.versionId, n]));
  const children = new Map<string, string[]>();
  for (const [child, parent] of parentOf) {
    if (!ids.has(child) || !ids.has(parent)) continue;
    if (!children.has(parent)) children.set(parent, []);
    children.get(parent)!.push(child);
  }
  const roots = mains.filter(n => !parentOf.has(n.versionId) || !ids.has(parentOf.get(n.versionId)!));
  const out: LineageNode[] = [];
  const seen = new Set<string>();
  const walk = (id: string) => {
    if (seen.has(id)) return;
    seen.add(id);
    const n = byId.get(id);
    if (n) out.push(n);
    for (const c of children.get(id) || []) walk(c);
  };
  for (const r of roots) walk(r.versionId);
  for (const n of mains) if (!seen.has(n.versionId)) out.push(n);
  return out;
}

// assignVersionColumns groups release-line revisions under the MAIN version they
// forked from. Forked releases start one column to the right of that MAIN node.
// The next MAIN version is placed after the block's longest revision chain.
function assignVersionColumns(
  nodes: LineageNode[],
  edges: LineageEdge[],
  laneOrder: string[],
  lanes: Map<string, LineageNode[]>,
): ColumnPlan {
  const nodeById = new Map(nodes.map(n => [n.versionId, n]));
  const parentOf = parentMap(edges);
  const forkOf = forkMap(edges);
  const mains = mainChain(nodes, parentOf);

  const lanesByAnchor = new Map<string, string[]>();
  const orphanLanes: string[] = [];
  for (const bid of laneOrder) {
    const laneNodes = lanes.get(bid)!;
    if (laneNodes[0].branchType === "MAIN") continue;
    const ordered = orderLane(laneNodes, parentOf);
    const anchor = walkToMain(ordered[0].versionId, nodeById, parentOf, forkOf);
    if (!anchor) {
      orphanLanes.push(bid);
      continue;
    }
    if (!lanesByAnchor.has(anchor)) lanesByAnchor.set(anchor, []);
    lanesByAnchor.get(anchor)!.push(bid);
  }

  const colOf = new Map<string, number>();
  let col = 0;
  for (const main of mains) {
    const related = lanesByAnchor.get(main.versionId) || [];
    const longest = related.reduce(
      (max, bid) => Math.max(max, orderLane(lanes.get(bid)!, parentOf).length),
      0,
    );
    // +1 column so the first forked release sits to the right of MAIN.
    const width = 1 + longest;
    colOf.set(main.versionId, col);
    for (const bid of related) {
      orderLane(lanes.get(bid)!, parentOf).forEach((n, i) => colOf.set(n.versionId, col + 1 + i));
    }
    col += width;
  }
  for (const bid of orphanLanes) {
    const ordered = orderLane(lanes.get(bid)!, parentOf);
    ordered.forEach((n, i) => colOf.set(n.versionId, col + i));
    col += Math.max(1, ordered.length);
  }

  return { colOf, colCount: Math.max(1, col), labels: [] };
}

function dateKey(n: LineageNode): string {
  return n.releaseDate ? n.releaseDate.slice(0, 10) : "";
}

function formatDateTick(key: string): string {
  return key || "(no date)";
}

// assignDateColumns maps every calendar date to one column so versions released
// on the same day share an x-position (vertical alignment across swimlanes).
function assignDateColumns(nodes: LineageNode[]): ColumnPlan {
  const dated: string[] = [];
  const seen = new Set<string>();
  let hasUndated = false;
  const sorted = [...nodes].sort((a, b) => {
    const da = dateKey(a);
    const db = dateKey(b);
    if (da && db && da !== db) return da < db ? -1 : 1;
    if (da && !db) return -1;
    if (!da && db) return 1;
    return a.createdAt < b.createdAt ? -1 : a.createdAt > b.createdAt ? 1 : 0;
  });
  for (const n of sorted) {
    const k = dateKey(n);
    if (!k) {
      hasUndated = true;
      continue;
    }
    if (!seen.has(k)) {
      seen.add(k);
      dated.push(k);
    }
  }
  const keys = hasUndated ? [...dated, ""] : dated;
  const index = new Map(keys.map((k, i) => [k, i]));
  const colOf = new Map<string, number>();
  for (const n of nodes) {
    colOf.set(n.versionId, index.get(dateKey(n)) ?? keys.length - 1);
  }
  return {
    colOf,
    colCount: Math.max(1, keys.length),
    labels: keys.map(formatDateTick),
  };
}
