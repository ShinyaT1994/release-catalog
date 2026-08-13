const API_BASE = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...options?.headers },
    ...options,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "UNKNOWN", message: res.statusText }));
    throw err;
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

// --- Types ---
export interface Product {
  id: string;
  name: string;
  displayName: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface BranchLine {
  id: string;
  productId: string;
  type: "MAIN" | "RELEASE";
  name: string;
  displayName: string;
  sourceBranchLineId?: string;
  forkedFromSnapshotId?: string;
  status: "active" | "maintenance" | "security_only" | "end_of_support" | "closed";
  createdAt: string;
  updatedAt: string;
  closedAt?: string;
}

export interface CurrentState {
  branchLineId: string;
  rootDtProjectUuid?: string;
  rootBomSerialNumber?: string;
  rootBomVersion?: number;
  rootBomSha256?: string;
  sourceRevision?: string;
  updatedAt: string;
}

export interface Snapshot {
  id: string;
  branchLineId: string;
  snapshotType: "MAIN_SNAPSHOT" | "RELEASE";
  version: string;
  status: string;
  rootDtProjectUuid?: string;
  rootBomSerialNumber?: string;
  rootBomVersion?: number;
  rootBomSha256?: string;
  sourceRevision?: string;
  createdAt: string;
  releasedAt?: string;
}

export interface GraphNode {
  id: string;
  projectUUID: string;
  projectName: string;
  projectVersion: string;
  bomSerialNumber?: string;
  bomVersion?: number;
  resolutionStatus: "resolved" | "missing_project" | "missing_bom" | "missing_bom_ref" | "invalid";
}

export interface GraphEdge {
  sourceNodeId: string;
  targetNodeId: string;
  bomRef?: string;
  resolutionStatus: string;
}

export interface ReleaseGraph {
  rootNodeId: string;
  nodes: GraphNode[];
  edges: GraphEdge[];
  metadata: {
    totalNodes: number;
    totalEdges: number;
    maxDepthReached: boolean;
    maxNodesReached: boolean;
    unresolvedLinks: number;
    cyclesDetected: number;
  };
}

// --- API Functions ---
export const api = {
  // Products
  listProducts: () => request<Product[]>("/api/v1/products"),
  getProduct: (id: string) => request<Product>(`/api/v1/products/${id}`),
  createProduct: (data: { name: string; displayName?: string; description?: string }) =>
    request<Product>("/api/v1/products", { method: "POST", body: JSON.stringify(data) }),
  deleteProduct: (id: string) => request<void>(`/api/v1/products/${id}`, { method: "DELETE" }),

  // Branches
  listBranches: (productId: string) => request<BranchLine[]>(`/api/v1/products/${productId}/branches`),
  getBranch: (branchId: string) => request<BranchLine>(`/api/v1/branches/${branchId}`),
  createReleaseLine: (productId: string, data: { name: string; displayName?: string }) =>
    request<BranchLine>(`/api/v1/products/${productId}/release-lines`, { method: "POST", body: JSON.stringify(data) }),
  updateBranch: (branchId: string, data: { displayName?: string; status?: string }) =>
    request<BranchLine>(`/api/v1/branches/${branchId}`, { method: "PATCH", body: JSON.stringify(data) }),

  // Current State
  getCurrentState: (branchId: string) => request<CurrentState>(`/api/v1/branches/${branchId}/current`),
  updateCurrentState: (branchId: string, data: Partial<Omit<CurrentState, "branchLineId" | "updatedAt">>) =>
    request<CurrentState>(`/api/v1/branches/${branchId}/current`, { method: "PUT", body: JSON.stringify(data) }),

  // Releases
  listReleases: (branchId: string) => request<Snapshot[]>(`/api/v1/branches/${branchId}/releases`),
  getRelease: (releaseId: string) => request<Snapshot>(`/api/v1/releases/${releaseId}`),
  createRelease: (branchId: string, data: { version: string }) =>
    request<Snapshot>(`/api/v1/branches/${branchId}/releases`, { method: "POST", body: JSON.stringify(data) }),
  createSnapshot: (branchId: string, data: { version: string }) =>
    request<Snapshot>(`/api/v1/branches/${branchId}/snapshots`, { method: "POST", body: JSON.stringify(data) }),

  // Graph
  getBranchGraph: (branchId: string, maxDepth = 10, maxNodes = 1000) =>
    request<ReleaseGraph>(`/api/v1/branches/${branchId}/current/graph?maxDepth=${maxDepth}&maxNodes=${maxNodes}`),
  getReleaseGraph: (releaseId: string, maxDepth = 10, maxNodes = 1000) =>
    request<ReleaseGraph>(`/api/v1/releases/${releaseId}/graph?maxDepth=${maxDepth}&maxNodes=${maxNodes}`),
};
