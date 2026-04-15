export type TaskStatus = "Todo" | "InProgress" | "Done" | "Canceled";
export type TaskPriority = "Low" | "Medium" | "High" | "P0";
export type TaskSortBy = "DueAt";
export type SortDir = "Asc" | "Desc";

export type Task = {
  id: string;
  projectId?: string;
  name: string;
  description: string;
  context: string;
  details: string;
  status: TaskStatus;
  priority: TaskPriority;
  dueAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateTaskInput = {
  projectId?: string;
  name: string;
  description: string;
  context: string;
  details: string;
  priority: TaskPriority;
  // RFC3339 string, use "" to clear.
  dueAt?: string;
  status?: TaskStatus;
};

export type UpdateTaskInput = {
  projectId?: string;
  name: string;
  description: string;
  context: string;
  details: string;
  priority: TaskPriority;
  dueAt?: string;
};

export type ListTasksQuery = {
  projectId?: string;
  status?: TaskStatus;
  search?: string;
  sortBy?: TaskSortBy;
  sortDir?: SortDir;
  limit?: number;
  offset?: number;
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
    const msg =
      (data && typeof data.error === "string" && data.error) ||
      `${res.status} ${res.statusText}`;
    throw new Error(msg);
  }

  return data as T;
}

export async function listTasksByProject(
  projectId: string,
  q?: ListTasksQuery,
): Promise<Task[]> {
  const sp = new URLSearchParams();
  if (q?.status) sp.set("status", q.status);
  if (q?.search) sp.set("search", q.search);
  if (q?.sortBy) sp.set("sortBy", q.sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  return request<Task[]>(
    `/api/v1/projects/${encodeURIComponent(projectId)}/tasks${qs ? `?${qs}` : ""}`,
  );
}

export async function listTasks(q?: ListTasksQuery): Promise<Task[]> {
  const sp = new URLSearchParams();
  if (q?.projectId) sp.set("projectId", q.projectId);
  if (q?.status) sp.set("status", q.status);
  if (q?.search) sp.set("search", q.search);
  if (q?.sortBy) sp.set("sortBy", q.sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  return request<Task[]>(`/api/v1/tasks${qs ? `?${qs}` : ""}`);
}

export async function getTask(id: string): Promise<Task> {
  return request<Task>(`/api/v1/tasks/${encodeURIComponent(id)}`);
}

export async function createTask(input: CreateTaskInput): Promise<Task> {
  const body: Record<string, unknown> = {
    name: input.name,
    description: input.description,
    context: input.context,
    details: input.details,
    priority: input.priority,
  };
  if (input.projectId !== undefined) body.projectId = input.projectId;
  if (input.dueAt !== undefined) body.dueAt = input.dueAt;
  if (input.status) body.status = input.status;

  return request<Task>("/api/v1/tasks", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export async function updateTask(
  id: string,
  input: UpdateTaskInput,
): Promise<Task> {
  const body: Record<string, unknown> = {
    name: input.name,
    description: input.description,
    context: input.context,
    details: input.details,
    priority: input.priority,
  };
  if (input.projectId !== undefined) body.projectId = input.projectId;
  if (input.dueAt !== undefined) body.dueAt = input.dueAt;

  return request<Task>(`/api/v1/tasks/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export async function setTaskStatus(
  id: string,
  status: TaskStatus,
): Promise<Task> {
  return request<Task>(`/api/v1/tasks/${encodeURIComponent(id)}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

export async function deleteTask(id: string): Promise<void> {
  await request<void>(`/api/v1/tasks/${encodeURIComponent(id)}`, {
    method: "DELETE",
  });
}
