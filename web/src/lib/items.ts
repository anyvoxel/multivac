export type ItemKind = "Inbox" | "Task" | "WaitingFor" | "SomedayMaybe";
export type ItemBucket = "Inbox" | "NextAction" | "WaitingFor" | "SomedayMaybe" | "Completed" | "Dropped";
export type SortDir = "Asc" | "Desc";
export type ItemSortBy = "CreatedAt" | "UpdatedAt" | "Title" | "DueAt" | "ExpectedAt" | "Priority";

export type Item = {
  id: string;
  kind: ItemKind;
  bucket: ItemBucket;
  projectId?: string;
  title: string;
  description: string;
  context: string;
  details: string;
  taskStatus?: string;
  priority?: string;
  waitingFor?: string;
  expectedAt?: string;
  dueAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type ListItemsQuery = {
  bucket?: ItemBucket;
  kind?: ItemKind;
  projectId?: string;
  taskStatus?: string;
  search?: string;
  sortBy?: ItemSortBy;
  sortDir?: SortDir;
  limit?: number;
  offset?: number;
};

export type CreateItemInput = {
  kind: ItemKind;
  bucket: ItemBucket;
  projectId?: string;
  title: string;
  description: string;
  context: string;
  details: string;
  taskStatus?: string;
  priority?: string;
  waitingFor?: string;
  expectedAt?: string;
  dueAt?: string;
};

export type UpdateItemInput = CreateItemInput;

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

export async function listItems(q?: ListItemsQuery): Promise<Item[]> {
  const sp = new URLSearchParams();
  if (q?.bucket) sp.set("bucket", q.bucket);
  if (q?.kind) sp.set("kind", q.kind);
  if (q?.projectId) sp.set("projectId", q.projectId);
  if (q?.taskStatus) sp.set("taskStatus", q.taskStatus);
  if (q?.search) sp.set("search", q.search);
  if (q?.sortBy) sp.set("sortBy", q.sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  return request<Item[]>(`/api/v1/items${qs ? `?${qs}` : ""}`);
}

export async function getItem(id: string): Promise<Item> {
  return request<Item>(`/api/v1/items/${encodeURIComponent(id)}`);
}

export async function createItem(input: CreateItemInput): Promise<Item> {
  return request<Item>("/api/v1/items", { method: "POST", body: JSON.stringify(input) });
}

export async function updateItem(id: string, input: UpdateItemInput): Promise<Item> {
  return request<Item>(`/api/v1/items/${encodeURIComponent(id)}`, { method: "PUT", body: JSON.stringify(input) });
}

export async function moveItemBucket(id: string, bucket: ItemBucket): Promise<Item> {
  return request<Item>(`/api/v1/items/${encodeURIComponent(id)}/bucket`, { method: "PATCH", body: JSON.stringify({ bucket }) });
}

export async function deleteItem(id: string): Promise<void> {
  await request<void>(`/api/v1/items/${encodeURIComponent(id)}`, { method: "DELETE" });
}
