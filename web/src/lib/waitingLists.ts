export type WaitingList = {
  id: string;
  name: string;
  details: string;
  owner: string;
  expectedAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateWaitingListInput = {
  name: string;
  details: string;
  owner: string;
  expectedAt?: string;
};

export type UpdateWaitingListInput = CreateWaitingListInput;

export type WaitingListSortBy = "CreatedAt" | "UpdatedAt" | "Name" | "ExpectedAt";
export type SortDir = "Asc" | "Desc";

export type ListWaitingListsQuery = {
  search?: string;
  sortBy?: WaitingListSortBy;
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

export async function listWaitingLists(q?: ListWaitingListsQuery): Promise<WaitingList[]> {
  const sp = new URLSearchParams();
  if (q?.search) sp.set("search", q.search);
  if (q?.sortBy) sp.set("sortBy", q.sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  return request<WaitingList[]>(`/api/v1/waiting-lists${qs ? `?${qs}` : ""}`);
}

export async function getWaitingList(id: string): Promise<WaitingList> {
  return request<WaitingList>(`/api/v1/waiting-lists/${encodeURIComponent(id)}`);
}

export async function createWaitingList(input: CreateWaitingListInput): Promise<WaitingList> {
  return request<WaitingList>("/api/v1/waiting-lists", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateWaitingList(id: string, input: UpdateWaitingListInput): Promise<WaitingList> {
  return request<WaitingList>(`/api/v1/waiting-lists/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteWaitingList(id: string): Promise<void> {
  await request<void>(`/api/v1/waiting-lists/${encodeURIComponent(id)}`, { method: "DELETE" });
}
