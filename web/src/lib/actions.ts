import type { SortDir } from "./items";

export type ActionKind = "Task" | "Waiting" | "Scheduled";

export type ActionLabel = {
  name: string;
};

export type ActionTaskStatus = "Pending" | "Active" | "Completed";

export type ActionTaskAttributes = {
  expectedAt?: string;
  status?: ActionTaskStatus;
};

export type ActionWaitingAttributes = {
  delegatee: string;
  dueAt?: string;
};

export type ActionScheduledAttributes = {
  startAt?: string;
  endAt?: string;
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
  projectId?: string;
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
  contextIds?: string[];
  sortBy?: ActionSortBy;
  sortDir?: SortDir;
  limit?: number;
  offset?: number;
};

export type CreateActionInput = {
  title: string;
  description: string;
  projectId?: string;
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
    projectId: input.projectId,
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
  if (q?.contextIds?.length) sp.set("contextIds", q.contextIds.join(","));
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

export type ConvertInboxToActionInput = {
  title?: string;
  description?: string;
  projectId?: string;
  kind?: ActionKind;
  context?: string[];
  labels?: ActionLabel[];
  attributes?: ActionAttributes;
};

export async function convertInboxToAction(id: string, input?: ConvertInboxToActionInput): Promise<Action> {
  const body = input ? JSON.stringify({
    ...(input.title !== undefined ? { title: input.title } : {}),
    ...(input.description !== undefined ? { description: input.description } : {}),
    ...(input.projectId !== undefined ? { projectId: input.projectId } : {}),
    ...(input.kind !== undefined ? { kind: input.kind } : {}),
    ...(input.context !== undefined ? { context: input.context } : {}),
    ...(input.labels !== undefined ? { labels: input.labels } : {}),
    ...(input.attributes !== undefined ? { attributes: input.attributes } : {}),
  }) : undefined;
  return request<Action>(`/api/v1/inboxes/${encodeURIComponent(id)}/convert-to-action`, {
    method: "POST",
    ...(body ? { body } : {}),
  });
}

export type ConvertActionKindInput = {
  kind: ActionKind;
  attributes: ActionAttributes;
};

export async function convertActionKind(id: string, input: ConvertActionKindInput): Promise<Action> {
  return request<Action>(`/api/v1/actions/${encodeURIComponent(id)}/convert`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}
