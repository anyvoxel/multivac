export type SortDir = "Asc" | "Desc";

export type Context = {
  id: string;
  title: string;
  description: string;
  color: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateContextInput = {
  title: string;
  description: string;
  color: string;
};

export type UpdateContextInput = CreateContextInput;

export type ContextSortBy = "CreatedAt" | "UpdatedAt" | "Title";

export type ListContextsQuery = {
  search?: string;
  sortBy?: ContextSortBy;
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
    const msg = (data && typeof data.error === "string" && data.error) || `${res.status} ${res.statusText}`;
    throw new Error(msg);
  }

  return data as T;
}

export async function listContexts(q?: ListContextsQuery): Promise<Context[]> {
  const sp = new URLSearchParams();
  if (q?.search) sp.set("search", q.search);
  if (q?.sortBy) sp.set("sortBy", q.sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  return request<Context[]>(`/api/v1/contexts${qs ? `?${qs}` : ""}`);
}

export async function getContext(id: string): Promise<Context> {
  return request<Context>(`/api/v1/contexts/${encodeURIComponent(id)}`);
}

export async function createContext(input: CreateContextInput): Promise<Context> {
  return request<Context>("/api/v1/contexts", { method: "POST", body: JSON.stringify(input) });
}

export async function updateContext(id: string, input: UpdateContextInput): Promise<Context> {
  return request<Context>(`/api/v1/contexts/${encodeURIComponent(id)}`, { method: "PUT", body: JSON.stringify(input) });
}

export async function deleteContext(id: string): Promise<void> {
  await request<void>(`/api/v1/contexts/${encodeURIComponent(id)}`, { method: "DELETE" });
}
