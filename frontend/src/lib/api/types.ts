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
  total: number;
}

export interface AuthenticatedUser {
  account_id: string;
  username: string;
  email?: string;
  display_name: string;
  preferred_timezone?: string;
  last_login_at?: string;
  permissions: string[];
}

export interface LoginResult {
  token: string;
  token_type: string;
  expires_at: string;
  user: AuthenticatedUser;
}
