const API_BASE = import.meta.env.VITE_API_BASE_URL || "";

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
  forkedFromVersionId?: string;
  status: "active" | "maintenance" | "security_only" | "end_of_support" | "closed";
  createdAt: string;
  updatedAt: string;
  closedAt?: string;
}

export type Role = "ROOT" | "PROFILE" | "SUB";
export type VersionStatus = "incomplete" | "finalized";

export interface VersionDTProject {
  id: number;
  versionId: string;
  role: Role;
  dtProjectUuid?: string;
  bomSerialNumber?: string;
  bomVersion?: number;
  bomSha256?: string;
  sourceRevision?: string;
  label?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Version {
  id: string;
  branchLineId: string;
  versionString: string;
  status: VersionStatus;
  parentVersionId?: string;
  forkedFromVersionId?: string;
  location?: string;
  customer?: string;
  releaseDate?: string;
  createdAt: string;
  updatedAt: string;
  projects?: VersionDTProject[];
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

export type LineageAxis = "version" | "releaseDate";

export interface LineageNode {
  versionId: string;
  branchLineId: string;
  branchName: string;
  branchType: "MAIN" | "RELEASE";
  versionString: string;
  status: VersionStatus;
  releaseDate?: string;
  createdAt: string;
}

export interface LineageEdge {
  sourceVersionId: string;
  targetVersionId: string;
  kind: "parent" | "fork";
}

export interface LineageGraph {
  productId: string;
  axis: LineageAxis;
  nodes: LineageNode[];
  edges: LineageEdge[];
}

export interface TimelineEntry {
  versionId: string;
  branchLineId: string;
  branchName: string;
  branchType: "MAIN" | "RELEASE";
  versionString: string;
  status: VersionStatus;
  releaseDate?: string;
  createdAt: string;
}

export interface Timeline {
  entries: TimelineEntry[];
}

export interface DTProject {
  uuid: string;
  name: string;
  version: string;
}

// --- Input DTOs ---
export interface CreateVersionInput {
  versionString: string;
  forkedFromVersionId?: string;
}

export interface UpdateVersionInput {
  versionString?: string;
  status?: VersionStatus;
  location?: string;
  customer?: string;
  releaseDate?: string;
}

export interface SetProjectInput {
  dtProjectUuid?: string;
  bomSerialNumber?: string;
  bomVersion?: number;
  bomSha256?: string;
  sourceRevision?: string;
  label?: string;
}

export interface AddProjectInput extends SetProjectInput {
  role: "PROFILE" | "SUB";
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
  createReleaseLine: (productId: string, data: { name: string; displayName?: string; forkedFromVersionId?: string }) =>
    request<BranchLine>(`/api/v1/products/${productId}/release-lines`, { method: "POST", body: JSON.stringify(data) }),
  updateBranch: (branchId: string, data: { displayName?: string; status?: string }) =>
    request<BranchLine>(`/api/v1/branches/${branchId}`, { method: "PATCH", body: JSON.stringify(data) }),

  // Versions
  listVersions: (branchId: string) => request<Version[]>(`/api/v1/branches/${branchId}/versions`),
  getVersion: (versionId: string) => request<Version>(`/api/v1/versions/${versionId}`),
  createVersion: (branchId: string, data: CreateVersionInput) =>
    request<Version>(`/api/v1/branches/${branchId}/versions`, { method: "POST", body: JSON.stringify(data) }),
  updateVersion: (versionId: string, data: UpdateVersionInput) =>
    request<Version>(`/api/v1/versions/${versionId}`, { method: "PATCH", body: JSON.stringify(data) }),
  deleteVersion: (versionId: string) =>
    request<void>(`/api/v1/versions/${versionId}`, { method: "DELETE" }),

  // Role-tagged DT projects
  setVersionRoot: (versionId: string, data: SetProjectInput) =>
    request<VersionDTProject>(`/api/v1/versions/${versionId}/projects/root`, { method: "PUT", body: JSON.stringify(data) }),
  addVersionProject: (versionId: string, data: AddProjectInput) =>
    request<VersionDTProject>(`/api/v1/versions/${versionId}/projects`, { method: "POST", body: JSON.stringify(data) }),
  removeVersionProject: (versionId: string, bindingId: number) =>
    request<void>(`/api/v1/versions/${versionId}/projects/${bindingId}`, { method: "DELETE" }),

  // Graphs
  getProjectGraph: (versionId: string, role: Role, maxDepth = 10, maxNodes = 1000) =>
    request<ReleaseGraph>(`/api/v1/versions/${versionId}/projects/${role}/graph?maxDepth=${maxDepth}&maxNodes=${maxNodes}`),
  getProductLineage: (productId: string, axis: LineageAxis = "releaseDate") =>
    request<LineageGraph>(`/api/v1/products/${productId}/lineage-graph?axis=${axis}`),

  // Timelines
  getProductTimeline: (productId: string) => request<Timeline>(`/api/v1/products/${productId}/timeline`),
  getBranchTimeline: (branchId: string) => request<Timeline>(`/api/v1/branches/${branchId}/timeline`),

  // DT Projects
  searchDTProjects: (name?: string) =>
    request<DTProject[]>(name ? `/api/v1/dt/projects?name=${encodeURIComponent(name)}` : "/api/v1/dt/projects"),
};
