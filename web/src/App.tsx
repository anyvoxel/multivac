import type { PointerEvent as ReactPointerEvent, ReactNode } from "react";
import { useEffect, useMemo, useState } from "react";

import {
  createProject,
  deleteProject,
  getProject,
  listProjects,
  setProjectStatus,
  updateProject,
} from "./lib/projects";
import type { Project, ProjectStatus } from "./lib/projects";
import {
  createTask,
  deleteTask,
  getTask,
  listTasks,
  setTaskStatus,
  updateTask,
} from "./lib/tasks";
import type { SortDir, Task, TaskPriority, TaskStatus } from "./lib/tasks";
import {
  createInbox,
  deleteInbox,
  getInbox,
  listInboxes,
  updateInbox,
} from "./lib/inboxes";
import type { Inbox } from "./lib/inboxes";
import {
  createSomeday,
  deleteSomeday,
  getSomeday,
  listSomedays,
  updateSomeday,
} from "./lib/somedays";
import type { Someday } from "./lib/somedays";
import {
  createWaitingList,
  deleteWaitingList,
  getWaitingList,
  listWaitingLists,
  updateWaitingList,
} from "./lib/waitingLists";
import type { WaitingList } from "./lib/waitingLists";
import {
  formatDate,
  formatDateTime,
  fromDateTimeLocalValue,
  fromDateValue,
  toDateTimeLocalValue,
  toDateValue,
} from "./lib/time";

type Route = "projects" | "tasks" | "inboxes" | "somedays" | "waitingLists";

const PROJECT_STATUSES: ProjectStatus[] = [
  "Draft",
  "Active",
  "Completed",
  "Archived",
];
const TASK_STATUSES: TaskStatus[] = ["Todo", "InProgress", "Done", "Canceled"];
const TASK_PRIORITIES: TaskPriority[] = ["P0", "High", "Medium", "Low"];
const PAGE_SIZE_OPTIONS = [10, 20, 50];
const LAST_TASK_PROJECT_STORAGE_KEY = "multivac:last-task-project-id";
const EMPTY_TASK_PROJECT_SENTINEL = "__NONE__";

function classNames(...parts: Array<string | false | null | undefined>) {
  return parts.filter(Boolean).join(" ");
}

function projectStatusLabel(s: ProjectStatus): string {
  switch (s) {
    case "Draft":
      return "草稿";
    case "Active":
      return "进行中";
    case "Completed":
      return "已完成";
    case "Archived":
      return "归档";
    default:
      return s;
  }
}

function taskStatusLabel(s: TaskStatus): string {
  switch (s) {
    case "Todo":
      return "待办";
    case "InProgress":
      return "进行中";
    case "Done":
      return "已完成";
    case "Canceled":
      return "已取消";
    default:
      return s;
  }
}

function taskPriorityLabel(p: TaskPriority): string {
  switch (p) {
    case "P0":
      return "P0";
    case "High":
      return "高";
    case "Medium":
      return "中";
    case "Low":
      return "低";
    default:
      return p;
  }
}

function pageNumber(offset: number, pageSize: number): number {
  return Math.floor(offset / pageSize) + 1;
}

function IconLogo() {
  return (
    <div className="flex h-8 w-8 items-center justify-center rounded-md bg-[#4F46E5] text-white">
      <span className="text-sm font-bold">M</span>
    </div>
  );
}

function IconProject() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M4 6.5C4 5.12 5.12 4 6.5 4h11C18.88 4 20 5.12 20 6.5v11c0 1.38-1.12 2.5-2.5 2.5h-11C5.12 20 4 18.88 4 17.5v-11Z"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path
        d="M7 8h10M7 12h7"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  );
}

function IconTask() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M7 7h10M7 12h10M7 17h10"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <path
        d="M4.5 7.5l.8.8 1.6-1.9M4.5 12.5l.8.8 1.6-1.9M4.5 17.5l.8.8 1.6-1.9"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function IconInbox() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M4 7.5C4 6.12 5.12 5 6.5 5h11C18.88 5 20 6.12 20 7.5v9c0 1.38-1.12 2.5-2.5 2.5h-11C5.12 19 4 17.88 4 16.5v-9Z"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path
        d="M4 14h4l1.5 2h5L16 14h4"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function IconWaitingList() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M7 4.75h10A2.25 2.25 0 0 1 19.25 7v10A2.25 2.25 0 0 1 17 19.25H7A2.25 2.25 0 0 1 4.75 17V7A2.25 2.25 0 0 1 7 4.75Z"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path
        d="M8 8.5h8M8 12h8M8 15.5h5"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <path d="M15.5 4.75v3.5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

function IconSearch() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M11 19a8 8 0 1 1 0-16 8 8 0 0 1 0 16Z"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path
        d="M21 21l-4.3-4.3"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
    </svg>
  );
}

function IconRefresh() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      aria-hidden="true"
    >
      <path
        d="M20 12a8 8 0 0 1-14.93 4M4 12a8 8 0 0 1 14.93-4"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <path
        d="M20 7v5h-5M4 17v-5h5"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function IconClose() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path d="M6 6l12 12M18 6L6 18" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

function IconClear() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path
        d="M8 8l8 8M16 8l-8 8"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
      />
      <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  );
}

function IconClock() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path
        d="M12 21a9 9 0 1 0-9-9 9 9 0 0 0 9 9Z"
        stroke="currentColor"
        strokeWidth="1.5"
      />
      <path
        d="M12 7v6l4 2"
        stroke="currentColor"
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

function IconSort(props: { dir: SortDir | null }) {
  const activeUp = props.dir === "Asc";
  const activeDown = props.dir === "Desc";
  return (
    <svg width="12" height="12" viewBox="0 0 12 12" fill="none" aria-hidden="true">
      <path
        d="M4 5 6 3l2 2"
        stroke={activeUp ? "currentColor" : "#9CA3AF"}
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M4 7 6 9l2-2"
        stroke={activeDown ? "currentColor" : "#9CA3AF"}
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export default function App() {
  const [route, setRoute] = useState<Route>("projects");

  const [search, setSearch] = useState<string>("");

  // Projects list page state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [projects, setProjects] = useState<Project[]>([]);
  const [allProjects, setAllProjects] = useState<Project[]>([]);
  const [projectStatusFilter, setProjectStatusFilter] = useState<
    ProjectStatus | ""
  >("");
  const [projectOffset, setProjectOffset] = useState(0);
  const [projectPageSize, setProjectPageSize] = useState(PAGE_SIZE_OPTIONS[0]);
  const [projectHasNext, setProjectHasNext] = useState(false);

  // Inboxes list page state
  const [inboxLoading, setInboxLoading] = useState(false);
  const [inboxError, setInboxError] = useState<string>("");
  const [inboxes, setInboxes] = useState<Inbox[]>([]);
  const [inboxOffset, setInboxOffset] = useState(0);
  const [inboxPageSize, setInboxPageSize] = useState(PAGE_SIZE_OPTIONS[0]);
  const [inboxHasNext, setInboxHasNext] = useState(false);

  // Somedays list page state
  const [somedayLoading, setSomedayLoading] = useState(false);
  const [somedayError, setSomedayError] = useState<string>("");
  const [somedays, setSomedays] = useState<Someday[]>([]);
  const [somedayOffset, setSomedayOffset] = useState(0);
  const [somedayPageSize, setSomedayPageSize] = useState(PAGE_SIZE_OPTIONS[0]);
  const [somedayHasNext, setSomedayHasNext] = useState(false);

  // Waiting lists page state
  const [waitingListLoading, setWaitingListLoading] = useState(false);
  const [waitingListError, setWaitingListError] = useState<string>("");
  const [waitingLists, setWaitingLists] = useState<WaitingList[]>([]);
  const [waitingListOffset, setWaitingListOffset] = useState(0);
  const [waitingListPageSize, setWaitingListPageSize] = useState(PAGE_SIZE_OPTIONS[0]);
  const [waitingListHasNext, setWaitingListHasNext] = useState(false);

  // Drawer state
  const [drawer, setDrawer] = useState<
    | { type: "none" }
    | { type: "project"; mode: "create" }
    | { type: "project"; mode: "edit"; id: string }
    | { type: "task"; mode: "create" }
    | { type: "task"; mode: "edit"; id: string }
    | { type: "inbox"; mode: "create" }
    | { type: "inbox"; mode: "edit"; id: string }
    | { type: "someday"; mode: "create" }
    | { type: "someday"; mode: "edit"; id: string }
    | { type: "waitingList"; mode: "create" }
    | { type: "waitingList"; mode: "edit"; id: string }
  >({ type: "none" });
  const [drawerLoading, setDrawerLoading] = useState(false);
  const [drawerSaving, setDrawerSaving] = useState(false);
  const [drawerProject, setDrawerProject] = useState<Project | null>(null);
  const [drawerTask, setDrawerTask] = useState<Task | null>(null);
  const [drawerInbox, setDrawerInbox] = useState<Inbox | null>(null);
  const [drawerSomeday, setDrawerSomeday] = useState<Someday | null>(null);
  const [drawerWaitingList, setDrawerWaitingList] = useState<WaitingList | null>(null);

  // Tasks page filters
  const [taskProjectId, setTaskProjectId] = useState<string>("");
  const [taskStatusFilter, setTaskStatusFilter] = useState<TaskStatus | "">("");
  const [taskListVersion, setTaskListVersion] = useState(0);
  const [highlightTaskId, setHighlightTaskId] = useState<string | null>(null);
  const [lastTaskProjectId, setLastTaskProjectId] = useState<string | null>(() => {
    if (typeof window === "undefined") return null;
    const stored = window.localStorage.getItem(LAST_TASK_PROJECT_STORAGE_KEY);
    if (stored === null) return null;
    if (stored === EMPTY_TASK_PROJECT_SENTINEL) return "";
    return stored;
  });

  const normalizedSearch = search.trim();

  function closeDrawer() {
    setDrawer({ type: "none" });
    setDrawerProject(null);
    setDrawerTask(null);
    setDrawerInbox(null);
    setDrawerSomeday(null);
    setDrawerWaitingList(null);
  }

  useEffect(() => {
    if (!highlightTaskId) return;
    const timer = window.setTimeout(() => setHighlightTaskId(null), 2500);
    return () => window.clearTimeout(timer);
  }, [highlightTaskId]);

  useEffect(() => {
    if (typeof window === "undefined") return;
    if (lastTaskProjectId === null) {
      window.localStorage.removeItem(LAST_TASK_PROJECT_STORAGE_KEY);
      return;
    }
    window.localStorage.setItem(
      LAST_TASK_PROJECT_STORAGE_KEY,
      lastTaskProjectId === "" ? EMPTY_TASK_PROJECT_SENTINEL : lastTaskProjectId,
    );
  }, [lastTaskProjectId]);

  async function refreshProjects() {
    setLoading(true);
    setError("");
    try {
      const list = await listProjects({
        status: projectStatusFilter || undefined,
        search: normalizedSearch || undefined,
        limit: projectPageSize + 1,
        offset: projectOffset,
      });
      setProjects(list.slice(0, projectPageSize));
      setProjectHasNext(list.length > projectPageSize);
      const all = await listProjects();
      setAllProjects(all);
      if (taskProjectId) {
        const still = all.find((p) => p.id === taskProjectId);
        if (!still) setTaskProjectId("");
      }
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  async function refreshInboxes() {
    setInboxLoading(true);
    setInboxError("");
    try {
      const list = await listInboxes({
        search: normalizedSearch || undefined,
        limit: inboxPageSize + 1,
        offset: inboxOffset,
      });
      setInboxes(list.slice(0, inboxPageSize));
      setInboxHasNext(list.length > inboxPageSize);
    } catch (e) {
      setInboxError(String(e));
    } finally {
      setInboxLoading(false);
    }
  }

  async function refreshSomedays() {
    setSomedayLoading(true);
    setSomedayError("");
    try {
      const list = await listSomedays({
        search: normalizedSearch || undefined,
        limit: somedayPageSize + 1,
        offset: somedayOffset,
      });
      setSomedays(list.slice(0, somedayPageSize));
      setSomedayHasNext(list.length > somedayPageSize);
    } catch (e) {
      setSomedayError(String(e));
    } finally {
      setSomedayLoading(false);
    }
  }

  async function refreshWaitingLists() {
    setWaitingListLoading(true);
    setWaitingListError("");
    try {
      const list = await listWaitingLists({
        search: normalizedSearch || undefined,
        limit: waitingListPageSize + 1,
        offset: waitingListOffset,
      });
      setWaitingLists(list.slice(0, waitingListPageSize));
      setWaitingListHasNext(list.length > waitingListPageSize);
    } catch (e) {
      setWaitingListError(String(e));
    } finally {
      setWaitingListLoading(false);
    }
  }

  useEffect(() => {
    setProjectOffset(0);
  }, [projectStatusFilter, search, projectPageSize]);

  useEffect(() => {
    setInboxOffset(0);
  }, [search, inboxPageSize]);

  useEffect(() => {
    setSomedayOffset(0);
  }, [search, somedayPageSize]);

  useEffect(() => {
    setWaitingListOffset(0);
  }, [search, waitingListPageSize]);

  useEffect(() => {
    void refreshProjects();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectStatusFilter, projectOffset, projectPageSize, normalizedSearch]);

  useEffect(() => {
    if (route !== "inboxes") return;
    void refreshInboxes();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [route, inboxOffset, inboxPageSize, normalizedSearch]);

  useEffect(() => {
    if (route !== "somedays") return;
    void refreshSomedays();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [route, somedayOffset, somedayPageSize, normalizedSearch]);

  useEffect(() => {
    if (route !== "waitingLists") return;
    void refreshWaitingLists();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [route, waitingListOffset, waitingListPageSize, normalizedSearch]);

  const pageLabelMap: Record<Route, string> = {
    projects: "项目管理",
    tasks: "任务管理",
    inboxes: "收集箱管理",
    somedays: "将来/也许管理",
    waitingLists: "等待列表管理",
  };
  const pageLabel = pageLabelMap[route];

  async function openProjectDrawer(mode: "create" | "edit", id?: string) {
    setDrawerProject(null);
    setDrawerTask(null);
    setDrawerInbox(null);
    setDrawerSomeday(null);
    setDrawerWaitingList(null);
    setDrawerLoading(true);
    setError("");
    try {
      if (mode === "create") {
        setDrawer({ type: "project", mode: "create" });
        setDrawerProject({
          id: "",
          name: "",
          goal: "",
          principles: "",
          visionResult: "",
          description: "",
          status: "Draft",
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        });
        return;
      }
      if (!id) return;
      setDrawer({ type: "project", mode: "edit", id });
      const p = await getProject(id);
      setDrawerProject(p);
    } catch (e) {
      setError(String(e));
    } finally {
      setDrawerLoading(false);
    }
  }

  async function openTaskDrawer(mode: "create" | "edit", id?: string, t?: Task) {
    setDrawerProject(null);
    setDrawerTask(null);
    setDrawerInbox(null);
    setDrawerSomeday(null);
    setDrawerWaitingList(null);
    setDrawerLoading(true);
    setError("");
    try {
      if (mode === "create") {
        setDrawer({ type: "task", mode: "create" });
        const rememberedProject =
          lastTaskProjectId === ""
            ? ""
            : lastTaskProjectId && allProjects.some((p) => p.id === lastTaskProjectId)
              ? lastTaskProjectId
              : null;
        const defaultProject = rememberedProject ?? (taskProjectId || "");
        setDrawerTask({
          id: "",
          projectId: defaultProject,
          name: "",
          description: "",
          context: "默认",
          details: "",
          status: "Todo",
          priority: "Medium",
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        });
        return;
      }
      if (!id) return;
      setDrawer({ type: "task", mode: "edit", id });
      if (t) {
        setDrawerTask(t);
        return;
      }
      const found = await getTask(id);
      setDrawerTask(found);
    } catch (e) {
      setError(String(e));
    } finally {
      setDrawerLoading(false);
    }
  }

  async function openInboxDrawer(mode: "create" | "edit", id?: string, inbox?: Inbox) {
    setDrawerProject(null);
    setDrawerTask(null);
    setDrawerInbox(null);
    setDrawerSomeday(null);
    setDrawerWaitingList(null);
    setDrawerLoading(true);
    setInboxError("");
    try {
      if (mode === "create") {
        setDrawer({ type: "inbox", mode: "create" });
        setDrawerInbox({
          id: "",
          name: "",
          description: "",
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        });
        return;
      }
      if (!id) return;
      setDrawer({ type: "inbox", mode: "edit", id });
      if (inbox) {
        setDrawerInbox(inbox);
        return;
      }
      const found = await getInbox(id);
      setDrawerInbox(found);
    } catch (e) {
      setInboxError(String(e));
    } finally {
      setDrawerLoading(false);
    }
  }

  async function openSomedayDrawer(mode: "create" | "edit", id?: string, someday?: Someday) {
    setDrawerProject(null);
    setDrawerTask(null);
    setDrawerInbox(null);
    setDrawerSomeday(null);
    setDrawerWaitingList(null);
    setDrawerLoading(true);
    setSomedayError("");
    try {
      if (mode === "create") {
        setDrawer({ type: "someday", mode: "create" });
        setDrawerSomeday({
          id: "",
          name: "",
          description: "",
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        });
        return;
      }
      if (!id) return;
      setDrawer({ type: "someday", mode: "edit", id });
      if (someday) {
        setDrawerSomeday(someday);
        return;
      }
      const found = await getSomeday(id);
      setDrawerSomeday(found);
    } catch (e) {
      setSomedayError(String(e));
    } finally {
      setDrawerLoading(false);
    }
  }

  async function openWaitingListDrawer(mode: "create" | "edit", id?: string, waitingList?: WaitingList) {
    setDrawerProject(null);
    setDrawerTask(null);
    setDrawerInbox(null);
    setDrawerSomeday(null);
    setDrawerWaitingList(null);
    setDrawerLoading(true);
    setWaitingListError("");
    try {
      if (mode === "create") {
        setDrawer({ type: "waitingList", mode: "create" });
        setDrawerWaitingList({
          id: "",
          name: "",
          details: "",
          owner: "",
          expectedAt: undefined,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        });
        return;
      }
      if (!id) return;
      setDrawer({ type: "waitingList", mode: "edit", id });
      if (waitingList) {
        setDrawerWaitingList(waitingList);
        return;
      }
      const found = await getWaitingList(id);
      setDrawerWaitingList(found);
    } catch (e) {
      setWaitingListError(String(e));
    } finally {
      setDrawerLoading(false);
    }
  }

  async function onDeleteProject(p: Project) {
    if (!confirm(`确定删除 Project: ${p.name} ?`)) return;
    setLoading(true);
    setError("");
    try {
      await deleteProject(p.id);
      if (drawer.type === "project" && drawer.mode === "edit" && drawer.id === p.id) {
        setDrawer({ type: "none" });
        setDrawerProject(null);
      }
      await refreshProjects();
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  async function onRefreshAll() {
    switch (route) {
      case "projects":
        await refreshProjects();
        return;
      case "inboxes":
        await refreshInboxes();
        return;
      case "somedays":
        await refreshSomedays();
        return;
      case "waitingLists":
        await refreshWaitingLists();
        return;
      case "tasks":
      default:
        await refreshProjects();
        return;
    }
  }

  return (
    <div className="min-h-full bg-[#F5F6FA]">
      <div className="flex min-h-screen">
        <aside className="w-72 border-r border-[#E6E8F0] bg-[#F7F7FA] px-5 py-3">
          <div className="flex items-center gap-2 px-2 py-3">
            <IconLogo />
            <div className="text-sm font-semibold tracking-[0.08em] text-[#111827]">MULTIVAC</div>
          </div>

          <div className="px-2 pb-3 pt-4 text-sm font-semibold text-[#6B7280]">专注</div>
          <nav className="grid gap-1 px-2">
            <NavItemLight
              active={route === "inboxes"}
              icon={<IconInbox />}
              label="收集箱"
              onClick={() => {
                setRoute("inboxes");
                setSearch("");
              }}
            />
          </nav>

          <div className="px-2 pb-3 pt-8 text-sm font-semibold text-[#6B7280]">列表</div>
          <nav className="grid gap-1 px-2">
            <NavItemLight
              active={route === "projects"}
              icon={<IconProject />}
              label="项目"
              onClick={() => {
                setRoute("projects");
                setSearch("");
              }}
            />
            <NavItemLight
              active={route === "tasks"}
              icon={<IconTask />}
              label="任务"
              onClick={() => {
                setRoute("tasks");
                setSearch("");
              }}
            />
            <NavItemLight
              active={route === "somedays"}
              icon={<IconClock />}
              label="将来/也许"
              onClick={() => {
                setRoute("somedays");
                setSearch("");
              }}
            />
            <NavItemLight
              active={route === "waitingLists"}
              icon={<IconWaitingList />}
              label="等待中"
              onClick={() => {
                setRoute("waitingLists");
                setSearch("");
              }}
            />
          </nav>
        </aside>

        <div className="flex min-w-0 flex-1 flex-col">
          <header className="border-b border-[#E6E8F0] bg-white">
            <div className="flex items-center justify-between gap-3 px-6 py-3">
              <div className="flex items-center gap-2 text-sm text-[#6B7280]">
                <span className="font-semibold text-[#111827]">MULTIVAC</span>
                <span>/</span>
                <span>{pageLabel}</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="relative">
                  <div className="pointer-events-none absolute left-2 top-1/2 -translate-y-1/2 text-[#9CA3AF]">
                    <IconSearch />
                  </div>
                  <input
                    className="w-64 rounded-md border border-[#E6E8F0] bg-white py-1.5 pl-8 pr-9 text-sm outline-none focus:ring-2 focus:ring-[#4F46E5]"
                    placeholder="搜索..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                  />
                  {search ? (
                    <button
                      type="button"
                      className="absolute right-2 top-1/2 -translate-y-1/2 text-[#9CA3AF] hover:text-[#6B7280]"
                      onClick={() => setSearch("")}
                      aria-label="清空搜索"
                    >
                      <IconClear />
                    </button>
                  ) : null}
                </div>
                <button
                  className="flex items-center justify-center rounded-md border border-[#E6E8F0] bg-white p-2 text-[#6B7280] hover:bg-[#F5F6FA]"
                  type="button"
                  onClick={() => void onRefreshAll()}
                  aria-label="refresh"
                >
                  <IconRefresh />
                </button>
                <button
                  className="rounded-md bg-[#4F46E5] px-3 py-1.5 text-sm font-medium text-white hover:opacity-90"
                  type="button"
                  onClick={() => {
                    if (route === "projects") void openProjectDrawer("create");
                    else if (route === "tasks") void openTaskDrawer("create");
                    else if (route === "inboxes") void openInboxDrawer("create");
                    else if (route === "somedays") void openSomedayDrawer("create");
                    else void openWaitingListDrawer("create");
                  }}
                >
                  + 新建
                </button>
              </div>
            </div>
          </header>

          <main className="min-w-0 flex-1 px-6 py-6">
            {route === "projects" && error ? (
              <div className="mb-4 rounded-md border border-[#FCA5A5] bg-[#FEF2F2] px-3 py-2 text-sm text-[#B91C1C]">
                {error}
              </div>
            ) : null}
            {route === "inboxes" && inboxError ? (
              <div className="mb-4 rounded-md border border-[#FCA5A5] bg-[#FEF2F2] px-3 py-2 text-sm text-[#B91C1C]">
                {inboxError}
              </div>
            ) : null}
            {route === "somedays" && somedayError ? (
              <div className="mb-4 rounded-md border border-[#FCA5A5] bg-[#FEF2F2] px-3 py-2 text-sm text-[#B91C1C]">
                {somedayError}
              </div>
            ) : null}
            {route === "waitingLists" && waitingListError ? (
              <div className="mb-4 rounded-md border border-[#FCA5A5] bg-[#FEF2F2] px-3 py-2 text-sm text-[#B91C1C]">
                {waitingListError}
              </div>
            ) : null}

            {route === "projects" ? (
              <ProjectsPage
                items={projects}
                loading={loading}
                statusFilter={projectStatusFilter}
                pageSize={projectPageSize}
                offset={projectOffset}
                hasNext={projectHasNext}
                paginationEnabled
                onStatusFilter={setProjectStatusFilter}
                onPageSizeChange={setProjectPageSize}
                onPrevPage={() => setProjectOffset((v) => Math.max(0, v - projectPageSize))}
                onNextPage={() => setProjectOffset((v) => v + projectPageSize)}
                onRefresh={() => void refreshProjects()}
                onOpen={(id) => void openProjectDrawer("edit", id)}
              />
            ) : route === "tasks" ? (
              <TasksPageNew
                projects={allProjects}
                projectId={taskProjectId}
                status={taskStatusFilter}
                search={normalizedSearch}
                version={taskListVersion}
                highlightTaskId={highlightTaskId}
                onProjectId={setTaskProjectId}
                onStatus={setTaskStatusFilter}
                onOpen={(id, t) => void openTaskDrawer("edit", id, t)}
              />
            ) : route === "inboxes" ? (
              <InboxesPage
                items={inboxes}
                loading={inboxLoading}
                pageSize={inboxPageSize}
                offset={inboxOffset}
                hasNext={inboxHasNext}
                onPageSizeChange={setInboxPageSize}
                onPrevPage={() => setInboxOffset((v) => Math.max(0, v - inboxPageSize))}
                onNextPage={() => setInboxOffset((v) => v + inboxPageSize)}
                onRefresh={() => void refreshInboxes()}
                onOpen={(id, inbox) => void openInboxDrawer("edit", id, inbox)}
              />
            ) : route === "somedays" ? (
              <SomedaysPage
                items={somedays}
                loading={somedayLoading}
                pageSize={somedayPageSize}
                offset={somedayOffset}
                hasNext={somedayHasNext}
                onPageSizeChange={setSomedayPageSize}
                onPrevPage={() => setSomedayOffset((v) => Math.max(0, v - somedayPageSize))}
                onNextPage={() => setSomedayOffset((v) => v + somedayPageSize)}
                onRefresh={() => void refreshSomedays()}
                onOpen={(id, someday) => void openSomedayDrawer("edit", id, someday)}
              />
            ) : (
              <WaitingListsPage
                items={waitingLists}
                loading={waitingListLoading}
                pageSize={waitingListPageSize}
                offset={waitingListOffset}
                hasNext={waitingListHasNext}
                onPageSizeChange={setWaitingListPageSize}
                onPrevPage={() => setWaitingListOffset((v) => Math.max(0, v - waitingListPageSize))}
                onNextPage={() => setWaitingListOffset((v) => v + waitingListPageSize)}
                onRefresh={() => void refreshWaitingLists()}
                onOpen={(id, waitingList) => void openWaitingListDrawer("edit", id, waitingList)}
              />
            )}
          </main>
        </div>

        {drawer.type !== "none" ? (
          <Drawer
            title={
              drawer.type === "project"
                ? "项目详情"
                : drawer.type === "task"
                  ? "任务详情"
                  : drawer.type === "inbox"
                    ? "收集箱详情"
                    : drawer.type === "someday"
                      ? "将来/也许详情"
                      : "等待列表详情"
            }
            onClose={closeDrawer}
            action={
              <button
                className={classNames(
                  "rounded-md px-3 py-1.5 text-sm font-medium",
                  drawerSaving
                    ? "cursor-not-allowed bg-[#E5E7EB] text-[#9CA3AF]"
                    : "bg-[#4F46E5] text-white hover:opacity-90",
                )}
                type="button"
                disabled={drawerSaving}
                onClick={() => void onSaveDrawer()}
              >
                保存
              </button>
            }
          >
            {drawerLoading ? (
              <div className="px-4 py-6 text-sm text-[#6B7280]">加载中...</div>
            ) : drawer.type === "project" ? (
              <ProjectDrawerForm
                project={drawerProject}
                mode={drawer.mode}
                onChange={setDrawerProject}
                onDelete={async () => {
                  if (!drawerProject) return;
                  await onDeleteProject(drawerProject);
                }}
                onGotoTasks={(pid) => {
                  setTaskProjectId(pid);
                  setRoute("tasks");
                  closeDrawer();
                }}
              />
            ) : drawer.type === "task" ? (
              <TaskDrawerForm
                task={drawerTask}
                mode={drawer.mode}
                projects={allProjects}
                onChange={setDrawerTask}
                onDelete={async () => {
                  if (!drawerTask) return;
                  if (!confirm(`确定删除 Task: ${drawerTask.name} ?`)) return;
                  await deleteTask(drawerTask.id);
                  setTaskListVersion((v) => v + 1);
                  setDrawer({ type: "none" });
                  setDrawerTask(null);
                }}
              />
            ) : drawer.type === "inbox" ? (
              <InboxDrawerForm
                inbox={drawerInbox}
                mode={drawer.mode}
                onChange={setDrawerInbox}
                onDelete={async () => {
                  if (!drawerInbox) return;
                  if (!confirm(`确定删除收集箱: ${drawerInbox.name} ?`)) return;
                  await deleteInbox(drawerInbox.id);
                  await refreshInboxes();
                  setDrawer({ type: "none" });
                  setDrawerInbox(null);
                }}
              />
            ) : drawer.type === "someday" ? (
              <SomedayDrawerForm
                someday={drawerSomeday}
                mode={drawer.mode}
                onChange={setDrawerSomeday}
                onDelete={async () => {
                  if (!drawerSomeday) return;
                  if (!confirm(`确定删除将来/也许: ${drawerSomeday.name} ?`)) return;
                  await deleteSomeday(drawerSomeday.id);
                  await refreshSomedays();
                  setDrawer({ type: "none" });
                  setDrawerSomeday(null);
                }}
              />
            ) : (
              <WaitingListDrawerForm
                waitingList={drawerWaitingList}
                mode={drawer.mode}
                onChange={setDrawerWaitingList}
                onDelete={async () => {
                  if (!drawerWaitingList) return;
                  if (!confirm(`确定删除等待列表: ${drawerWaitingList.name} ?`)) return;
                  await deleteWaitingList(drawerWaitingList.id);
                  await refreshWaitingLists();
                  setDrawer({ type: "none" });
                  setDrawerWaitingList(null);
                }}
              />
            )}
          </Drawer>
        ) : null}
      </div>
    </div>
  );

  async function onSaveDrawer() {
    setDrawerSaving(true);
    setError("");
    try {
      if (drawer.type === "project") {
        if (!drawerProject) return;
        const status = drawerProject.status;
        if (drawer.mode === "create") {
          const p = await createProject({
            name: drawerProject.name,
            goal: drawerProject.goal,
            principles: drawerProject.principles,
            visionResult: drawerProject.visionResult,
            description: drawerProject.description,
          });
          if (status !== "Draft") {
            await setProjectStatus(p.id, status);
          }
          await refreshProjects();
          await openProjectDrawer("edit", p.id);
          return;
        }
        // edit
        const current = await getProject(drawerProject.id);
        await updateProject(drawerProject.id, {
          name: drawerProject.name,
          goal: drawerProject.goal,
          principles: drawerProject.principles,
          visionResult: drawerProject.visionResult,
          description: drawerProject.description,
        });
        if (current.status !== status) {
          await setProjectStatus(drawerProject.id, status);
        }
        await refreshProjects();
        setDrawer({ type: "none" });
        setDrawerProject(null);
        return;
      }

      if (drawer.type === "task") {
        if (!drawerTask) return;
        const dueISO = fromDateValue(drawerTask.dueAt ? toDateValue(drawerTask.dueAt) : "");
        if (drawer.mode === "create") {
          const created = await createTask({
            projectId: drawerTask.projectId || undefined,
            name: drawerTask.name,
            description: drawerTask.description,
            context: drawerTask.context,
            details: drawerTask.details,
            priority: drawerTask.priority,
            status: drawerTask.status,
            dueAt: dueISO,
          });
          setLastTaskProjectId(drawerTask.projectId ?? "");
          setHighlightTaskId(created.id);
          setTaskListVersion((v) => v + 1);
          closeDrawer();
          return;
        }
        const t = await updateTask(drawerTask.id, {
          projectId: drawerTask.projectId || undefined,
          name: drawerTask.name,
          description: drawerTask.description,
          context: drawerTask.context,
          details: drawerTask.details,
          priority: drawerTask.priority,
          dueAt: dueISO ?? "",
        });
        setLastTaskProjectId(drawerTask.projectId ?? "");
        if (t.status !== drawerTask.status) {
          await setTaskStatus(drawerTask.id, drawerTask.status);
        }
        setTaskListVersion((v) => v + 1);
        setDrawer({ type: "none" });
        setDrawerTask(null);
        return;
      }

      if (drawer.type === "inbox") {
        if (!drawerInbox) return;
        if (drawer.mode === "create") {
          await createInbox({
            name: drawerInbox.name,
            description: drawerInbox.description,
          });
        } else {
          await updateInbox(drawerInbox.id, {
            name: drawerInbox.name,
            description: drawerInbox.description,
          });
        }
        await refreshInboxes();
        closeDrawer();
        return;
      }

      if (drawer.type === "someday") {
        if (!drawerSomeday) return;
        if (drawer.mode === "create") {
          await createSomeday({
            name: drawerSomeday.name,
            description: drawerSomeday.description,
          });
        } else {
          await updateSomeday(drawerSomeday.id, {
            name: drawerSomeday.name,
            description: drawerSomeday.description,
          });
        }
        await refreshSomedays();
        closeDrawer();
        return;
      }

      if (drawer.type === "waitingList") {
        if (!drawerWaitingList) return;
        const expectedAt = fromDateTimeLocalValue(
          drawerWaitingList.expectedAt ? toDateTimeLocalValue(drawerWaitingList.expectedAt) : "",
        );
        if (drawer.mode === "create") {
          await createWaitingList({
            name: drawerWaitingList.name,
            details: drawerWaitingList.details,
            owner: drawerWaitingList.owner,
            expectedAt,
          });
        } else {
          await updateWaitingList(drawerWaitingList.id, {
            name: drawerWaitingList.name,
            details: drawerWaitingList.details,
            owner: drawerWaitingList.owner,
            expectedAt,
          });
        }
        await refreshWaitingLists();
        closeDrawer();
      }
    } catch (e) {
      setError(String(e));
    } finally {
      setDrawerSaving(false);
    }
  }
}

function NavItemLight(props: {
  active: boolean;
  icon: ReactNode;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      className={classNames(
        "flex w-full items-center gap-3 rounded-2xl px-4 py-4 text-left text-[15px] transition-colors",
        props.active
          ? "bg-[#ECECEF] text-[#111827]"
          : "text-[#6B7280] hover:bg-[#F0F1F5]",
      )}
      onClick={props.onClick}
    >
      <span className={classNames("shrink-0", props.active ? "text-[#111827]" : "text-[#6B7280]")}>
        {props.icon}
      </span>
      <span className="font-semibold tracking-[0.01em]">{props.label}</span>
    </button>
  );
}

function Badge(props: { color: "gray" | "indigo" | "green" | "red" | "amber"; children: ReactNode }) {
  const cls =
    props.color === "indigo"
      ? "bg-[#EEF2FF] text-[#4F46E5]"
      : props.color === "green"
        ? "bg-[#ECFDF5] text-[#059669]"
        : props.color === "red"
          ? "bg-[#FEF2F2] text-[#DC2626]"
          : props.color === "amber"
            ? "bg-[#FFFBEB] text-[#D97706]"
            : "bg-[#F3F4F6] text-[#6B7280]";
  return (
    <span className={classNames("inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium", cls)}>
      {props.children}
    </span>
  );
}

type ProjectColumnKey = "name" | "status" | "updatedAt" | "actions";

type ProjectColumnWidths = Record<ProjectColumnKey, number>;

type ProjectResizeState = {
  key: ProjectColumnKey;
  startX: number;
  startWidth: number;
} | null;

const DEFAULT_PROJECT_COLUMN_WIDTHS: ProjectColumnWidths = {
  name: 320,
  status: 160,
  updatedAt: 180,
  actions: 120,
};

const PROJECT_COLUMN_WIDTHS_STORAGE_KEY = "multivac:project-column-widths";

const MIN_PROJECT_COLUMN_WIDTHS: ProjectColumnWidths = {
  name: 180,
  status: 120,
  updatedAt: 140,
  actions: 100,
};

type TaskColumnKey = "name" | "project" | "priority" | "status" | "dueAt" | "actions";

type TaskColumnWidths = Record<TaskColumnKey, number>;

type TaskResizeState = {
  key: TaskColumnKey;
  startX: number;
  startWidth: number;
} | null;

const DEFAULT_TASK_COLUMN_WIDTHS: TaskColumnWidths = {
  name: 280,
  project: 220,
  priority: 120,
  status: 140,
  dueAt: 180,
  actions: 120,
};

const TASK_COLUMN_WIDTHS_STORAGE_KEY = "multivac:task-column-widths";

const MIN_TASK_COLUMN_WIDTHS: TaskColumnWidths = {
  name: 180,
  project: 160,
  priority: 100,
  status: 120,
  dueAt: 140,
  actions: 100,
};

type InboxColumnKey = "name" | "description" | "updatedAt" | "actions";

type InboxColumnWidths = Record<InboxColumnKey, number>;

type InboxResizeState = {
  key: InboxColumnKey;
  startX: number;
  startWidth: number;
} | null;

const DEFAULT_INBOX_COLUMN_WIDTHS: InboxColumnWidths = {
  name: 260,
  description: 360,
  updatedAt: 180,
  actions: 120,
};

const INBOX_COLUMN_WIDTHS_STORAGE_KEY = "multivac:inbox-column-widths";

const MIN_INBOX_COLUMN_WIDTHS: InboxColumnWidths = {
  name: 180,
  description: 220,
  updatedAt: 140,
  actions: 100,
};

type SomedayColumnKey = "name" | "description" | "updatedAt" | "actions";

type SomedayColumnWidths = Record<SomedayColumnKey, number>;

type SomedayResizeState = {
  key: SomedayColumnKey;
  startX: number;
  startWidth: number;
} | null;

const DEFAULT_SOMEDAY_COLUMN_WIDTHS: SomedayColumnWidths = {
  name: 260,
  description: 360,
  updatedAt: 180,
  actions: 120,
};

const SOMEDAY_COLUMN_WIDTHS_STORAGE_KEY = "multivac:someday-column-widths";

const MIN_SOMEDAY_COLUMN_WIDTHS: SomedayColumnWidths = {
  name: 180,
  description: 220,
  updatedAt: 140,
  actions: 100,
};

type WaitingListColumnKey = "name" | "details" | "owner" | "expectedAt" | "updatedAt" | "actions";

type WaitingListColumnWidths = Record<WaitingListColumnKey, number>;

type WaitingListResizeState = {
  key: WaitingListColumnKey;
  startX: number;
  startWidth: number;
} | null;

const DEFAULT_WAITING_LIST_COLUMN_WIDTHS: WaitingListColumnWidths = {
  name: 220,
  details: 320,
  owner: 160,
  expectedAt: 200,
  updatedAt: 180,
  actions: 120,
};

const WAITING_LIST_COLUMN_WIDTHS_STORAGE_KEY = "multivac:waiting-list-column-widths";

const MIN_WAITING_LIST_COLUMN_WIDTHS: WaitingListColumnWidths = {
  name: 180,
  details: 220,
  owner: 120,
  expectedAt: 180,
  updatedAt: 140,
  actions: 100,
};

function normalizeProjectColumnWidths(value: unknown): ProjectColumnWidths {
  const widths = typeof value === "object" && value !== null ? value as Partial<Record<ProjectColumnKey, unknown>> : {};
  return {
    name: typeof widths.name === "number" ? Math.max(MIN_PROJECT_COLUMN_WIDTHS.name, widths.name) : DEFAULT_PROJECT_COLUMN_WIDTHS.name,
    status: typeof widths.status === "number" ? Math.max(MIN_PROJECT_COLUMN_WIDTHS.status, widths.status) : DEFAULT_PROJECT_COLUMN_WIDTHS.status,
    updatedAt: typeof widths.updatedAt === "number" ? Math.max(MIN_PROJECT_COLUMN_WIDTHS.updatedAt, widths.updatedAt) : DEFAULT_PROJECT_COLUMN_WIDTHS.updatedAt,
    actions: typeof widths.actions === "number" ? Math.max(MIN_PROJECT_COLUMN_WIDTHS.actions, widths.actions) : DEFAULT_PROJECT_COLUMN_WIDTHS.actions,
  };
}

function loadProjectColumnWidths(): ProjectColumnWidths {
  if (typeof window === "undefined") return DEFAULT_PROJECT_COLUMN_WIDTHS;
  try {
    const raw = window.localStorage.getItem(PROJECT_COLUMN_WIDTHS_STORAGE_KEY);
    if (!raw) return DEFAULT_PROJECT_COLUMN_WIDTHS;
    return normalizeProjectColumnWidths(JSON.parse(raw));
  } catch {
    return DEFAULT_PROJECT_COLUMN_WIDTHS;
  }
}

function normalizeTaskColumnWidths(value: unknown): TaskColumnWidths {
  const widths = typeof value === "object" && value !== null ? value as Partial<Record<TaskColumnKey, unknown>> : {};
  return {
    name: typeof widths.name === "number" ? Math.max(MIN_TASK_COLUMN_WIDTHS.name, widths.name) : DEFAULT_TASK_COLUMN_WIDTHS.name,
    project: typeof widths.project === "number" ? Math.max(MIN_TASK_COLUMN_WIDTHS.project, widths.project) : DEFAULT_TASK_COLUMN_WIDTHS.project,
    priority: typeof widths.priority === "number" ? Math.max(MIN_TASK_COLUMN_WIDTHS.priority, widths.priority) : DEFAULT_TASK_COLUMN_WIDTHS.priority,
    status: typeof widths.status === "number" ? Math.max(MIN_TASK_COLUMN_WIDTHS.status, widths.status) : DEFAULT_TASK_COLUMN_WIDTHS.status,
    dueAt: typeof widths.dueAt === "number" ? Math.max(MIN_TASK_COLUMN_WIDTHS.dueAt, widths.dueAt) : DEFAULT_TASK_COLUMN_WIDTHS.dueAt,
    actions: typeof widths.actions === "number" ? Math.max(MIN_TASK_COLUMN_WIDTHS.actions, widths.actions) : DEFAULT_TASK_COLUMN_WIDTHS.actions,
  };
}

function loadTaskColumnWidths(): TaskColumnWidths {
  if (typeof window === "undefined") return DEFAULT_TASK_COLUMN_WIDTHS;
  try {
    const raw = window.localStorage.getItem(TASK_COLUMN_WIDTHS_STORAGE_KEY);
    if (!raw) return DEFAULT_TASK_COLUMN_WIDTHS;
    return normalizeTaskColumnWidths(JSON.parse(raw));
  } catch {
    return DEFAULT_TASK_COLUMN_WIDTHS;
  }
}

function normalizeInboxColumnWidths(value: unknown): InboxColumnWidths {
  const widths = typeof value === "object" && value !== null ? value as Partial<Record<InboxColumnKey, unknown>> : {};
  return {
    name: typeof widths.name === "number" ? Math.max(MIN_INBOX_COLUMN_WIDTHS.name, widths.name) : DEFAULT_INBOX_COLUMN_WIDTHS.name,
    description: typeof widths.description === "number" ? Math.max(MIN_INBOX_COLUMN_WIDTHS.description, widths.description) : DEFAULT_INBOX_COLUMN_WIDTHS.description,
    updatedAt: typeof widths.updatedAt === "number" ? Math.max(MIN_INBOX_COLUMN_WIDTHS.updatedAt, widths.updatedAt) : DEFAULT_INBOX_COLUMN_WIDTHS.updatedAt,
    actions: typeof widths.actions === "number" ? Math.max(MIN_INBOX_COLUMN_WIDTHS.actions, widths.actions) : DEFAULT_INBOX_COLUMN_WIDTHS.actions,
  };
}

function loadInboxColumnWidths(): InboxColumnWidths {
  if (typeof window === "undefined") return DEFAULT_INBOX_COLUMN_WIDTHS;
  try {
    const raw = window.localStorage.getItem(INBOX_COLUMN_WIDTHS_STORAGE_KEY);
    if (!raw) return DEFAULT_INBOX_COLUMN_WIDTHS;
    return normalizeInboxColumnWidths(JSON.parse(raw));
  } catch {
    return DEFAULT_INBOX_COLUMN_WIDTHS;
  }
}

function normalizeSomedayColumnWidths(value: unknown): SomedayColumnWidths {
  const widths = typeof value === "object" && value !== null ? value as Partial<Record<SomedayColumnKey, unknown>> : {};
  return {
    name: typeof widths.name === "number" ? Math.max(MIN_SOMEDAY_COLUMN_WIDTHS.name, widths.name) : DEFAULT_SOMEDAY_COLUMN_WIDTHS.name,
    description: typeof widths.description === "number" ? Math.max(MIN_SOMEDAY_COLUMN_WIDTHS.description, widths.description) : DEFAULT_SOMEDAY_COLUMN_WIDTHS.description,
    updatedAt: typeof widths.updatedAt === "number" ? Math.max(MIN_SOMEDAY_COLUMN_WIDTHS.updatedAt, widths.updatedAt) : DEFAULT_SOMEDAY_COLUMN_WIDTHS.updatedAt,
    actions: typeof widths.actions === "number" ? Math.max(MIN_SOMEDAY_COLUMN_WIDTHS.actions, widths.actions) : DEFAULT_SOMEDAY_COLUMN_WIDTHS.actions,
  };
}

function loadSomedayColumnWidths(): SomedayColumnWidths {
  if (typeof window === "undefined") return DEFAULT_SOMEDAY_COLUMN_WIDTHS;
  try {
    const raw = window.localStorage.getItem(SOMEDAY_COLUMN_WIDTHS_STORAGE_KEY);
    if (!raw) return DEFAULT_SOMEDAY_COLUMN_WIDTHS;
    return normalizeSomedayColumnWidths(JSON.parse(raw));
  } catch {
    return DEFAULT_SOMEDAY_COLUMN_WIDTHS;
  }
}

function normalizeWaitingListColumnWidths(value: unknown): WaitingListColumnWidths {
  const widths = typeof value === "object" && value !== null ? value as Partial<Record<WaitingListColumnKey, unknown>> : {};
  return {
    name: typeof widths.name === "number" ? Math.max(MIN_WAITING_LIST_COLUMN_WIDTHS.name, widths.name) : DEFAULT_WAITING_LIST_COLUMN_WIDTHS.name,
    details: typeof widths.details === "number" ? Math.max(MIN_WAITING_LIST_COLUMN_WIDTHS.details, widths.details) : DEFAULT_WAITING_LIST_COLUMN_WIDTHS.details,
    owner: typeof widths.owner === "number" ? Math.max(MIN_WAITING_LIST_COLUMN_WIDTHS.owner, widths.owner) : DEFAULT_WAITING_LIST_COLUMN_WIDTHS.owner,
    expectedAt: typeof widths.expectedAt === "number" ? Math.max(MIN_WAITING_LIST_COLUMN_WIDTHS.expectedAt, widths.expectedAt) : DEFAULT_WAITING_LIST_COLUMN_WIDTHS.expectedAt,
    updatedAt: typeof widths.updatedAt === "number" ? Math.max(MIN_WAITING_LIST_COLUMN_WIDTHS.updatedAt, widths.updatedAt) : DEFAULT_WAITING_LIST_COLUMN_WIDTHS.updatedAt,
    actions: typeof widths.actions === "number" ? Math.max(MIN_WAITING_LIST_COLUMN_WIDTHS.actions, widths.actions) : DEFAULT_WAITING_LIST_COLUMN_WIDTHS.actions,
  };
}

function loadWaitingListColumnWidths(): WaitingListColumnWidths {
  if (typeof window === "undefined") return DEFAULT_WAITING_LIST_COLUMN_WIDTHS;
  try {
    const raw = window.localStorage.getItem(WAITING_LIST_COLUMN_WIDTHS_STORAGE_KEY);
    if (!raw) return DEFAULT_WAITING_LIST_COLUMN_WIDTHS;
    return normalizeWaitingListColumnWidths(JSON.parse(raw));
  } catch {
    return DEFAULT_WAITING_LIST_COLUMN_WIDTHS;
  }
}

function ProjectsPage(props: {
  items: Project[];
  loading: boolean;
  statusFilter: ProjectStatus | "";
  pageSize: number;
  offset: number;
  hasNext: boolean;
  paginationEnabled: boolean;
  onStatusFilter: (v: ProjectStatus | "") => void;
  onPageSizeChange: (v: number) => void;
  onPrevPage: () => void;
  onNextPage: () => void;
  onRefresh: () => void;
  onOpen: (id: string) => void;
}) {
  const [columnWidths, setColumnWidths] = useState<ProjectColumnWidths>(() => loadProjectColumnWidths());

  useEffect(() => {
    window.localStorage.setItem(PROJECT_COLUMN_WIDTHS_STORAGE_KEY, JSON.stringify(columnWidths));
  }, [columnWidths]);

  function startResize(key: ProjectColumnKey, event: ReactPointerEvent<HTMLDivElement>) {
    event.preventDefault();

    const state: ProjectResizeState = {
      key,
      startX: event.clientX,
      startWidth: columnWidths[key],
    };

    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    const onPointerMove = (moveEvent: PointerEvent) => {
      const delta = moveEvent.clientX - state.startX;
      const nextWidth = Math.max(MIN_PROJECT_COLUMN_WIDTHS[key], state.startWidth + delta);
      setColumnWidths((current) => ({ ...current, [key]: nextWidth }));
    };

    const stopResize = () => {
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      window.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", stopResize);
    };

    window.addEventListener("pointermove", onPointerMove);
    window.addEventListener("pointerup", stopResize);
  }

  function headerCell(label: string, key: ProjectColumnKey, alignRight?: boolean) {
    const showResizeMarker = key !== "actions";
    return (
      <th className={classNames("relative px-4 py-3", alignRight && "text-right")}>
        <div className={classNames("relative", alignRight && "pr-3")}>{label}</div>
        {showResizeMarker ? (
          <div className="pointer-events-none absolute right-0 top-1/2 h-5 -translate-y-1/2 border-r border-[#D1D5DB]" />
        ) : null}
        <div
          role="separator"
          aria-orientation="vertical"
          className="absolute right-0 top-0 h-full w-2 cursor-col-resize hover:bg-[#EEF2FF]/80"
          onPointerDown={(event) => startResize(key, event)}
        />
      </th>
    );
  }

  return (
    <div className="mx-auto grid w-full max-w-7xl gap-3">
      <div className="flex items-center justify-between">
        <div className="text-lg font-semibold text-[#111827]">项目列表</div>
        <div className="flex items-center gap-2">
          <select
            className="rounded-md border border-[#E6E8F0] bg-white px-2 py-1 text-sm text-[#374151]"
            value={props.statusFilter}
            onChange={(e) => props.onStatusFilter(e.target.value as ProjectStatus | "")}
          >
            <option value="">全部状态</option>
            {PROJECT_STATUSES.map((s) => (
              <option key={s} value={s}>
                {projectStatusLabel(s)}
              </option>
            ))}
          </select>
          <button
            className="flex items-center gap-1 rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] hover:bg-[#F5F6FA]"
            type="button"
            onClick={props.onRefresh}
          >
            <IconRefresh />
            刷新
          </button>
        </div>
      </div>

      <div className="overflow-x-auto overflow-y-hidden rounded-lg border border-[#E6E8F0] bg-white">
        <table className="w-full min-w-[780px] table-fixed text-left text-sm">
          <colgroup>
            <col style={{ width: columnWidths.name }} />
            <col style={{ width: columnWidths.status }} />
            <col style={{ width: columnWidths.updatedAt }} />
            <col style={{ width: columnWidths.actions }} />
          </colgroup>
          <thead className="bg-[#F9FAFB] text-xs text-[#6B7280]">
            <tr>
              {headerCell("项目名称", "name")}
              {headerCell("状态", "status")}
              {headerCell("更新时间", "updatedAt")}
              {headerCell("操作", "actions", true)}
            </tr>
          </thead>
          <tbody>
            {props.items.length === 0 && !props.loading ? (
              <tr>
                <td colSpan={4} className="px-4 py-10 text-center text-[#6B7280]">
                  暂无 Project
                </td>
              </tr>
            ) : null}
            {props.items.map((p) => (
              <tr key={p.id} className="border-t border-[#E6E8F0] hover:bg-[#F9FAFB]">
                <td className="px-4 py-3">
                  <button
                    type="button"
                    className="font-medium text-[#111827] hover:underline"
                    onClick={() => props.onOpen(p.id)}
                  >
                    {p.name}
                  </button>
                </td>
                <td className="px-4 py-3">
                  <Badge color={p.status === "Active" ? "indigo" : p.status === "Completed" ? "green" : p.status === "Archived" ? "gray" : "gray"}>
                    <span className={classNames("h-2 w-2 rounded-full", p.status === "Active" ? "bg-[#4F46E5]" : p.status === "Completed" ? "bg-[#10B981]" : "bg-[#9CA3AF]")} />
                    {projectStatusLabel(p.status)}
                  </Badge>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">{formatDate(p.updatedAt)}</td>
                <td className="px-4 py-3 text-right">
                  <button
                    type="button"
                    className="text-sm font-medium text-[#4F46E5] hover:underline"
                    onClick={() => props.onOpen(p.id)}
                  >
                    详情 →
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {props.paginationEnabled ? (
        <div className="flex items-center justify-between rounded-lg border border-[#E6E8F0] bg-white px-4 py-3 text-sm">
          <div className="flex items-center gap-2 text-[#6B7280]">
            <span>每页显示</span>
            <select
              className="rounded-md border border-[#E6E8F0] bg-white px-2 py-1 text-sm text-[#374151]"
              value={props.pageSize}
              onChange={(e) => props.onPageSizeChange(Number(e.target.value))}
              disabled={props.loading}
            >
              {PAGE_SIZE_OPTIONS.map((size) => (
                <option key={size} value={size}>
                  {size}
                </option>
              ))}
            </select>
            <span>第 {pageNumber(props.offset, props.pageSize)} 页</span>
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
              onClick={props.onPrevPage}
              disabled={props.loading || props.offset === 0}
            >
              上一页
            </button>
            <button
              type="button"
              className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
              onClick={props.onNextPage}
              disabled={props.loading || !props.hasNext}
            >
              下一页
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function InboxesPage(props: {
  items: Inbox[];
  loading: boolean;
  pageSize: number;
  offset: number;
  hasNext: boolean;
  onPageSizeChange: (v: number) => void;
  onPrevPage: () => void;
  onNextPage: () => void;
  onRefresh: () => void;
  onOpen: (id: string, inbox?: Inbox) => void;
}) {
  const [columnWidths, setColumnWidths] = useState<InboxColumnWidths>(() => loadInboxColumnWidths());

  useEffect(() => {
    window.localStorage.setItem(INBOX_COLUMN_WIDTHS_STORAGE_KEY, JSON.stringify(columnWidths));
  }, [columnWidths]);

  function startResize(key: InboxColumnKey, event: ReactPointerEvent<HTMLDivElement>) {
    event.preventDefault();

    const state: InboxResizeState = {
      key,
      startX: event.clientX,
      startWidth: columnWidths[key],
    };

    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    const onPointerMove = (moveEvent: PointerEvent) => {
      const delta = moveEvent.clientX - state.startX;
      const nextWidth = Math.max(MIN_INBOX_COLUMN_WIDTHS[key], state.startWidth + delta);
      setColumnWidths((current) => ({ ...current, [key]: nextWidth }));
    };

    const stopResize = () => {
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      window.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", stopResize);
    };

    window.addEventListener("pointermove", onPointerMove);
    window.addEventListener("pointerup", stopResize);
  }

  function headerCell(label: string, key: InboxColumnKey, alignRight?: boolean) {
    const showResizeMarker = key !== "actions";
    return (
      <th className={classNames("relative px-4 py-3", alignRight && "text-right")}>
        <div className={classNames("relative", alignRight && "pr-3")}>{label}</div>
        {showResizeMarker ? (
          <div className="pointer-events-none absolute right-0 top-1/2 h-5 -translate-y-1/2 border-r border-[#D1D5DB]" />
        ) : null}
        <div
          role="separator"
          aria-orientation="vertical"
          className="absolute right-0 top-0 h-full w-2 cursor-col-resize hover:bg-[#EEF2FF]/80"
          onPointerDown={(event) => startResize(key, event)}
        />
      </th>
    );
  }

  return (
    <div className="mx-auto grid w-full max-w-7xl gap-3">
      <div className="flex items-center justify-between">
        <div className="text-lg font-semibold text-[#111827]">收集箱列表</div>
        <div className="flex items-center gap-2">
          <button
            className="flex items-center gap-1 rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] hover:bg-[#F5F6FA]"
            type="button"
            onClick={props.onRefresh}
          >
            <IconRefresh />
            刷新
          </button>
        </div>
      </div>

      <div className="overflow-x-auto overflow-y-hidden rounded-lg border border-[#E6E8F0] bg-white">
        <table className="w-full min-w-[920px] table-fixed text-left text-sm">
          <colgroup>
            <col style={{ width: columnWidths.name }} />
            <col style={{ width: columnWidths.description }} />
            <col style={{ width: columnWidths.updatedAt }} />
            <col style={{ width: columnWidths.actions }} />
          </colgroup>
          <thead className="bg-[#F9FAFB] text-xs text-[#6B7280]">
            <tr>
              {headerCell("名称", "name")}
              {headerCell("详细描述", "description")}
              {headerCell("更新时间", "updatedAt")}
              {headerCell("操作", "actions", true)}
            </tr>
          </thead>
          <tbody>
            {props.items.length === 0 && !props.loading ? (
              <tr>
                <td colSpan={4} className="px-4 py-10 text-center text-[#6B7280]">
                  暂无收集箱
                </td>
              </tr>
            ) : null}
            {props.items.map((inbox) => (
              <tr key={inbox.id} className="border-t border-[#E6E8F0] hover:bg-[#F9FAFB]">
                <td className="px-4 py-3">
                  <div className="font-medium text-[#111827]">{inbox.name}</div>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">
                  <div className="line-clamp-2 break-words">{inbox.description || "-"}</div>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">{formatDate(inbox.updatedAt)}</td>
                <td className="px-4 py-3 text-right">
                  <button
                    type="button"
                    className="text-sm font-medium text-[#4F46E5] hover:underline"
                    onClick={() => props.onOpen(inbox.id, inbox)}
                  >
                    详情 →
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between rounded-lg border border-[#E6E8F0] bg-white px-4 py-3 text-sm">
        <div className="flex items-center gap-2 text-[#6B7280]">
          <span>每页显示</span>
          <select
            className="rounded-md border border-[#E6E8F0] bg-white px-2 py-1 text-sm text-[#374151]"
            value={props.pageSize}
            onChange={(e) => props.onPageSizeChange(Number(e.target.value))}
            disabled={props.loading}
          >
            {PAGE_SIZE_OPTIONS.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
          <span>第 {pageNumber(props.offset, props.pageSize)} 页</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
            onClick={props.onPrevPage}
            disabled={props.loading || props.offset === 0}
          >
            上一页
          </button>
          <button
            type="button"
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
            onClick={props.onNextPage}
            disabled={props.loading || !props.hasNext}
          >
            下一页
          </button>
        </div>
      </div>
    </div>
  );
}

function SomedaysPage(props: {
  items: Someday[];
  loading: boolean;
  pageSize: number;
  offset: number;
  hasNext: boolean;
  onPageSizeChange: (v: number) => void;
  onPrevPage: () => void;
  onNextPage: () => void;
  onRefresh: () => void;
  onOpen: (id: string, someday?: Someday) => void;
}) {
  const [columnWidths, setColumnWidths] = useState<SomedayColumnWidths>(() => loadSomedayColumnWidths());

  useEffect(() => {
    window.localStorage.setItem(SOMEDAY_COLUMN_WIDTHS_STORAGE_KEY, JSON.stringify(columnWidths));
  }, [columnWidths]);

  function startResize(key: SomedayColumnKey, event: ReactPointerEvent<HTMLDivElement>) {
    event.preventDefault();

    const state: SomedayResizeState = {
      key,
      startX: event.clientX,
      startWidth: columnWidths[key],
    };

    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    const onPointerMove = (moveEvent: PointerEvent) => {
      const delta = moveEvent.clientX - state.startX;
      const nextWidth = Math.max(MIN_SOMEDAY_COLUMN_WIDTHS[key], state.startWidth + delta);
      setColumnWidths((current) => ({ ...current, [key]: nextWidth }));
    };

    const stopResize = () => {
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      window.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", stopResize);
    };

    window.addEventListener("pointermove", onPointerMove);
    window.addEventListener("pointerup", stopResize);
  }

  function headerCell(label: string, key: SomedayColumnKey, alignRight?: boolean) {
    const showResizeMarker = key !== "actions";
    return (
      <th className={classNames("relative px-4 py-3", alignRight && "text-right")}>
        <div className={classNames("relative", alignRight && "pr-3")}>{label}</div>
        {showResizeMarker ? (
          <div className="pointer-events-none absolute right-0 top-1/2 h-5 -translate-y-1/2 border-r border-[#D1D5DB]" />
        ) : null}
        <div
          role="separator"
          aria-orientation="vertical"
          className="absolute right-0 top-0 h-full w-2 cursor-col-resize hover:bg-[#EEF2FF]/80"
          onPointerDown={(event) => startResize(key, event)}
        />
      </th>
    );
  }

  return (
    <div className="mx-auto grid w-full max-w-7xl gap-3">
      <div className="flex items-center justify-between">
        <div className="text-lg font-semibold text-[#111827]">将来/也许列表</div>
        <div className="flex items-center gap-2">
          <button
            className="flex items-center gap-1 rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] hover:bg-[#F5F6FA]"
            type="button"
            onClick={props.onRefresh}
          >
            <IconRefresh />
            刷新
          </button>
        </div>
      </div>

      <div className="overflow-x-auto overflow-y-hidden rounded-lg border border-[#E6E8F0] bg-white">
        <table className="w-full min-w-[920px] table-fixed text-left text-sm">
          <colgroup>
            <col style={{ width: columnWidths.name }} />
            <col style={{ width: columnWidths.description }} />
            <col style={{ width: columnWidths.updatedAt }} />
            <col style={{ width: columnWidths.actions }} />
          </colgroup>
          <thead className="bg-[#F9FAFB] text-xs text-[#6B7280]">
            <tr>
              {headerCell("任务名称", "name")}
              {headerCell("描述", "description")}
              {headerCell("更新时间", "updatedAt")}
              {headerCell("操作", "actions", true)}
            </tr>
          </thead>
          <tbody>
            {props.items.length === 0 && !props.loading ? (
              <tr>
                <td colSpan={4} className="px-4 py-10 text-center text-[#6B7280]">
                  暂无将来/也许
                </td>
              </tr>
            ) : null}
            {props.items.map((someday) => (
              <tr key={someday.id} className="border-t border-[#E6E8F0] hover:bg-[#F9FAFB]">
                <td className="px-4 py-3">
                  <div className="font-medium text-[#111827]">{someday.name}</div>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">
                  <div className="line-clamp-2 break-words">{someday.description || "-"}</div>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">{formatDate(someday.updatedAt)}</td>
                <td className="px-4 py-3 text-right">
                  <button
                    type="button"
                    className="text-sm font-medium text-[#4F46E5] hover:underline"
                    onClick={() => props.onOpen(someday.id, someday)}
                  >
                    详情 →
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between rounded-lg border border-[#E6E8F0] bg-white px-4 py-3 text-sm">
        <div className="flex items-center gap-2 text-[#6B7280]">
          <span>每页显示</span>
          <select
            className="rounded-md border border-[#E6E8F0] bg-white px-2 py-1 text-sm text-[#374151]"
            value={props.pageSize}
            onChange={(e) => props.onPageSizeChange(Number(e.target.value))}
            disabled={props.loading}
          >
            {PAGE_SIZE_OPTIONS.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
          <span>第 {pageNumber(props.offset, props.pageSize)} 页</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
            onClick={props.onPrevPage}
            disabled={props.loading || props.offset === 0}
          >
            上一页
          </button>
          <button
            type="button"
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
            onClick={props.onNextPage}
            disabled={props.loading || !props.hasNext}
          >
            下一页
          </button>
        </div>
      </div>
    </div>
  );
}

function TasksPageNew(props: {
  projects: Project[];
  projectId: string;
  status: TaskStatus | "";
  search: string;
  version: number;
  highlightTaskId: string | null;
  onProjectId: (v: string) => void;
  onStatus: (v: TaskStatus | "") => void;
  onOpen: (id: string, t?: Task) => void;
}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [items, setItems] = useState<Task[]>([]);
  const [offset, setOffset] = useState(0);
  const [pageSize, setPageSize] = useState(PAGE_SIZE_OPTIONS[0]);
  const [hasNext, setHasNext] = useState(false);
  const [sortDir, setSortDir] = useState<SortDir | null>("Asc");
  const [columnWidths, setColumnWidths] = useState<TaskColumnWidths>(() => loadTaskColumnWidths());

  useEffect(() => {
    window.localStorage.setItem(TASK_COLUMN_WIDTHS_STORAGE_KEY, JSON.stringify(columnWidths));
  }, [columnWidths]);

  function startResize(key: TaskColumnKey, event: ReactPointerEvent<HTMLDivElement>) {
    event.preventDefault();

    const state: TaskResizeState = {
      key,
      startX: event.clientX,
      startWidth: columnWidths[key],
    };

    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    const onPointerMove = (moveEvent: PointerEvent) => {
      const delta = moveEvent.clientX - state.startX;
      const nextWidth = Math.max(MIN_TASK_COLUMN_WIDTHS[key], state.startWidth + delta);
      setColumnWidths((current) => ({ ...current, [key]: nextWidth }));
    };

    const stopResize = () => {
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      window.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", stopResize);
    };

    window.addEventListener("pointermove", onPointerMove);
    window.addEventListener("pointerup", stopResize);
  }

  function toggleDueAtSort() {
    setSortDir((current: SortDir | null) => {
      if (current === "Asc") return "Desc";
      if (current === "Desc") return null;
      return "Asc";
    });
  }

  function headerCell(label: string, key: TaskColumnKey, alignRight?: boolean) {
    const showResizeMarker = key !== "actions";
    const isDueAt = key === "dueAt";
    return (
      <th className={classNames("relative px-4 py-3", alignRight && "text-right")}>
        <div className={classNames("relative", alignRight && "pr-3")}>
          {isDueAt ? (
            <button
              type="button"
              className="inline-flex items-center gap-1 text-inherit hover:text-[#374151]"
              onClick={toggleDueAtSort}
            >
              <span>{label}</span>
              <IconSort dir={sortDir} />
            </button>
          ) : (
            label
          )}
        </div>
        {showResizeMarker ? (
          <div className="pointer-events-none absolute right-0 top-1/2 h-5 -translate-y-1/2 border-r border-[#D1D5DB]" />
        ) : null}
        <div
          role="separator"
          aria-orientation="vertical"
          className="absolute right-0 top-0 h-full w-2 cursor-col-resize hover:bg-[#EEF2FF]/80"
          onPointerDown={(event) => startResize(key, event)}
        />
      </th>
    );
  }

  async function refresh() {
    setLoading(true);
    setError("");
    try {
      const list = await listTasks({
        projectId: props.projectId || undefined,
        status: props.status || undefined,
        search: props.search || undefined,
        sortBy: sortDir ? "DueAt" : undefined,
        sortDir: sortDir || undefined,
        limit: pageSize + 1,
        offset,
      });
      setItems(list.slice(0, pageSize));
      setHasNext(list.length > pageSize);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    setOffset(0);
  }, [props.projectId, props.status, props.search, sortDir, pageSize]);

  useEffect(() => {
    void refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.projectId, props.status, props.search, props.version, sortDir, offset, pageSize]);

  const projectNameById = useMemo(() => {
    return new Map(props.projects.map((p) => [p.id, p.name]));
  }, [props.projects]);

  useEffect(() => {
    if (!props.highlightTaskId) return;
    const timer = window.setTimeout(() => {
      const row = document.getElementById(`task-row-${props.highlightTaskId}`);
      row?.scrollIntoView({ behavior: "smooth", block: "center" });
    }, 50);
    return () => window.clearTimeout(timer);
  }, [props.highlightTaskId, items]);

  return (
    <div className="mx-auto grid w-full max-w-7xl gap-3">
      <div className="flex items-center justify-between">
        <div className="text-lg font-semibold text-[#111827]">任务列表</div>
        <div className="flex items-center gap-2">
          <select
            className="rounded-md border border-[#E6E8F0] bg-white px-2 py-1 text-sm text-[#374151]"
            value={props.projectId}
            onChange={(e) => props.onProjectId(e.target.value)}
          >
            <option value="">全部项目</option>
            {props.projects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
          <select
            className="rounded-md border border-[#E6E8F0] bg-white px-2 py-1 text-sm text-[#374151]"
            value={props.status}
            onChange={(e) => props.onStatus(e.target.value as TaskStatus | "")}
          >
            <option value="">全部状态</option>
            {TASK_STATUSES.map((s) => (
              <option key={s} value={s}>
                {taskStatusLabel(s)}
              </option>
            ))}
          </select>
          <button
            className="flex items-center gap-1 rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] hover:bg-[#F5F6FA]"
            type="button"
            onClick={() => void refresh()}
          >
            <IconRefresh />
            刷新
          </button>
        </div>
      </div>

      {error ? (
        <div className="rounded-md border border-[#FCA5A5] bg-[#FEF2F2] px-3 py-2 text-sm text-[#B91C1C]">
          {error}
        </div>
      ) : null}

      <div className="overflow-x-auto overflow-y-hidden rounded-lg border border-[#E6E8F0] bg-white">
        <table className="w-full min-w-[980px] table-fixed text-left text-sm">
          <colgroup>
            <col style={{ width: columnWidths.name }} />
            <col style={{ width: columnWidths.project }} />
            <col style={{ width: columnWidths.priority }} />
            <col style={{ width: columnWidths.status }} />
            <col style={{ width: columnWidths.dueAt }} />
            <col style={{ width: columnWidths.actions }} />
          </colgroup>
          <thead className="bg-[#F9FAFB] text-xs text-[#6B7280]">
            <tr>
              {headerCell("任务名称", "name")}
              {headerCell("所属项目", "project")}
              {headerCell("优先级", "priority")}
              {headerCell("状态", "status")}
              {headerCell("截止日期", "dueAt")}
              {headerCell("操作", "actions", true)}
            </tr>
          </thead>
          <tbody>
            {items.length === 0 && !loading ? (
              <tr>
                <td colSpan={6} className="px-4 py-10 text-center text-[#6B7280]">
                  暂无 Task
                </td>
              </tr>
            ) : null}
            {items.map((t) => (
              <tr
                key={t.id}
                id={`task-row-${t.id}`}
                className={classNames(
                  "border-t border-[#E6E8F0] hover:bg-[#F9FAFB]",
                  props.highlightTaskId === t.id && "bg-[#EEF2FF] transition-colors",
                )}
              >
                <td className="px-4 py-3">
                  <div className="font-medium text-[#111827]">{t.name}</div>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">{projectNameById.get(t.projectId ?? "") ?? "-"}</td>
                <td className="px-4 py-3">
                  <Badge
                    color={t.priority === "P0" || t.priority === "High" ? "red" : t.priority === "Medium" ? "amber" : "gray"}
                  >
                    {taskPriorityLabel(t.priority)}
                  </Badge>
                </td>
                <td className="px-4 py-3">
                  <Badge color={t.status === "InProgress" ? "indigo" : t.status === "Done" ? "green" : t.status === "Canceled" ? "gray" : "gray"}>
                    <span className={classNames("h-2 w-2 rounded-full", t.status === "InProgress" ? "bg-[#4F46E5]" : t.status === "Done" ? "bg-[#10B981]" : "bg-[#9CA3AF]")} />
                    {taskStatusLabel(t.status)}
                  </Badge>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">
                  <span className="inline-flex items-center gap-1">
                    <span className="text-[#9CA3AF]">
                      <IconClock />
                    </span>
                    {t.dueAt ? formatDate(t.dueAt) : "-"}
                  </span>
                </td>
                <td className="px-4 py-3 text-right">
                  <button
                    type="button"
                    className="text-sm font-medium text-[#4F46E5] hover:underline"
                    onClick={() => props.onOpen(t.id, t)}
                  >
                    详情 →
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between rounded-lg border border-[#E6E8F0] bg-white px-4 py-3 text-sm">
        <div className="flex items-center gap-2 text-[#6B7280]">
          <span>每页显示</span>
          <select
            className="rounded-md border border-[#E6E8F0] bg-white px-2 py-1 text-sm text-[#374151]"
            value={pageSize}
            onChange={(e) => setPageSize(Number(e.target.value))}
            disabled={loading}
          >
            {PAGE_SIZE_OPTIONS.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
          <span>第 {pageNumber(offset, pageSize)} 页</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
            onClick={() => setOffset((v) => Math.max(0, v - pageSize))}
            disabled={loading || offset === 0}
          >
            上一页
          </button>
          <button
            type="button"
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
            onClick={() => setOffset((v) => v + pageSize)}
            disabled={loading || !hasNext}
          >
            下一页
          </button>
        </div>
      </div>
    </div>
  );
}

function Drawer(props: { title: string; action: ReactNode; onClose: () => void; children: ReactNode }) {
  return (
    <div className="fixed inset-0 z-50 flex">
      <div className="flex-1 bg-black/20" onClick={props.onClose} />
      <div className="h-full w-[420px] border-l border-[#E6E8F0] bg-white shadow-xl">
        <div className="flex items-center justify-between border-b border-[#E6E8F0] px-4 py-3">
          <div className="text-sm font-semibold text-[#111827]">{props.title}</div>
          <div className="flex items-center gap-2">
            {props.action}
            <button
              className="flex items-center justify-center rounded-md border border-[#E6E8F0] bg-white p-2 text-[#6B7280] hover:bg-[#F5F6FA]"
              type="button"
              onClick={props.onClose}
              aria-label="close"
            >
              <IconClose />
            </button>
          </div>
        </div>
        <div className="h-[calc(100%-52px)] overflow-auto">{props.children}</div>
      </div>
    </div>
  );
}

function Field(props: { label: string; children: ReactNode }) {
  return (
    <div className="grid gap-1">
      <div className="text-xs font-medium text-[#6B7280]">{props.label}</div>
      {props.children}
    </div>
  );
}

function TextInput(props: { value: string; onChange: (v: string) => void; placeholder?: string }) {
  return (
    <input
      className="w-full rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm text-[#111827] outline-none focus:ring-2 focus:ring-[#4F46E5]"
      value={props.value}
      placeholder={props.placeholder}
      onChange={(e) => props.onChange(e.target.value)}
    />
  );
}

function SearchableProjectSelect(props: {
  projects: Project[];
  value: string;
  onChange: (value: string) => void;
}) {
  const [query, setQuery] = useState("");
  const [open, setOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);

  const selectedProject = props.projects.find((p) => p.id === props.value) ?? null;

  const filteredProjects = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return props.projects;
    return props.projects.filter((p) => p.name.toLowerCase().includes(q));
  }, [props.projects, query]);

  const options = useMemo(() => {
    return [{ id: "", name: "不选择项目" }, ...filteredProjects];
  }, [filteredProjects]);

  useEffect(() => {
    if (!open) setQuery("");
  }, [open]);

  useEffect(() => {
    setActiveIndex(0);
  }, [query, open]);

  function renderHighlightedName(name: string) {
    const q = query.trim();
    if (!q) return name;
    const index = name.toLowerCase().indexOf(q.toLowerCase());
    if (index < 0) return name;
    return (
      <>
        {name.slice(0, index)}
        <span className="bg-[#FEF3C7] text-[#92400E]">{name.slice(index, index + q.length)}</span>
        {name.slice(index + q.length)}
      </>
    );
  }

  function selectOption(index: number) {
    const option = options[index];
    if (!option) return;
    props.onChange(option.id);
    setOpen(false);
  }

  return (
    <div className="relative grid gap-2">
      <input
        className="w-full rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm text-[#111827] outline-none focus:ring-2 focus:ring-[#4F46E5]"
        value={open ? query : selectedProject?.name ?? ""}
        placeholder="搜索项目..."
        onFocus={() => setOpen(true)}
        onBlur={() => {
          window.setTimeout(() => setOpen(false), 100);
        }}
        onChange={(e) => {
          setOpen(true);
          setQuery(e.target.value);
        }}
        onKeyDown={(e) => {
          if (!open && (e.key === "ArrowDown" || e.key === "Enter")) {
            setOpen(true);
            return;
          }
          if (!open) return;
          if (e.key === "ArrowDown") {
            e.preventDefault();
            setActiveIndex((v) => Math.min(options.length - 1, v + 1));
            return;
          }
          if (e.key === "ArrowUp") {
            e.preventDefault();
            setActiveIndex((v) => Math.max(0, v - 1));
            return;
          }
          if (e.key === "Enter") {
            e.preventDefault();
            selectOption(activeIndex);
            return;
          }
          if (e.key === "Escape") {
            e.preventDefault();
            setOpen(false);
          }
        }}
      />
      {open ? (
        <div className="max-h-56 overflow-auto rounded-md border border-[#E6E8F0] bg-white shadow-sm">
          {options.length === 1 && filteredProjects.length === 0 ? (
            <div className="px-3 py-2 text-sm text-[#6B7280]">没有匹配的项目</div>
          ) : (
            options.map((option, index) => (
              <button
                key={option.id || "none"}
                type="button"
                className={classNames(
                  "block w-full px-3 py-2 text-left text-sm hover:bg-[#F5F6FA]",
                  props.value === option.id && "text-[#4F46E5]",
                  activeIndex === index && "bg-[#EEF2FF]",
                )}
                onMouseDown={(e) => e.preventDefault()}
                onMouseEnter={() => setActiveIndex(index)}
                onClick={() => selectOption(index)}
              >
                {renderHighlightedName(option.name)}
              </button>
            ))
          )}
        </div>
      ) : null}
    </div>
  );
}

function TextArea(props: {
  value: string;
  onChange: (v: string) => void;
  rows?: number;
  placeholder?: string;
}) {
  return (
    <textarea
      className="w-full resize-y rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm text-[#111827] outline-none placeholder:text-[#9CA3AF] focus:ring-2 focus:ring-[#4F46E5]"
      value={props.value}
      rows={props.rows ?? 3}
      placeholder={props.placeholder}
      onChange={(e) => props.onChange(e.target.value)}
    />
  );
}

function ProjectDrawerForm(props: {
  project: Project | null;
  mode: "create" | "edit";
  onChange: (p: Project | null) => void;
  onDelete: () => Promise<void>;
  onGotoTasks: (projectId: string) => void;
}) {
  const p = props.project;
  if (!p) return <div className="px-4 py-6 text-sm text-[#6B7280]">无数据</div>;
  return (
    <div className="grid gap-3 px-4 py-4">
      <Field label="项目名称">
        <TextInput value={p.name} onChange={(v) => props.onChange({ ...p, name: v })} />
      </Field>
      <div className="grid grid-cols-2 gap-3">
        <Field label="状态">
          <select
            className="w-full rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm"
            value={p.status}
            onChange={(e) => props.onChange({ ...p, status: e.target.value as ProjectStatus })}
          >
            {PROJECT_STATUSES.map((s) => (
              <option key={s} value={s}>
                {projectStatusLabel(s)}
              </option>
            ))}
          </select>
        </Field>
        <Field label="创建时间">
          <input
            className="w-full rounded-md border border-[#E6E8F0] bg-[#F9FAFB] px-3 py-2 text-sm text-[#6B7280]"
            value={formatDate(p.createdAt)}
            disabled
          />
        </Field>
      </div>
      <Field label="目标">
        <TextInput value={p.goal} onChange={(v) => props.onChange({ ...p, goal: v })} />
      </Field>
      <Field label="原则">
        <TextArea value={p.principles} onChange={(v) => props.onChange({ ...p, principles: v })} rows={3} />
      </Field>
      <Field label="背景 & 结果">
        <TextArea value={p.visionResult} onChange={(v) => props.onChange({ ...p, visionResult: v })} rows={3} />
      </Field>
      <Field label="详细描述">
        <TextArea value={p.description} onChange={(v) => props.onChange({ ...p, description: v })} rows={6} />
      </Field>

      {props.mode === "edit" ? (
        <div className="flex items-center justify-between pt-2">
          <button
            className="text-sm font-medium text-[#4F46E5] hover:underline"
            type="button"
            onClick={() => props.onGotoTasks(p.id)}
          >
            查看该项目任务 →
          </button>
          <button
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm text-[#DC2626] hover:bg-[#F5F6FA]"
            type="button"
            onClick={() => void props.onDelete()}
          >
            删除
          </button>
        </div>
      ) : null}
    </div>
  );
}

function InboxDrawerForm(props: {
  inbox: Inbox | null;
  mode: "create" | "edit";
  onChange: (inbox: Inbox | null) => void;
  onDelete: () => Promise<void>;
}) {
  const inbox = props.inbox;
  if (!inbox) return <div className="px-4 py-6 text-sm text-[#6B7280]">无数据</div>;
  return (
    <div className="grid gap-3 px-4 py-4">
      <Field label="名称">
        <TextInput value={inbox.name} onChange={(v) => props.onChange({ ...inbox, name: v })} />
      </Field>
      <Field label="详细描述">
        <TextArea
          value={inbox.description}
          onChange={(v) => props.onChange({ ...inbox, description: v })}
          rows={8}
        />
      </Field>

      {props.mode === "edit" ? (
        <div className="pt-2">
          <button
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm text-[#DC2626] hover:bg-[#F5F6FA]"
            type="button"
            onClick={() => void props.onDelete()}
          >
            删除
          </button>
        </div>
      ) : null}
    </div>
  );
}

function WaitingListsPage(props: {
  items: WaitingList[];
  loading: boolean;
  pageSize: number;
  offset: number;
  hasNext: boolean;
  onPageSizeChange: (v: number) => void;
  onPrevPage: () => void;
  onNextPage: () => void;
  onRefresh: () => void;
  onOpen: (id: string, waitingList?: WaitingList) => void;
}) {
  const [columnWidths, setColumnWidths] = useState<WaitingListColumnWidths>(() => loadWaitingListColumnWidths());

  useEffect(() => {
    window.localStorage.setItem(WAITING_LIST_COLUMN_WIDTHS_STORAGE_KEY, JSON.stringify(columnWidths));
  }, [columnWidths]);

  function startResize(key: WaitingListColumnKey, event: ReactPointerEvent<HTMLDivElement>) {
    event.preventDefault();

    const state: WaitingListResizeState = {
      key,
      startX: event.clientX,
      startWidth: columnWidths[key],
    };

    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";

    const onPointerMove = (moveEvent: PointerEvent) => {
      const delta = moveEvent.clientX - state.startX;
      const nextWidth = Math.max(MIN_WAITING_LIST_COLUMN_WIDTHS[key], state.startWidth + delta);
      setColumnWidths((current) => ({ ...current, [key]: nextWidth }));
    };

    const stopResize = () => {
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
      window.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", stopResize);
    };

    window.addEventListener("pointermove", onPointerMove);
    window.addEventListener("pointerup", stopResize);
  }

  function headerCell(label: string, key: WaitingListColumnKey, alignRight?: boolean) {
    const showResizeMarker = key !== "actions";
    return (
      <th className={classNames("relative px-4 py-3", alignRight && "text-right")}>
        <div className={classNames("relative", alignRight && "pr-3")}>{label}</div>
        {showResizeMarker ? (
          <div className="pointer-events-none absolute right-0 top-1/2 h-5 -translate-y-1/2 border-r border-[#D1D5DB]" />
        ) : null}
        <div
          role="separator"
          aria-orientation="vertical"
          className="absolute right-0 top-0 h-full w-2 cursor-col-resize hover:bg-[#EEF2FF]/80"
          onPointerDown={(event) => startResize(key, event)}
        />
      </th>
    );
  }

  return (
    <div className="mx-auto grid w-full max-w-7xl gap-3">
      <div className="flex items-center justify-between">
        <div className="text-lg font-semibold text-[#111827]">等待列表</div>
        <div className="flex items-center gap-2">
          <button
            className="flex items-center gap-1 rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] hover:bg-[#F5F6FA]"
            type="button"
            onClick={props.onRefresh}
          >
            <IconRefresh />
            刷新
          </button>
        </div>
      </div>

      <div className="overflow-x-auto overflow-y-hidden rounded-lg border border-[#E6E8F0] bg-white">
        <table className="w-full min-w-[1220px] table-fixed text-left text-sm">
          <colgroup>
            <col style={{ width: columnWidths.name }} />
            <col style={{ width: columnWidths.details }} />
            <col style={{ width: columnWidths.owner }} />
            <col style={{ width: columnWidths.expectedAt }} />
            <col style={{ width: columnWidths.updatedAt }} />
            <col style={{ width: columnWidths.actions }} />
          </colgroup>
          <thead className="bg-[#F9FAFB] text-xs text-[#6B7280]">
            <tr>
              {headerCell("名称", "name")}
              {headerCell("详细信息", "details")}
              {headerCell("负责人", "owner")}
              {headerCell("预期完成时间", "expectedAt")}
              {headerCell("更新时间", "updatedAt")}
              {headerCell("操作", "actions", true)}
            </tr>
          </thead>
          <tbody>
            {props.items.length === 0 && !props.loading ? (
              <tr>
                <td colSpan={6} className="px-4 py-10 text-center text-[#6B7280]">
                  暂无等待列表
                </td>
              </tr>
            ) : null}
            {props.items.map((waitingList) => (
              <tr key={waitingList.id} className="border-t border-[#E6E8F0] hover:bg-[#F9FAFB]">
                <td className="px-4 py-3">
                  <div className="font-medium text-[#111827]">{waitingList.name}</div>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">
                  <div className="line-clamp-2 break-words">{waitingList.details || "-"}</div>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">
                  <div className="truncate">{waitingList.owner || "-"}</div>
                </td>
                <td className="px-4 py-3 text-[#6B7280]">{waitingList.expectedAt ? formatDateTime(waitingList.expectedAt) : "-"}</td>
                <td className="px-4 py-3 text-[#6B7280]">{formatDate(waitingList.updatedAt)}</td>
                <td className="px-4 py-3 text-right">
                  <button
                    type="button"
                    className="text-sm font-medium text-[#4F46E5] hover:underline"
                    onClick={() => props.onOpen(waitingList.id, waitingList)}
                  >
                    详情 →
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <div className="flex items-center justify-between rounded-lg border border-[#E6E8F0] bg-white px-4 py-3 text-sm">
        <div className="flex items-center gap-2 text-[#6B7280]">
          <span>每页显示</span>
          <select
            className="rounded-md border border-[#E6E8F0] bg-white px-2 py-1 text-sm text-[#374151]"
            value={props.pageSize}
            onChange={(e) => props.onPageSizeChange(Number(e.target.value))}
            disabled={props.loading}
          >
            {PAGE_SIZE_OPTIONS.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
          <span>第 {pageNumber(props.offset, props.pageSize)} 页</span>
        </div>
        <div className="flex items-center gap-2">
          <button
            type="button"
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
            onClick={props.onPrevPage}
            disabled={props.loading || props.offset === 0}
          >
            上一页
          </button>
          <button
            type="button"
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-1 text-sm text-[#374151] disabled:cursor-not-allowed disabled:text-[#9CA3AF]"
            onClick={props.onNextPage}
            disabled={props.loading || !props.hasNext}
          >
            下一页
          </button>
        </div>
      </div>
    </div>
  );
}

function SomedayDrawerForm(props: {
  someday: Someday | null;
  mode: "create" | "edit";
  onChange: (someday: Someday | null) => void;
  onDelete: () => Promise<void>;
}) {
  const someday = props.someday;
  if (!someday) return <div className="px-4 py-6 text-sm text-[#6B7280]">无数据</div>;
  return (
    <div className="grid gap-3 px-4 py-4">
      <Field label="任务名称">
        <TextInput value={someday.name} onChange={(v) => props.onChange({ ...someday, name: v })} />
      </Field>
      <Field label="描述">
        <TextArea
          value={someday.description}
          onChange={(v) => props.onChange({ ...someday, description: v })}
          rows={8}
        />
      </Field>

      {props.mode === "edit" ? (
        <div className="pt-2">
          <button
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm text-[#DC2626] hover:bg-[#F5F6FA]"
            type="button"
            onClick={() => void props.onDelete()}
          >
            删除
          </button>
        </div>
      ) : null}
    </div>
  );
}

function WaitingListDrawerForm(props: {
  waitingList: WaitingList | null;
  mode: "create" | "edit";
  onChange: (waitingList: WaitingList | null) => void;
  onDelete: () => Promise<void>;
}) {
  const waitingList = props.waitingList;
  if (!waitingList) return <div className="px-4 py-6 text-sm text-[#6B7280]">无数据</div>;
  return (
    <div className="grid gap-3 px-4 py-4">
      <Field label="名称">
        <TextInput value={waitingList.name} onChange={(v) => props.onChange({ ...waitingList, name: v })} />
      </Field>
      <Field label="详细信息">
        <TextArea
          value={waitingList.details}
          onChange={(v) => props.onChange({ ...waitingList, details: v })}
          rows={8}
        />
      </Field>
      <Field label="负责人">
        <TextInput value={waitingList.owner} onChange={(v) => props.onChange({ ...waitingList, owner: v })} />
      </Field>
      <Field label="预期完成时间">
        <input
          className="w-full rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm"
          type="datetime-local"
          value={toDateTimeLocalValue(waitingList.expectedAt)}
          onChange={(e) => props.onChange({ ...waitingList, expectedAt: e.target.value ? fromDateTimeLocalValue(e.target.value) : undefined })}
        />
      </Field>

      {props.mode === "edit" ? (
        <div className="pt-2">
          <button
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm text-[#DC2626] hover:bg-[#F5F6FA]"
            type="button"
            onClick={() => void props.onDelete()}
          >
            删除
          </button>
        </div>
      ) : null}
    </div>
  );
}

function TaskDrawerForm(props: {
  task: Task | null;
  mode: "create" | "edit";
  projects: Project[];
  onChange: (t: Task | null) => void;
  onDelete: () => Promise<void>;
}) {
  const t = props.task;
  if (!t) return <div className="px-4 py-6 text-sm text-[#6B7280]">无数据</div>;
  return (
    <div className="grid gap-3 px-4 py-4">
      <Field label="任务名称">
        <TextInput value={t.name} onChange={(v) => props.onChange({ ...t, name: v })} />
      </Field>
      <Field label={props.mode === "create" ? "所属项目（可选）" : "所属项目"}>
        <SearchableProjectSelect
          projects={props.projects}
          value={t.projectId ?? ""}
          onChange={(value) => props.onChange({ ...t, projectId: value })}
        />
      </Field>

      <div className="grid grid-cols-2 gap-3">
        <Field label="状态">
          <select
            className="w-full rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm"
            value={t.status}
            onChange={(e) => props.onChange({ ...t, status: e.target.value as TaskStatus })}
          >
            {TASK_STATUSES.map((s) => (
              <option key={s} value={s}>
                {taskStatusLabel(s)}
              </option>
            ))}
          </select>
        </Field>
        <Field label="优先级">
          <select
            className="w-full rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm"
            value={t.priority}
            onChange={(e) => props.onChange({ ...t, priority: e.target.value as TaskPriority })}
          >
            {TASK_PRIORITIES.map((p) => (
              <option key={p} value={p}>
                {taskPriorityLabel(p)}
              </option>
            ))}
          </select>
        </Field>
      </div>

      <Field label="截止日期">
        <input
          className="w-full rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm"
          type="date"
          value={toDateValue(t.dueAt)}
          onChange={(e) => props.onChange({ ...t, dueAt: e.target.value ? fromDateValue(e.target.value) : undefined })}
        />
      </Field>

      <Field label="任务描述（可选）">
        <TextArea
          value={t.description}
          onChange={(v) => props.onChange({ ...t, description: v })}
          rows={3}
          placeholder="可选"
        />
      </Field>

      <Field label="情境分类">
        <TextInput value={t.context} onChange={(v) => props.onChange({ ...t, context: v })} />
      </Field>
      <Field label="详细信息（可选）">
        <TextArea
          value={t.details}
          onChange={(v) => props.onChange({ ...t, details: v })}
          rows={6}
          placeholder="可选"
        />
      </Field>

      {props.mode === "edit" ? (
        <div className="pt-2">
          <button
            className="rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm text-[#DC2626] hover:bg-[#F5F6FA]"
            type="button"
            onClick={() => void props.onDelete()}
          >
            删除
          </button>
        </div>
      ) : null}
    </div>
  );
}
