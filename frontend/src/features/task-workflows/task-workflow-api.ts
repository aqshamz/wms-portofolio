import { apiRequest } from "@/lib/api/client";
import type {
  CreateTaskPriorityRequest,
  CreateTaskStatusRequest,
  CreateTaskTransitionRequest,
  CreateTaskTypeRequest,
  TaskListFilters,
  TaskPriority,
  TaskPriorityPage,
  TaskStatus,
  TaskStatusPage,
  TaskTransition,
  TaskTransitionPage,
  TaskType,
  TaskTypePage,
  UpdateTaskPriorityRequest,
  UpdateTaskStatusRequest,
  UpdateTaskTransitionRequest,
  UpdateTaskTypeRequest,
} from "@/features/task-workflows/task-workflow-types";

const paths = {
  types: "/api/v1/master/task-types",
  statuses: "/api/v1/master/task-statuses",
  transitions: "/api/v1/master/task-transitions",
  priorities: "/api/v1/master/task-priorities",
} as const;

export const taskWorkflowKeys = {
  all: ["task-workflows"] as const,
  list: (view: string, filters: TaskListFilters) =>
    [...taskWorkflowKeys.all, view, "list", filters] as const,
  detail: (view: string, id: string) =>
    [...taskWorkflowKeys.all, view, "detail", id] as const,
  permissions: () => [...taskWorkflowKeys.all, "permissions"] as const,
};

export function taskListPath(
  view: keyof typeof paths,
  filters: TaskListFilters,
) {
  const query = new URLSearchParams({
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  if (filters.search.trim()) query.set("search", filters.search.trim());
  if (filters.active !== "all") {
    query.set("active", String(filters.active === "active"));
  }
  return `${paths[view]}?${query}`;
}

export function listTaskTypes(filters: TaskListFilters) {
  return apiRequest<TaskTypePage>(taskListPath("types", filters));
}
export function getTaskType(id: string) {
  return apiRequest<TaskType>(`${paths.types}/${id}`);
}
export function createTaskType(request: CreateTaskTypeRequest) {
  return apiRequest<TaskType>(paths.types, { method: "POST", body: request });
}
export function updateTaskType(id: string, request: UpdateTaskTypeRequest) {
  return apiRequest<TaskType>(`${paths.types}/${id}`, {
    method: "PUT",
    body: request,
  });
}
export function deactivateTaskType(id: string) {
  return apiRequest<TaskType>(`${paths.types}/${id}/deactivate`, {
    method: "PATCH",
  });
}

export function listTaskStatuses(filters: TaskListFilters) {
  return apiRequest<TaskStatusPage>(taskListPath("statuses", filters));
}
export function getTaskStatus(id: string) {
  return apiRequest<TaskStatus>(`${paths.statuses}/${id}`);
}
export function createTaskStatus(request: CreateTaskStatusRequest) {
  return apiRequest<TaskStatus>(paths.statuses, {
    method: "POST",
    body: request,
  });
}
export function updateTaskStatus(id: string, request: UpdateTaskStatusRequest) {
  return apiRequest<TaskStatus>(`${paths.statuses}/${id}`, {
    method: "PUT",
    body: request,
  });
}
export function deactivateTaskStatus(id: string) {
  return apiRequest<TaskStatus>(`${paths.statuses}/${id}/deactivate`, {
    method: "PATCH",
  });
}

export function listTaskTransitions(filters: TaskListFilters) {
  return apiRequest<TaskTransitionPage>(taskListPath("transitions", filters));
}
export function getTaskTransition(id: string) {
  return apiRequest<TaskTransition>(`${paths.transitions}/${id}`);
}
export function createTaskTransition(request: CreateTaskTransitionRequest) {
  return apiRequest<TaskTransition>(paths.transitions, {
    method: "POST",
    body: request,
  });
}
export function updateTaskTransition(
  id: string,
  request: UpdateTaskTransitionRequest,
) {
  return apiRequest<TaskTransition>(`${paths.transitions}/${id}`, {
    method: "PUT",
    body: request,
  });
}
export function deactivateTaskTransition(id: string) {
  return apiRequest<TaskTransition>(`${paths.transitions}/${id}/deactivate`, {
    method: "PATCH",
  });
}

export function listTaskPriorities(filters: TaskListFilters) {
  return apiRequest<TaskPriorityPage>(taskListPath("priorities", filters));
}
export function getTaskPriority(id: string) {
  return apiRequest<TaskPriority>(`${paths.priorities}/${id}`);
}
export function createTaskPriority(request: CreateTaskPriorityRequest) {
  return apiRequest<TaskPriority>(paths.priorities, {
    method: "POST",
    body: request,
  });
}
export function updateTaskPriority(
  id: string,
  request: UpdateTaskPriorityRequest,
) {
  return apiRequest<TaskPriority>(`${paths.priorities}/${id}`, {
    method: "PUT",
    body: request,
  });
}
export function deactivateTaskPriority(id: string) {
  return apiRequest<TaskPriority>(`${paths.priorities}/${id}/deactivate`, {
    method: "PATCH",
  });
}
