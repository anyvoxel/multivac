import {
  createItem,
  deleteItem,
  getItem,
  listItems,
  updateItem,
} from "./items";
import type { Item, ItemBucket, ItemSortBy, SortDir } from "./items";

export type TaskStatus = "Todo" | "InProgress" | "Done" | "Canceled";
export type TaskPriority = "Low" | "Medium" | "High" | "P0";
export type TaskSortBy = "DueAt";
export type { SortDir };

export type Task = {
  id: string;
  projectId?: string;
  name: string;
  description: string;
  context: string;
  details: string;
  status: TaskStatus;
  priority: TaskPriority;
  dueAt?: string;
  createdAt: string;
  updatedAt: string;
};

export type CreateTaskInput = {
  projectId?: string;
  name: string;
  description: string;
  context: string;
  details: string;
  priority: TaskPriority;
  // RFC3339 string, use "" to clear.
  dueAt?: string;
  status?: TaskStatus;
};

export type UpdateTaskInput = {
  projectId?: string;
  name: string;
  description: string;
  context: string;
  details: string;
  priority: TaskPriority;
  dueAt?: string;
};

export type ListTasksQuery = {
  projectId?: string;
  status?: TaskStatus;
  search?: string;
  sortBy?: TaskSortBy;
  sortDir?: SortDir;
  limit?: number;
  offset?: number;
};

function fromItem(item: Item): Task {
  return {
    id: item.id,
    projectId: item.projectId,
    name: item.title,
    description: item.description,
    context: item.context,
    details: item.details,
    status: (item.taskStatus || "Todo") as TaskStatus,
    priority: (item.priority || "Medium") as TaskPriority,
    dueAt: item.dueAt,
    createdAt: item.createdAt,
    updatedAt: item.updatedAt,
  };
}

function taskBucket(status?: TaskStatus): ItemBucket | undefined {
  switch (status) {
    case "Done":
      return "Completed";
    case "Canceled":
      return "Dropped";
    case "Todo":
    case "InProgress":
    case undefined:
      return undefined;
    default:
      return undefined;
  }
}

function taskSortBy(sortBy?: TaskSortBy): ItemSortBy | undefined {
  switch (sortBy) {
    case "DueAt":
      return "DueAt";
    default:
      return undefined;
  }
}

export async function listTasksByProject(
  projectId: string,
  q?: ListTasksQuery,
): Promise<Task[]> {
  return listTasks({ ...q, projectId });
}

export async function listTasks(q?: ListTasksQuery): Promise<Task[]> {
  const items = await listItems({
    kind: "Task",
    bucket: taskBucket(q?.status),
    projectId: q?.projectId,
    taskStatus: q?.status === "Todo" || q?.status === "InProgress" ? q.status : undefined,
    search: q?.search,
    sortBy: taskSortBy(q?.sortBy),
    sortDir: q?.sortDir,
    limit: q?.limit,
    offset: q?.offset,
  });
  return items.map(fromItem);
}

export async function getTask(id: string): Promise<Task> {
  return fromItem(await getItem(id));
}

export async function createTask(input: CreateTaskInput): Promise<Task> {
  return fromItem(
    await createItem({
      kind: "Task",
      bucket: taskBucket(input.status) ?? "NextAction",
      projectId: input.projectId,
      title: input.name,
      description: input.description,
      context: input.context,
      details: input.details,
      taskStatus: input.status ?? "Todo",
      priority: input.priority,
      dueAt: input.dueAt,
    }),
  );
}

export async function updateTask(
  id: string,
  input: UpdateTaskInput,
): Promise<Task> {
  const current = await getItem(id);
  return fromItem(
    await updateItem(id, {
      kind: current.kind,
      bucket: current.bucket,
      projectId: input.projectId,
      title: input.name,
      description: input.description,
      context: input.context,
      details: input.details,
      taskStatus: current.taskStatus,
      priority: input.priority,
      waitingFor: current.waitingFor,
      expectedAt: current.expectedAt,
      dueAt: input.dueAt,
    }),
  );
}

export async function setTaskStatus(
  id: string,
  status: TaskStatus,
): Promise<Task> {
  const current = await getItem(id);
  return fromItem(
    await updateItem(id, {
      kind: current.kind,
      bucket: taskBucket(status) ?? "NextAction",
      projectId: current.projectId,
      title: current.title,
      description: current.description,
      context: current.context,
      details: current.details,
      taskStatus: status,
      priority: current.priority,
      waitingFor: current.waitingFor,
      expectedAt: current.expectedAt,
      dueAt: current.dueAt,
    }),
  );
}

export async function deleteTask(id: string): Promise<void> {
  await deleteItem(id);
}
