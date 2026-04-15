export type Someday = {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateSomedayInput = {
  name: string;
  description: string;
};

export type UpdateSomedayInput = CreateSomedayInput;

export type SomedaySortBy = "CreatedAt" | "UpdatedAt" | "Name";
export type SortDir = "Asc" | "Desc";

export type ListSomedaysQuery = {
  search?: string;
  sortBy?: SomedaySortBy;
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

export async function listSomedays(q?: ListSomedaysQuery): Promise<Someday[]> {
  const sp = new URLSearchParams();
  if (q?.search) sp.set("search", q.search);
  if (q?.sortBy) sp.set("sortBy", q.sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  return request<Someday[]>(`/api/v1/somedays${qs ? `?${qs}` : ""}`);
}

export async function getSomeday(id: string): Promise<Someday> {
  return request<Someday>(`/api/v1/somedays/${encodeURIComponent(id)}`);
}

export async function createSomeday(input: CreateSomedayInput): Promise<Someday> {
  return request<Someday>("/api/v1/somedays", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateSomeday(id: string, input: UpdateSomedayInput): Promise<Someday> {
  return request<Someday>(`/api/v1/somedays/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteSomeday(id: string): Promise<void> {
  await request<void>(`/api/v1/somedays/${encodeURIComponent(id)}`, { method: "DELETE" });
}
