import {
  createItem,
  deleteItem,
  getItem,
  listItems,
  updateItem,
} from "./items";
import type { Item, ItemBucket, ItemSortBy, SortDir } from "./items";

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

export type ListInboxesQuery = {
  search?: string;
  sortBy?: InboxSortBy;
  sortDir?: SortDir;
  limit?: number;
  offset?: number;
};

function mapSortBy(sortBy?: InboxSortBy): ItemSortBy | undefined {
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

function fromItem(item: Item): Inbox {
  return {
    id: item.id,
    name: item.title,
    description: item.description,
    createdAt: item.createdAt,
    updatedAt: item.updatedAt,
  };
}

function inboxBucket(): ItemBucket {
  return "Inbox";
}

export async function listInboxes(q?: ListInboxesQuery): Promise<Inbox[]> {
  const items = await listItems({
    bucket: inboxBucket(),
    search: q?.search,
    sortBy: mapSortBy(q?.sortBy),
    sortDir: q?.sortDir,
    limit: q?.limit,
    offset: q?.offset,
  });
  return items.map(fromItem);
}

export async function getInbox(id: string): Promise<Inbox> {
  return fromItem(await getItem(id));
}

export async function createInbox(input: CreateInboxInput): Promise<Inbox> {
  return fromItem(
    await createItem({
      kind: "Inbox",
      bucket: inboxBucket(),
      title: input.name,
      description: input.description,
      labels: [],
      context: "",
      details: "",
    }),
  );
}

export async function updateInbox(id: string, input: UpdateInboxInput): Promise<Inbox> {
  const current = await getItem(id);
  return fromItem(
    await updateItem(id, {
      kind: current.kind,
      bucket: current.bucket,
      projectId: current.projectId,
      title: input.name,
      description: input.description,
      labels: current.labels,
      context: current.context,
      details: current.details,
      taskStatus: current.taskStatus,
      priority: current.priority,
      waitingFor: current.waitingFor,
      expectedAt: current.expectedAt,
      dueAt: current.dueAt,
    }),
  );
}

export async function deleteInbox(id: string): Promise<void> {
  await deleteItem(id);
}
