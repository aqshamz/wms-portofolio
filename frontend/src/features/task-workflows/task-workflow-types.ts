import type { PaginatedData } from "@/lib/api/types";

export type ActiveFilter = "all" | "active" | "inactive";
export type TaskWorkflowView =
  "types" | "statuses" | "transitions" | "priorities";

export interface TaskType {
  task_type_id: string;
  code: string;
  name: string;
  description?: string;
  is_active: boolean;
}

export interface TaskStatus {
  task_status_id: string;
  code: string;
  name: string;
  is_initial: boolean;
  is_final: boolean;
  is_cancelled: boolean;
  is_active: boolean;
}

export interface TaskTransition {
  task_status_transition_id: string;
  from_status_id: string;
  to_status_id: string;
  required_permission_id?: string;
  is_active: boolean;
}

export interface TaskPriority {
  task_priority_id: string;
  code: string;
  name: string;
  priority_value: number;
  is_active: boolean;
}

export type TaskTypePage = PaginatedData<TaskType>;
export type TaskStatusPage = PaginatedData<TaskStatus>;
export type TaskTransitionPage = PaginatedData<TaskTransition>;
export type TaskPriorityPage = PaginatedData<TaskPriority>;

export interface TaskListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface CreateTaskTypeRequest {
  code: string;
  name: string;
  description?: string;
}

export interface UpdateTaskTypeRequest {
  name: string;
  description?: string;
  is_active: boolean;
}

export interface CreateTaskStatusRequest {
  code: string;
  name: string;
  is_initial: boolean;
  is_final: boolean;
  is_cancelled: boolean;
}

export interface UpdateTaskStatusRequest extends Omit<
  CreateTaskStatusRequest,
  "code"
> {
  is_active: boolean;
}

export interface CreateTaskTransitionRequest {
  from_status_id: string;
  to_status_id: string;
  required_permission_id?: string;
}

export interface UpdateTaskTransitionRequest {
  required_permission_id?: string;
  is_active: boolean;
}

export interface CreateTaskPriorityRequest {
  code: string;
  name: string;
  priority_value: number;
}

export interface UpdateTaskPriorityRequest extends Omit<
  CreateTaskPriorityRequest,
  "code"
> {
  is_active: boolean;
}
