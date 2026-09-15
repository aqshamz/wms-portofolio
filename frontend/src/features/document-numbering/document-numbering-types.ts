import type { PaginatedData } from "@/lib/api/types";

export type NumberingView = "rules" | "allocate" | "counters";

export interface DocumentNumberRule {
  document_number_rule_id: string;
  document_type_id: string;
  prefix: string;
  separator: string;
  sequence_length: number;
  include_partner_code: boolean;
  include_warehouse_code: boolean;
  is_active: boolean;
  effective_from: string;
  effective_until?: string;
  created_at: string;
  created_by?: string;
}

export interface DocumentDailyCounter {
  document_type_id: string;
  business_date: string;
  last_number: string;
  updated_at: string;
}

export interface GeneratedDocumentId {
  document_id: string;
  document_type_id: string;
  document_number_rule_id: string;
  business_date: string;
  sequence_number: string;
}

export type DocumentNumberRulePage = PaginatedData<DocumentNumberRule>;
export type DocumentDailyCounterPage = PaginatedData<DocumentDailyCounter>;

export interface ReplaceDocumentNumberRuleRequest {
  prefix: string;
  separator: string;
  sequence_length: number;
  include_partner_code: boolean;
  include_warehouse_code: boolean;
  effective_from: string;
}

export interface GenerateDocumentIdRequest {
  business_date: string;
  partner_id?: string;
  warehouse_id?: string;
}

export interface CounterFilters {
  dateFrom: string;
  dateTo: string;
  page: number;
  pageSize: number;
}
