import {
  createAction,
  deleteAction,
  getAction,
  listActions,
  updateAction,
} from "./actions";
import type { Action, ActionLabel } from "./actions";
import type { SortDir } from "./items";
import type { Label } from "./projects";

export type TaskStatus = "Pending" | "Active" | "Completed";
export type TaskPriority = "Low" | "Medium" | "High" | "P0";
export type TaskSortBy = "DueAt";
export type { SortDir };

export type Task = {
  id: string;
  projectId?: string;
  name: string;
  description: string;
  labels: Label[];
  contexts: string[];
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
  labels: Label[];
  contexts: string[];
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
  labels: Label[];
  contexts: string[];
  details: string;
  priority: TaskPriority;
  dueAt?: string;
  status?: TaskStatus;
};

export type ListTasksQuery = {
  projectId?: string;
  status?: TaskStatus;
  search?: string;
  contextIds?: string[];
  tags?: string[];
  sortBy?: TaskSortBy;
  sortDir?: SortDir;
  limit?: number;
  offset?: number;
};

function parseLabelName(name: string): Label {
  const raw = name.trim();
  if (!raw) return { value: "", kind: "Tag", filterable: false };
  if (raw.startsWith("@")) {
    return {
      value: raw.slice(1).trim().toLowerCase(),
      kind: "Context",
      filterable: true,
    };
  }
  if (raw.startsWith("#")) {
    return {
      value: raw.slice(1).trim().toLowerCase(),
      kind: "Tag",
      filterable: true,
    };
  }
  return {
    value: raw.toLowerCase(),
    kind: "Tag",
    filterable: false,
  };
}

function toActionLabel(label: Label): ActionLabel {
  const prefix = label.filterable ? (label.kind === "Context" ? "@" : "#") : "";
  return { name: `${prefix}${label.value}` };
}

function normalizeTaskDueAt(value?: string): string | undefined {
  if (!value) return undefined;
  const trimmed = value.trim();
  return trimmed || undefined;
}

function fromAction(action: Action): Task {
  const labels = action.labels
    .map((label) => parseLabelName(label.name))
    .filter((label) => label.value !== "");
  return {
    id: action.id,
    projectId: action.projectId,
    name: action.title,
    description: action.description,
    labels,
    contexts: action.context,
    details: "",
    status: action.attributes.task?.status ?? "Pending",
    priority: "Medium",
    dueAt: action.attributes.task?.expectedAt,
    createdAt: action.createdAt,
    updatedAt: action.updatedAt,
  };
}

function expectedAtTimestamp(value?: string): number | null {
  if (!value) return null;
  const timestamp = Date.parse(value);
  return Number.isNaN(timestamp) ? null : timestamp;
}

function sortByDueAt(list: Task[], dir: SortDir): Task[] {
  return list
    .map((item, index) => ({ item, index, timestamp: expectedAtTimestamp(item.dueAt) }))
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

function matchesTags(task: Task, tags?: string[]): boolean {
  if (!tags || tags.length === 0) return true;
  const wanted = new Set(tags.map((tag) => tag.trim().toLowerCase()).filter(Boolean));
  if (wanted.size === 0) return true;
  return task.labels.some((label) => label.kind === "Tag" && wanted.has(label.value));
}

function matchesStatus(task: Task, status?: TaskStatus): boolean {
  if (!status) return true;
  return task.status === status;
}

export async function listTasksByProject(
  projectId: string,
  q?: ListTasksQuery,
): Promise<Task[]> {
  return listTasks({ ...q, projectId });
}

export async function listTasks(q?: ListTasksQuery): Promise<Task[]> {
  const needLocalProcess =
    q?.sortBy === "DueAt" ||
    (q?.tags?.length ?? 0) > 0 ||
    !!q?.status;

  const actions = await listActions({
    search: q?.search,
    kind: "Task",
    projectId: q?.projectId,
    contextIds: q?.contextIds,
    sortDir: q?.sortDir,
    limit: needLocalProcess ? undefined : q?.limit,
    offset: needLocalProcess ? undefined : q?.offset,
  });

  let list = actions.filter((action) => action.kind === "Task").map(fromAction);
  list = list.filter((task) => matchesStatus(task, q?.status));
  list = list.filter((task) => matchesTags(task, q?.tags));

  if (q?.sortBy === "DueAt") {
    list = sortByDueAt(list, q?.sortDir ?? "Asc");
  }

  if (!needLocalProcess) {
    return list;
  }

  const offset = q?.offset ?? 0;
  if (offset > 0) {
    list = list.slice(offset);
  }
  if (q?.limit !== undefined) {
    list = list.slice(0, q.limit);
  }
  return list;
}

export async function getTask(id: string): Promise<Task> {
  return fromAction(await getAction(id));
}

export async function createTask(input: CreateTaskInput): Promise<Task> {
  return fromAction(
    await createAction({
      title: input.name,
      description: input.description,
      projectId: input.projectId,
      kind: "Task",
      context: input.contexts.filter((c) => c.trim()),
      labels: input.labels.map(toActionLabel),
      attributes: {
        task: {
          expectedAt: normalizeTaskDueAt(input.dueAt),
          status: input.status ?? "Pending",
        },
      },
    }),
  );
}

export async function updateTask(
  id: string,
  input: UpdateTaskInput,
): Promise<Task> {
  return fromAction(
    await updateAction(id, {
      title: input.name,
      description: input.description,
      projectId: input.projectId,
      kind: "Task",
      context: input.contexts.filter((c) => c.trim()),
      labels: input.labels.map(toActionLabel),
      attributes: {
        task: {
          expectedAt: normalizeTaskDueAt(input.dueAt),
          status: input.status ?? "Pending",
        },
      },
    }),
  );
}

export async function setTaskStatus(
  id: string,
  _status: TaskStatus,
): Promise<Task> {
  return getTask(id);
}

export async function deleteTask(id: string): Promise<void> {
  await deleteAction(id);
}
