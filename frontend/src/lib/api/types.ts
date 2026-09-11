export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data?: T;
  error?: unknown;
  request_id?: string;
}

export interface PaginatedData<T> {
  items: T[];
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
}

export type { AuthenticatedUser, LoginResult } from "@/lib/auth/schemas";
