export type SortDir = "Asc" | "Desc";

export type ReferenceLink = {
  title: string;
  url: string;
};

export type Reference = {
  id: string;
  title: string;
  description: string;
  references: ReferenceLink[];
  createdAt: string;
  updatedAt: string;
};

export type CreateReferenceInput = {
  title: string;
  description: string;
  references: ReferenceLink[];
};

export type UpdateReferenceInput = CreateReferenceInput;

export type ReferenceSortBy = "CreatedAt" | "UpdatedAt" | "Title";

export type ListReferencesQuery = {
  search?: string;
  sortBy?: ReferenceSortBy;
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

export async function listReferences(q?: ListReferencesQuery): Promise<Reference[]> {
  const sp = new URLSearchParams();
  if (q?.search) sp.set("search", q.search);
  if (q?.sortBy) sp.set("sortBy", q.sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  return request<Reference[]>(`/api/v1/references${qs ? `?${qs}` : ""}`);
}

export async function getReference(id: string): Promise<Reference> {
  return request<Reference>(`/api/v1/references/${encodeURIComponent(id)}`);
}

export async function createReference(input: CreateReferenceInput): Promise<Reference> {
  return request<Reference>("/api/v1/references", { method: "POST", body: JSON.stringify(input) });
}

export async function updateReference(id: string, input: UpdateReferenceInput): Promise<Reference> {
  return request<Reference>(`/api/v1/references/${encodeURIComponent(id)}`, { method: "PUT", body: JSON.stringify(input) });
}

export async function deleteReference(id: string): Promise<void> {
  await request<void>(`/api/v1/references/${encodeURIComponent(id)}`, { method: "DELETE" });
}
