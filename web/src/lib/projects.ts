export type ProjectStatus = "Draft" | "Active" | "Completed" | "Archived";

export type Project = {
  id: string;
  name: string;
  goal: string;
  principles: string;
  visionResult: string;
  description: string;
  status: ProjectStatus;
  startedAt?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateProjectInput = {
  name: string;
  goal: string;
  principles: string;
  visionResult: string;
  description: string;
};

export type UpdateProjectInput = CreateProjectInput;

export type ListProjectsQuery = {
  status?: ProjectStatus;
};

function apiBase(): string {
  const v = import.meta.env.VITE_API_BASE as string | undefined;
  if (v === undefined) return "";
  return v;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${apiBase()}${path}`, {
    ...init,
    headers: {
      "content-type": "application/json",
      ...(init?.headers ?? {}),
    },
  });

  if (res.status === 204) {
    return undefined as unknown as T;
  }

  const text = await res.text();
  const data = text ? JSON.parse(text) : undefined;

  if (!res.ok) {
    const msg = (data && typeof data.error === "string" && data.error) || `${res.status} ${res.statusText}`;
    throw new Error(msg);
  }

  return data as T;
}

export async function listProjects(q?: ListProjectsQuery): Promise<Project[]> {
  const sp = new URLSearchParams();
  if (q?.status) sp.set("status", q.status);
  const qs = sp.toString();
  return request<Project[]>(`/api/v1/projects${qs ? `?${qs}` : ""}`);
}

export async function getProject(id: string): Promise<Project> {
  return request<Project>(`/api/v1/projects/${encodeURIComponent(id)}`);
}

export async function createProject(input: CreateProjectInput): Promise<Project> {
  return request<Project>("/api/v1/projects", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateProject(id: string, input: UpdateProjectInput): Promise<Project> {
  return request<Project>(`/api/v1/projects/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function setProjectStatus(id: string, status: ProjectStatus): Promise<Project> {
  return request<Project>(`/api/v1/projects/${encodeURIComponent(id)}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

export async function deleteProject(id: string): Promise<void> {
  await request<void>(`/api/v1/projects/${encodeURIComponent(id)}`, { method: "DELETE" });
}
