import {
  createItem,
  deleteItem,
  getItem,
  listItems,
  updateItem,
} from "./items";
import type { Item, ItemBucket, ItemSortBy, SortDir } from "./items";

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

export type ListWaitingListsQuery = {
  search?: string;
  sortBy?: WaitingListSortBy;
  sortDir?: SortDir;
  limit?: number;
  offset?: number;
};

function mapSortBy(sortBy?: WaitingListSortBy): ItemSortBy | undefined {
  switch (sortBy) {
    case "Name":
      return "Title";
    case "CreatedAt":
      return "CreatedAt";
    case "UpdatedAt":
      return "UpdatedAt";
    case "ExpectedAt":
      return "ExpectedAt";
    default:
      return undefined;
  }
}

function fromItem(item: Item): WaitingList {
  return {
    id: item.id,
    name: item.title,
    details: item.details,
    owner: item.waitingFor ?? "",
    expectedAt: item.expectedAt,
    createdAt: item.createdAt,
    updatedAt: item.updatedAt,
  };
}

function waitingListBucket(): ItemBucket {
  return "WaitingFor";
}

export async function listWaitingLists(q?: ListWaitingListsQuery): Promise<WaitingList[]> {
  const items = await listItems({
    kind: "WaitingFor",
    bucket: waitingListBucket(),
    search: q?.search,
    sortBy: mapSortBy(q?.sortBy),
    sortDir: q?.sortDir,
    limit: q?.limit,
    offset: q?.offset,
  });
  return items.map(fromItem);
}

export async function getWaitingList(id: string): Promise<WaitingList> {
  return fromItem(await getItem(id));
}

export async function createWaitingList(input: CreateWaitingListInput): Promise<WaitingList> {
  return fromItem(
    await createItem({
      kind: "WaitingFor",
      bucket: waitingListBucket(),
      title: input.name,
      description: "",
      labels: [],
      context: "",
      details: input.details,
      waitingFor: input.owner,
      expectedAt: input.expectedAt,
    }),
  );
}

export async function updateWaitingList(id: string, input: UpdateWaitingListInput): Promise<WaitingList> {
  const current = await getItem(id);
  return fromItem(
    await updateItem(id, {
      kind: current.kind,
      bucket: current.bucket,
      projectId: current.projectId,
      title: input.name,
      description: current.description,
      labels: current.labels,
      context: current.context,
      details: input.details,
      taskStatus: current.taskStatus,
      priority: current.priority,
      waitingFor: input.owner,
      expectedAt: input.expectedAt,
      dueAt: current.dueAt,
    }),
  );
}

export async function deleteWaitingList(id: string): Promise<void> {
  await deleteItem(id);
}
