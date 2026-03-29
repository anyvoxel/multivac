import type { ReactNode } from "react";
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
  listTasks,
  setTaskStatus,
  updateTask,
} from "./lib/tasks";
import type { Task, TaskPriority, TaskStatus } from "./lib/tasks";
import {
  formatDate,
  fromDateValue,
  toDateValue,
} from "./lib/time";

type Route = "projects" | "tasks";

const PROJECT_STATUSES: ProjectStatus[] = [
  "Draft",
  "Active",
  "Completed",
  "Archived",
];
const TASK_STATUSES: TaskStatus[] = ["Todo", "InProgress", "Done", "Canceled"];
const TASK_PRIORITIES: TaskPriority[] = ["P0", "High", "Medium", "Low"];

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

export default function App() {
  const [route, setRoute] = useState<Route>("projects");

  const [search, setSearch] = useState<string>("");

  // Projects list page state
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectStatusFilter, setProjectStatusFilter] = useState<
    ProjectStatus | ""
  >("");

  // Drawer state
  const [drawer, setDrawer] = useState<
    | { type: "none" }
    | { type: "project"; mode: "create" }
    | { type: "project"; mode: "edit"; id: string }
    | { type: "task"; mode: "create" }
    | { type: "task"; mode: "edit"; id: string }
  >({ type: "none" });
  const [drawerLoading, setDrawerLoading] = useState(false);
  const [drawerSaving, setDrawerSaving] = useState(false);
  const [drawerProject, setDrawerProject] = useState<Project | null>(null);
  const [drawerTask, setDrawerTask] = useState<Task | null>(null);

  // Tasks page filters
  const [taskProjectId, setTaskProjectId] = useState<string>("");
  const [taskStatusFilter, setTaskStatusFilter] = useState<TaskStatus | "">("");

  async function refreshProjects() {
    setLoading(true);
    setError("");
    try {
      const list = await listProjects(
        projectStatusFilter ? { status: projectStatusFilter } : undefined,
      );
      setProjects(list);
      if (taskProjectId) {
        const still = list.find((p) => p.id === taskProjectId);
        if (!still) setTaskProjectId("");
      }
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void refreshProjects();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectStatusFilter]);

  const filteredProjects = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return projects;
    return projects.filter((p) => {
      return (
        p.name.toLowerCase().includes(q) ||
        p.goal.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q)
      );
    });
  }, [projects, search]);

  const filteredTasksQuery = useMemo(() => {
    const q = search.trim().toLowerCase();
    return q;
  }, [search]);

  const pageLabel = route === "projects" ? "项目管理" : "任务管理";

  async function openProjectDrawer(mode: "create" | "edit", id?: string) {
    setDrawerProject(null);
    setDrawerTask(null);
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
    setDrawerLoading(true);
    setError("");
    try {
      if (mode === "create") {
        setDrawer({ type: "task", mode: "create" });
        const defaultProject = taskProjectId || projects[0]?.id || "";
        setDrawerTask({
          id: "",
          projectId: defaultProject,
          name: "",
          description: "",
          context: "默认",
          details: "",
          status: "InProgress",
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
      // Fallback: fetch task by listing; we keep minimal by reusing listTasks.
      const list = await listTasks();
      const found = list.find((x) => x.id === id) ?? null;
      setDrawerTask(found);
    } catch (e) {
      setError(String(e));
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
    if (route === "projects") {
      await refreshProjects();
      return;
    }
    await refreshProjects();
    // tasks list refresh is handled inside TasksPage, but we keep this as no-op.
  }

  return (
    <div className="min-h-full bg-[#F5F6FA]">
      <div className="flex min-h-screen">
        <aside className="w-56 border-r border-[#E6E8F0] bg-white">
          <div className="flex items-center gap-2 px-4 py-4">
            <IconLogo />
            <div className="text-sm font-semibold text-[#111827]">MULTIVAC</div>
          </div>
          <div className="px-4 pb-2 text-xs font-medium text-[#6B7280]">菜单</div>
          <nav className="px-2">
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
                    className="w-64 rounded-md border border-[#E6E8F0] bg-white py-1.5 pl-8 pr-3 text-sm outline-none focus:ring-2 focus:ring-[#4F46E5]"
                    placeholder="搜索..."
                    value={search}
                    onChange={(e) => setSearch(e.target.value)}
                  />
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
                    else void openTaskDrawer("create");
                  }}
                >
                  + 新建
                </button>
              </div>
            </div>
          </header>

          <main className="min-w-0 flex-1 px-6 py-6">
            {error ? (
              <div className="mb-4 rounded-md border border-[#FCA5A5] bg-[#FEF2F2] px-3 py-2 text-sm text-[#B91C1C]">
                {error}
              </div>
            ) : null}

            {route === "projects" ? (
              <ProjectsPage
                items={filteredProjects}
                loading={loading}
                statusFilter={projectStatusFilter}
                onStatusFilter={setProjectStatusFilter}
                onRefresh={() => void refreshProjects()}
                onOpen={(id) => void openProjectDrawer("edit", id)}
              />
            ) : (
              <TasksPageNew
                projects={projects}
                projectId={taskProjectId}
                status={taskStatusFilter}
                search={filteredTasksQuery}
                onProjectId={setTaskProjectId}
                onStatus={setTaskStatusFilter}
                onOpen={(id, t) => void openTaskDrawer("edit", id, t)}
              />
            )}
          </main>
        </div>

        {drawer.type !== "none" ? (
          <Drawer
            title={drawer.type === "project" ? "项目详情" : "任务详情"}
            onClose={() => {
              setDrawer({ type: "none" });
              setDrawerProject(null);
              setDrawerTask(null);
            }}
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
                  setDrawer({ type: "none" });
                  setDrawerProject(null);
                }}
              />
            ) : (
              <TaskDrawerForm
                task={drawerTask}
                mode={drawer.mode}
                projects={projects}
                onChange={setDrawerTask}
                onDelete={async () => {
                  if (!drawerTask) return;
                  if (!confirm(`确定删除 Task: ${drawerTask.name} ?`)) return;
                  await deleteTask(drawerTask.id);
                  setDrawer({ type: "none" });
                  setDrawerTask(null);
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
        const p = await getProject(drawerProject.id);
        setDrawerProject(p);
        return;
      }

      if (drawer.type === "task") {
        if (!drawerTask) return;
        const dueISO = fromDateValue(drawerTask.dueAt ? toDateValue(drawerTask.dueAt) : "");
        if (drawer.mode === "create") {
          const t = await createTask({
            projectId: drawerTask.projectId,
            name: drawerTask.name,
            description: drawerTask.description,
            context: drawerTask.context,
            details: drawerTask.details,
            priority: drawerTask.priority,
            status: drawerTask.status,
            dueAt: dueISO,
          });
          setDrawer({ type: "task", mode: "edit", id: t.id });
          setDrawerTask(t);
          return;
        }
        const t = await updateTask(drawerTask.id, {
          name: drawerTask.name,
          description: drawerTask.description,
          context: drawerTask.context,
          details: drawerTask.details,
          priority: drawerTask.priority,
          dueAt: dueISO ?? "",
        });
        if (t.status !== drawerTask.status) {
          const out = await setTaskStatus(drawerTask.id, drawerTask.status);
          setDrawerTask(out);
        } else {
          setDrawerTask(t);
        }
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
        "flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm",
        props.active
          ? "bg-[#EEF2FF] text-[#4F46E5]"
          : "text-[#374151] hover:bg-[#F5F6FA]",
      )}
      onClick={props.onClick}
    >
      <span className={props.active ? "text-[#4F46E5]" : "text-[#6B7280]"}>
        {props.icon}
      </span>
      <span className="font-medium">{props.label}</span>
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

function ProjectsPage(props: {
  items: Project[];
  loading: boolean;
  statusFilter: ProjectStatus | "";
  onStatusFilter: (v: ProjectStatus | "") => void;
  onRefresh: () => void;
  onOpen: (id: string) => void;
}) {
  return (
    <div className="mx-auto grid max-w-4xl gap-3">
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

      <div className="overflow-hidden rounded-lg border border-[#E6E8F0] bg-white">
        <table className="w-full text-left text-sm">
          <thead className="bg-[#F9FAFB] text-xs text-[#6B7280]">
            <tr>
              <th className="px-4 py-3">项目名称</th>
              <th className="px-4 py-3">状态</th>
              <th className="px-4 py-3">目标</th>
              <th className="px-4 py-3">更新时间</th>
              <th className="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            {props.items.length === 0 && !props.loading ? (
              <tr>
                <td colSpan={5} className="px-4 py-10 text-center text-[#6B7280]">
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
                <td className="px-4 py-3 text-[#374151]">{p.goal}</td>
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
    </div>
  );
}

function TasksPageNew(props: {
  projects: Project[];
  projectId: string;
  status: TaskStatus | "";
  search: string;
  onProjectId: (v: string) => void;
  onStatus: (v: TaskStatus | "") => void;
  onOpen: (id: string, t?: Task) => void;
}) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string>("");
  const [items, setItems] = useState<Task[]>([]);

  async function refresh() {
    setLoading(true);
    setError("");
    try {
      const list = await listTasks(
        props.projectId || props.status
          ? { projectId: props.projectId || undefined, status: props.status || undefined }
          : undefined,
      );
      setItems(list);
    } catch (e) {
      setError(String(e));
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void refresh();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.projectId, props.status]);

  const shown = useMemo(() => {
    if (!props.search) return items;
    return items.filter((t) => {
      const q = props.search;
      return t.name.toLowerCase().includes(q) || t.description.toLowerCase().includes(q) || t.details.toLowerCase().includes(q);
    });
  }, [items, props.search]);

  return (
    <div className="mx-auto grid max-w-4xl gap-3">
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

      <div className="overflow-hidden rounded-lg border border-[#E6E8F0] bg-white">
        <table className="w-full text-left text-sm">
          <thead className="bg-[#F9FAFB] text-xs text-[#6B7280]">
            <tr>
              <th className="px-4 py-3">任务名称</th>
              <th className="px-4 py-3">优先级</th>
              <th className="px-4 py-3">状态</th>
              <th className="px-4 py-3">截止日期</th>
              <th className="px-4 py-3 text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            {shown.length === 0 && !loading ? (
              <tr>
                <td colSpan={5} className="px-4 py-10 text-center text-[#6B7280]">
                  暂无 Task
                </td>
              </tr>
            ) : null}
            {shown.map((t) => (
              <tr key={t.id} className="border-t border-[#E6E8F0] hover:bg-[#F9FAFB]">
                <td className="px-4 py-3">
                  <div className="font-medium text-[#111827]">{t.name}</div>
                </td>
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

function TextArea(props: { value: string; onChange: (v: string) => void; rows?: number }) {
  return (
    <textarea
      className="w-full resize-y rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm text-[#111827] outline-none focus:ring-2 focus:ring-[#4F46E5]"
      value={props.value}
      rows={props.rows ?? 3}
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
      <Field label="所属项目">
        <select
          className="w-full rounded-md border border-[#E6E8F0] bg-white px-3 py-2 text-sm"
          value={t.projectId}
          onChange={(e) => props.onChange({ ...t, projectId: e.target.value })}
        >
          <option value="">请选择</option>
          {props.projects.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </select>
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

      <Field label="任务描述">
        <TextArea value={t.description} onChange={(v) => props.onChange({ ...t, description: v })} rows={3} />
      </Field>

      <Field label="情境分类">
        <TextInput value={t.context} onChange={(v) => props.onChange({ ...t, context: v })} />
      </Field>
      <Field label="详细信息">
        <TextArea value={t.details} onChange={(v) => props.onChange({ ...t, details: v })} rows={6} />
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
