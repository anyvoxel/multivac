import type { SortDir } from "./items";

export type ActionKind = "Task" | "Waiting" | "Scheduled";

export type ActionLabel = {
  name: string;
};

export type ActionTaskAttributes = {
  expected_at?: string;
};

export type ActionWaitingAttributes = {
  delegatee: string;
  due_at?: string;
};

export type ActionScheduledAttributes = {
  start_at?: string;
  end_at?: string;
};

export type ActionAttributes = {
  task?: ActionTaskAttributes;
  waiting?: ActionWaitingAttributes;
  scheduled?: ActionScheduledAttributes;
};

export type Action = {
  id: string;
  title: string;
  description: string;
  project_id?: string;
  kind: ActionKind;
  context: string[];
  labels: ActionLabel[];
  attributes: ActionAttributes;
  createdAt: string;
  updatedAt: string;
};

export type ActionSortBy = "CreatedAt" | "UpdatedAt" | "Name";

export type ListActionsQuery = {
  search?: string;
  kind?: ActionKind;
  projectId?: string;
  sortBy?: ActionSortBy;
  sortDir?: SortDir;
  limit?: number;
  offset?: number;
};

export type CreateActionInput = {
  title: string;
  description: string;
  project_id?: string;
  kind: ActionKind;
  context?: string[];
  labels?: ActionLabel[];
  attributes: ActionAttributes;
};

export type UpdateActionInput = CreateActionInput;

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

function toApiSortBy(sortBy?: ActionSortBy): string | undefined {
  switch (sortBy) {
    case "Name":
      return "Name";
    case "CreatedAt":
      return "CreatedAt";
    case "UpdatedAt":
      return "UpdatedAt";
    default:
      return undefined;
  }
}

function toAction(input: CreateActionInput | UpdateActionInput): Omit<Action, "id" | "createdAt" | "updatedAt"> {
  return {
    title: input.title,
    description: input.description,
    project_id: input.project_id,
    kind: input.kind,
    context: input.context ?? [],
    labels: input.labels ?? [],
    attributes: input.attributes,
  };
}

export async function listActions(q?: ListActionsQuery): Promise<Action[]> {
  const sp = new URLSearchParams();
  if (q?.search) sp.set("search", q.search);
  if (q?.kind) sp.set("kind", q.kind);
  if (q?.projectId) sp.set("projectId", q.projectId);
  const sortBy = toApiSortBy(q?.sortBy);
  if (sortBy) sp.set("sortBy", sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  return request<Action[]>(`/api/v1/actions${qs ? `?${qs}` : ""}`);
}

export async function getAction(id: string): Promise<Action> {
  return request<Action>(`/api/v1/actions/${encodeURIComponent(id)}`);
}

export async function createAction(input: CreateActionInput): Promise<Action> {
  return request<Action>("/api/v1/actions", {
    method: "POST",
    body: JSON.stringify(toAction(input)),
  });
}

export async function updateAction(id: string, input: UpdateActionInput): Promise<Action> {
  return request<Action>(`/api/v1/actions/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(toAction(input)),
  });
}

export async function deleteAction(id: string): Promise<void> {
  await request<void>(`/api/v1/actions/${encodeURIComponent(id)}`, { method: "DELETE" });
}
