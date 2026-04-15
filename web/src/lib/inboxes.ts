export type Inbox = {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateInboxInput = {
  name: string;
  description: string;
};

export type UpdateInboxInput = CreateInboxInput;

export type InboxSortBy = "CreatedAt" | "UpdatedAt" | "Name";
export type SortDir = "Asc" | "Desc";

export type ListInboxesQuery = {
  search?: string;
  sortBy?: InboxSortBy;
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

export async function listInboxes(q?: ListInboxesQuery): Promise<Inbox[]> {
  const sp = new URLSearchParams();
  if (q?.search) sp.set("search", q.search);
  if (q?.sortBy) sp.set("sortBy", q.sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  return request<Inbox[]>(`/api/v1/inboxes${qs ? `?${qs}` : ""}`);
}

export async function getInbox(id: string): Promise<Inbox> {
  return request<Inbox>(`/api/v1/inboxes/${encodeURIComponent(id)}`);
}

export async function createInbox(input: CreateInboxInput): Promise<Inbox> {
  return request<Inbox>("/api/v1/inboxes", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function updateInbox(id: string, input: UpdateInboxInput): Promise<Inbox> {
  return request<Inbox>(`/api/v1/inboxes/${encodeURIComponent(id)}`, {
    method: "PUT",
    body: JSON.stringify(input),
  });
}

export async function deleteInbox(id: string): Promise<void> {
  await request<void>(`/api/v1/inboxes/${encodeURIComponent(id)}`, { method: "DELETE" });
}
