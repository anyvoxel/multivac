import {
  createItem,
  deleteItem,
  getItem,
  listItems,
  updateItem,
} from "./items";
import type { Item, ItemBucket, ItemSortBy, SortDir } from "./items";

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

export type ListSomedaysQuery = {
  search?: string;
  sortBy?: SomedaySortBy;
  sortDir?: SortDir;
  limit?: number;
  offset?: number;
};

function mapSortBy(sortBy?: SomedaySortBy): ItemSortBy | undefined {
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

function fromItem(item: Item): Someday {
  return {
    id: item.id,
    name: item.title,
    description: item.description,
    createdAt: item.createdAt,
    updatedAt: item.updatedAt,
  };
}

function somedayBucket(): ItemBucket {
  return "SomedayMaybe";
}

export async function listSomedays(q?: ListSomedaysQuery): Promise<Someday[]> {
  const items = await listItems({
    bucket: somedayBucket(),
    search: q?.search,
    sortBy: mapSortBy(q?.sortBy),
    sortDir: q?.sortDir,
    limit: q?.limit,
    offset: q?.offset,
  });
  return items.map(fromItem);
}

export async function getSomeday(id: string): Promise<Someday> {
  return fromItem(await getItem(id));
}

export async function createSomeday(input: CreateSomedayInput): Promise<Someday> {
  return fromItem(
    await createItem({
      kind: "SomedayMaybe",
      bucket: somedayBucket(),
      title: input.name,
      description: input.description,
      context: "",
      details: "",
    }),
  );
}

export async function updateSomeday(id: string, input: UpdateSomedayInput): Promise<Someday> {
  const current = await getItem(id);
  return fromItem(
    await updateItem(id, {
      kind: current.kind,
      bucket: current.bucket,
      projectId: current.projectId,
      title: input.name,
      description: input.description,
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

export async function deleteSomeday(id: string): Promise<void> {
  await deleteItem(id);
}
