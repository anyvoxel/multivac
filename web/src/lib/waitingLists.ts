import {
  createAction,
  deleteAction,
  getAction,
  listActions,
  updateAction,
} from "./actions";
import type {
  Action,
  ActionAttributes,
  ActionSortBy,
} from "./actions";
import type { SortDir } from "./items";

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

function mapSortBy(sortBy?: WaitingListSortBy): ActionSortBy | undefined {
  switch (sortBy) {
    case "Name":
      return "Name";
    case "CreatedAt":
      return "CreatedAt";
    case "UpdatedAt":
      return "UpdatedAt";
    default:
      return undefined;
  }
}

function toWaitingAttributes(owner: string, expectedAt?: string): ActionAttributes {
  return {
    waiting: {
      delegatee: owner,
      due_at: expectedAt,
    },
  };
}

function fromAction(action: Action): WaitingList {
  return {
    id: action.id,
    name: action.title,
    details: action.description,
    owner: action.attributes.waiting?.delegatee ?? "",
    expectedAt: action.attributes.waiting?.due_at,
    createdAt: action.createdAt,
    updatedAt: action.updatedAt,
  };
}

function expectedAtTimestamp(value?: string): number | null {
  if (!value) return null;
  const timestamp = Date.parse(value);
  return Number.isNaN(timestamp) ? null : timestamp;
}

function sortByExpectedAt(list: WaitingList[], dir: SortDir): WaitingList[] {
  return list
    .map((item, index) => ({ item, index, timestamp: expectedAtTimestamp(item.expectedAt) }))
    .sort((a, b) => {
      const aNil = a.timestamp === null;
      const bNil = b.timestamp === null;
      if (aNil && bNil) return a.index - b.index;
      if (aNil) return 1;
      if (bNil) return -1;
      const left = a.timestamp as number;
      const right = b.timestamp as number;
      const delta = left - right;
      if (delta !== 0) {
        return dir === "Desc" ? -delta : delta;
      }
      return a.index - b.index;
    })
    .map((x) => x.item);
}

export async function listWaitingLists(q?: ListWaitingListsQuery): Promise<WaitingList[]> {
  const sortByExpected = q?.sortBy === "ExpectedAt";
  const actions = await listActions({
    search: q?.search,
    kind: "Waiting",
    sortBy: sortByExpected ? undefined : mapSortBy(q?.sortBy),
    sortDir: q?.sortDir,
    limit: sortByExpected ? undefined : q?.limit,
    offset: sortByExpected ? undefined : q?.offset,
  });

  let list = actions.filter((action) => action.kind === "Waiting").map(fromAction);

  if (!sortByExpected) {
    return list;
  }

  list = sortByExpectedAt(list, q?.sortDir ?? "Asc");
  const offset = q?.offset ?? 0;
  if (offset > 0) {
    list = list.slice(offset);
  }
  if (q?.limit !== undefined) {
    list = list.slice(0, q.limit);
  }
  return list;
}

export async function getWaitingList(id: string): Promise<WaitingList> {
  return fromAction(await getAction(id));
}

export async function createWaitingList(input: CreateWaitingListInput): Promise<WaitingList> {
  return fromAction(
    await createAction({
      title: input.name,
      description: input.details,
      kind: "Waiting",
      context: [],
      labels: [],
      attributes: toWaitingAttributes(input.owner, input.expectedAt),
    }),
  );
}

export async function updateWaitingList(id: string, input: UpdateWaitingListInput): Promise<WaitingList> {
  const current = await getAction(id);
  return fromAction(
    await updateAction(id, {
      title: input.name,
      description: input.details,
      project_id: current.project_id,
      kind: "Waiting",
      context: current.context,
      labels: current.labels,
      attributes: toWaitingAttributes(input.owner, input.expectedAt),
    }),
  );
}

export async function deleteWaitingList(id: string): Promise<void> {
  await deleteAction(id);
}
