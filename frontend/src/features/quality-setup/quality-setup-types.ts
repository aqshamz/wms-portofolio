import type { PaginatedData } from "@/lib/api/types";

export type ActiveFilter = "all" | "active" | "inactive";
export type QualitySetupView = "statuses" | "results";

export interface QualityStatus {
  quality_status_id: string;
  code: string;
  name: string;
  description?: string;
  is_active: boolean;
}

export interface InspectionResult {
  inspection_result_id: string;
  code: string;
  name: string;
  description?: string;
  is_accepted: boolean;
  is_active: boolean;
}

export type QualityStatusPage = PaginatedData<QualityStatus>;
export type InspectionResultPage = PaginatedData<InspectionResult>;

export interface QualityListFilters {
  search: string;
  active: ActiveFilter;
  page: number;
  pageSize: number;
}

export interface CreateQualityStatusRequest {
  code: string;
  name: string;
  description?: string;
}

export interface UpdateQualityStatusRequest {
  name: string;
  description?: string;
  is_active: boolean;
}

export interface CreateInspectionResultRequest {
  code: string;
  name: string;
  description?: string;
  is_accepted: boolean;
}

export interface UpdateInspectionResultRequest {
  name: string;
  description?: string;
  is_accepted: boolean;
  is_active: boolean;
}
