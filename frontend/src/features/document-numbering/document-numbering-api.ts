import { apiRequest } from "@/lib/api/client";
import type {
  CounterFilters,
  DocumentDailyCounterPage,
  DocumentNumberRule,
  DocumentNumberRulePage,
  GenerateDocumentIdRequest,
  GeneratedDocumentId,
  ReplaceDocumentNumberRuleRequest,
} from "@/features/document-numbering/document-numbering-types";

const documentTypesPath = "/api/v1/master/document-types";

export const documentNumberingKeys = {
  all: ["document-numbering"] as const,
  rules: (documentTypeId: string) =>
    [...documentNumberingKeys.all, documentTypeId, "rules"] as const,
  counters: (documentTypeId: string, filters: CounterFilters) =>
    [
      ...documentNumberingKeys.all,
      documentTypeId,
      "counters",
      filters,
    ] as const,
  partners: () => [...documentNumberingKeys.all, "partners"] as const,
  warehouses: () => [...documentNumberingKeys.all, "warehouses"] as const,
};

function numberRulesPath(documentTypeId: string) {
  return `${documentTypesPath}/${documentTypeId}/number-rules`;
}

export function listDocumentNumberRules(documentTypeId: string) {
  return apiRequest<DocumentNumberRulePage>(
    `${numberRulesPath(documentTypeId)}?page=1&page_size=100`,
  );
}

export function replaceDocumentNumberRule(
  documentTypeId: string,
  request: ReplaceDocumentNumberRuleRequest,
) {
  return apiRequest<DocumentNumberRule>(numberRulesPath(documentTypeId), {
    method: "POST",
    body: request,
  });
}

export function deactivateDocumentNumberRule(
  documentTypeId: string,
  ruleId: string,
) {
  return apiRequest<DocumentNumberRule>(
    `${numberRulesPath(documentTypeId)}/${ruleId}/deactivate`,
    { method: "PATCH" },
  );
}

export function generateDocumentId(
  documentTypeId: string,
  request: GenerateDocumentIdRequest,
) {
  return apiRequest<GeneratedDocumentId>(
    `${documentTypesPath}/${documentTypeId}/document-ids`,
    { method: "POST", body: request },
  );
}

export function counterListPath(
  documentTypeId: string,
  filters: CounterFilters,
) {
  const query = new URLSearchParams({
    date_from: filters.dateFrom,
    date_to: filters.dateTo,
    page: String(filters.page),
    page_size: String(filters.pageSize),
  });
  return `${documentTypesPath}/${documentTypeId}/daily-counters?${query}`;
}

export function listDocumentDailyCounters(
  documentTypeId: string,
  filters: CounterFilters,
) {
  return apiRequest<DocumentDailyCounterPage>(
    counterListPath(documentTypeId, filters),
  );
}
