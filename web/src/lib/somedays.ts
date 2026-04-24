export type SortDir = "Asc" | "Desc";

export type Someday = {
  id: string;
  name: string;
  description: string;
  createdAt: string;
  updatedAt: string;
};

type ApiSomeday = {
  id: string;
  title: string;
  description: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateSomedayInput = {
  name: string;
  description: string;
};

export type UpdateSomedayInput = CreateSomedayInput;

export type ConvertInboxToSomedayInput = {
  name?: string;
  description?: string;
};

export type ConvertActionToSomedayInput = {
  name?: string;
  description?: string;
};

export type SomedaySortBy = "CreatedAt" | "UpdatedAt" | "Name";

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

function toApiSortBy(sortBy?: SomedaySortBy): string | undefined {
  switch (sortBy) {
    case "Name":
      return "Title";
    case "CreatedAt":
      return "CreatedAt";
    case "UpdatedAt":
      return "UpdatedAt";
    default:
      return undefined;
  }
}

function fromApiSomeday(someday: ApiSomeday): Someday {
  return {
    id: someday.id,
    name: someday.title,
    description: someday.description,
    createdAt: someday.createdAt,
    updatedAt: someday.updatedAt,
  };
}

export async function listSomedays(q?: ListSomedaysQuery): Promise<Someday[]> {
  const sp = new URLSearchParams();
  if (q?.search) sp.set("search", q.search);
  const sortBy = toApiSortBy(q?.sortBy);
  if (sortBy) sp.set("sortBy", sortBy);
  if (q?.sortDir) sp.set("sortDir", q.sortDir);
  if (q?.limit !== undefined) sp.set("limit", String(q.limit));
  if (q?.offset !== undefined) sp.set("offset", String(q.offset));
  const qs = sp.toString();
  const list = await request<ApiSomeday[]>(`/api/v1/somedays${qs ? `?${qs}` : ""}`);
  return list.map(fromApiSomeday);
}

export async function getSomeday(id: string): Promise<Someday> {
  return fromApiSomeday(await request<ApiSomeday>(`/api/v1/somedays/${encodeURIComponent(id)}`));
}

export async function createSomeday(input: CreateSomedayInput): Promise<Someday> {
  return fromApiSomeday(
    await request<ApiSomeday>("/api/v1/somedays", {
      method: "POST",
      body: JSON.stringify({ title: input.name, description: input.description }),
    }),
  );
}

export async function updateSomeday(id: string, input: UpdateSomedayInput): Promise<Someday> {
  return fromApiSomeday(
    await request<ApiSomeday>(`/api/v1/somedays/${encodeURIComponent(id)}`, {
      method: "PUT",
      body: JSON.stringify({ title: input.name, description: input.description }),
    }),
  );
}

export async function convertInboxToSomeday(id: string, input?: ConvertInboxToSomedayInput): Promise<Someday> {
  const body = input
    ? JSON.stringify({
        ...(input.name !== undefined ? { title: input.name } : {}),
        ...(input.description !== undefined ? { description: input.description } : {}),
      })
    : undefined;
  return fromApiSomeday(
    await request<ApiSomeday>(`/api/v1/inboxes/${encodeURIComponent(id)}/convert-to-someday`, {
      method: "POST",
      ...(body ? { body } : {}),
    }),
  );
}

export async function convertActionToSomeday(id: string, input?: ConvertActionToSomedayInput): Promise<Someday> {
  const body = input
    ? JSON.stringify({
        ...(input.name !== undefined ? { title: input.name } : {}),
        ...(input.description !== undefined ? { description: input.description } : {}),
      })
    : undefined;
  return fromApiSomeday(
    await request<ApiSomeday>(`/api/v1/actions/${encodeURIComponent(id)}/convert-to-someday`, {
      method: "POST",
      ...(body ? { body } : {}),
    }),
  );
}

export async function deleteSomeday(id: string): Promise<void> {
  await request<void>(`/api/v1/somedays/${encodeURIComponent(id)}`, { method: "DELETE" });
}
