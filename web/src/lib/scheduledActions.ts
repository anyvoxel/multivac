import { createAction, deleteAction, listActions, updateAction } from "./actions";
import type { Action } from "./actions";

export type ScheduledAction = {
  id: string;
  title: string;
  description: string;
  projectId?: string;
  context: string[];
  labels: { name: string }[];
  startAt: string;
  endAt: string;
  createdAt: string;
  updatedAt: string;
};

export type ListScheduledActionsQuery = {
  search?: string;
  limit?: number;
  offset?: number;
};

export type CreateScheduledActionInput = {
  title: string;
  description: string;
  projectId?: string;
  context?: string[];
  labels?: { name: string }[];
  startAt: string;
  endAt: string;
};

export type UpdateScheduledActionInput = CreateScheduledActionInput;

function fromAction(action: Action): ScheduledAction | null {
  if (action.kind !== "Scheduled") return null;
  const startAt = action.attributes.scheduled?.startAt;
  const endAt = action.attributes.scheduled?.endAt;
  if (!startAt || !endAt) return null;
  return {
    id: action.id,
    title: action.title,
    description: action.description,
    projectId: action.projectId,
    context: action.context,
    labels: action.labels,
    startAt,
    endAt,
    createdAt: action.createdAt,
    updatedAt: action.updatedAt,
  };
}

export async function listScheduledActions(q?: ListScheduledActionsQuery): Promise<ScheduledAction[]> {
  const actions = await listActions({
    search: q?.search,
    kind: "Scheduled",
    limit: q?.limit,
    offset: q?.offset,
  });

  return actions
    .map(fromAction)
    .filter((action): action is ScheduledAction => action !== null);
}

export async function createScheduledAction(input: CreateScheduledActionInput): Promise<ScheduledAction> {
  const action = await createAction({
    title: input.title,
    description: input.description,
    projectId: input.projectId,
    kind: "Scheduled",
    context: input.context ?? [],
    labels: input.labels ?? [],
    attributes: {
      scheduled: {
        startAt: input.startAt,
        endAt: input.endAt,
      },
    },
  });

  const out = fromAction(action);
  if (!out) {
    throw new Error("创建日程失败");
  }
  return out;
}

export async function updateScheduledAction(id: string, input: UpdateScheduledActionInput): Promise<ScheduledAction> {
  const action = await updateAction(id, {
    title: input.title,
    description: input.description,
    projectId: input.projectId,
    kind: "Scheduled",
    context: input.context ?? [],
    labels: input.labels ?? [],
    attributes: {
      scheduled: {
        startAt: input.startAt,
        endAt: input.endAt,
      },
    },
  });

  const out = fromAction(action);
  if (!out) {
    throw new Error("更新日程失败");
  }
  return out;
}

export async function deleteScheduledAction(id: string): Promise<void> {
  await deleteAction(id);
}
