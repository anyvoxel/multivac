export type ProjectStatus = "Draft" | "Active" | "Completed" | "Hold";

export type ProjectReference = {
  title: string;
  URL: string;
};

export type LabelKind = "Context" | "Tag";

export type Label = {
  value: string;
  kind: LabelKind;
  filterable: boolean;
};

export type Goal = {
  title: string;
  createdAt: string;
  completedAt?: string;
};

export type Project = {
  id: string;
  title: string;
  goals: Goal[];
  description: string;
  references: ProjectReference[];
  status: ProjectStatus;
  startAt?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateProjectInput = {
  title: string;
  goals: Goal[];
  description: string;
  references: ProjectReference[];
};

export type UpdateProjectInput = CreateProjectInput;

export type ListProjectsQuery = {
  status?: ProjectStatus;
  search?: string;
  limit?: number;
  offset?: number;
};

type ApiGoal = {
  title: string;
  created_at: string;
  completed_at?: string;
};

type ApiReference = {
  title: string;
  URL: string;
};

type ApiProject = {
  id: string;
  title: string;
  goals: ApiGoal[];
  description: string;
  references: ApiReference[];
  status: ProjectStatus;
  startAt?: string;
  completedAt?: string;
  createdAt: string;
  updatedAt: string;
};

type ApiCreateProjectInput = {
  title: string;
  goals: ApiGoal[];
  description: string;
  references: ApiReference[];
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

function fromApiGoal(goal: ApiGoal): Goal {
  return {
    title: goal.title,
    createdAt: goal.created_at,
    completedAt: goal.completed_at,
  };
}

function toApiGoal(goal: Goal): ApiGoal {
  return {
    title: goal.title,
    created_at: goal.createdAt,
    completed_at: goal.completedAt,
  };
}

function fromApiProject(project: ApiProject): Project {
  return {
    id: project.id,
    title: project.title,
    goals: project.goals.map(fromApiGoal),
    description: project.description,
    references: project.references,
    status: project.status,
    startAt: project.startAt,
    completedAt: project.completedAt,
    createdAt: project.createdAt,
    updatedAt: project.updatedAt,
  };
}

function toApiProjectInput(input: CreateProjectInput | UpdateProjectInput): ApiCreateProjectInput {
  return {
    title: input.title,
    goals: input.goals.map(toApiGoal),
    description: input.description,
    references: input.references,
  };
}

export async function listProjects(q?: ListProjectsQuery): Promise<Project[]> {
  const sp = new URLSearchParams();
  if (q?.status) sp.set("status", q.status);
  if (q?.search) sp.set("search", q.search);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  const list = await request<ApiProject[]>(`/api/v1/projects${qs ? `?${qs}` : ""}`);
  return list.map(fromApiProject);
}

export async function getProject(id: string): Promise<Project> {
  const project = await request<ApiProject>(`/api/v1/projects/${encodeURIComponent(id)}`);
  return fromApiProject(project);
}

export async function createProject(input: CreateProjectInput): Promise<Project> {
  const project = await request<ApiProject>("/api/v1/projects", {
    method: "POST",
    body: JSON.stringify(toApiProjectInput(input)),
  });
  return fromApiProject(project);
}

export async function updateProject(id: string, input: UpdateProjectInput): Promise<Project> {
  const project = await request<ApiProject>(`/api/v1/projects/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(toApiProjectInput(input)),
  });
  return fromApiProject(project);
}

export async function setProjectStatus(id: string, status: ProjectStatus): Promise<Project> {
  const project = await request<ApiProject>(`/api/v1/projects/${encodeURIComponent(id)}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
  return fromApiProject(project);
}

export async function deleteProject(id: string): Promise<void> {
  await request<void>(`/api/v1/projects/${encodeURIComponent(id)}`, { method: "DELETE" });
}
