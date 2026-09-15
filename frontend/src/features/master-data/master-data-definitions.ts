export const masterDataSectionSlugs = [
  "core",
  "catalog",
  "operational",
] as const;

export type MasterDataSectionSlug = (typeof masterDataSectionSlugs)[number];

export interface MasterDataResource {
  name: string;
  description: string;
  includes: readonly string[];
  href?: string;
}

export interface MasterDataSection {
  slug: MasterDataSectionSlug;
  name: string;
  navigationLabel: string;
  description: string;
  apiBoundary: string;
  resources: readonly MasterDataResource[];
}

export const masterDataSections = [
  {
    slug: "core",
    name: "Core masters",
    navigationLabel: "Core masters",
    description:
      "Define the organizations, facilities, storage layout, and access scope that every warehouse operation depends on.",
    apiBoundary: "registerMasterRoutes",
    resources: [
      {
        name: "Organizations",
        description:
          "Maintain owners and the organizations represented in warehouse operations.",
        includes: ["Organization profile", "Active status"],
        href: "/master-data/core/organizations",
      },
      {
        name: "Warehouses",
        description:
          "Maintain warehouse identity and assign the owners served by each facility.",
        includes: ["Warehouse profile", "Owner assignments"],
        href: "/master-data/core/warehouses",
      },
      {
        name: "Storage layout",
        description:
          "Structure each warehouse with location types, zones, and physical locations.",
        includes: ["Location types", "Zones", "Locations"],
        href: "/master-data/core/storage-layout",
      },
      {
        name: "Access scopes",
        description:
          "Control which owners and warehouses an account is allowed to work with.",
        includes: ["Owner access", "Warehouse access"],
        href: "/master-data/core/access-scopes",
      },
    ],
  },
  {
    slug: "catalog",
    name: "Catalog",
    navigationLabel: "Catalog",
    description:
      "Maintain the commercial partners, item definitions, units, barcodes, and stock classifications used by transactions.",
    apiBoundary: "registerCatalogRoutes",
    resources: [
      {
        name: "Business partners",
        description:
          "Maintain vendors, customers, and other partners with their assigned partner types.",
        includes: ["Partner types", "Partner profiles", "Type assignments"],
        href: "/master-data/catalog/business-partners",
      },
      {
        name: "Item catalog",
        description:
          "Maintain owner-scoped categories and item handling requirements.",
        includes: ["Item categories", "Items", "Control settings"],
        href: "/master-data/catalog/items",
      },
      {
        name: "Units and packaging",
        description:
          "Define global units, item conversions, packaging, and scannable identifiers.",
        includes: ["UOMs", "Handling units", "Item UOMs", "Barcodes"],
        href: "/master-data/catalog/units-packaging",
      },
      {
        name: "Inventory classifications",
        description:
          "Define whether inventory is available, held, blocked, or otherwise controlled.",
        includes: ["Inventory statuses"],
        href: "/master-data/catalog/inventory-classifications",
      },
      {
        name: "Quality setup",
        description:
          "Standardize quality states and the outcomes recorded during inspection.",
        includes: ["Quality statuses", "Inspection results"],
        href: "/master-data/catalog/quality-setup",
      },
    ],
  },
  {
    slug: "operational",
    name: "Operational configuration",
    navigationLabel: "Operational setup",
    description:
      "Configure document lifecycles, warehouse tasks, numbering, and the strategies that guide execution.",
    apiBoundary: "registerOperationalRoutes",
    resources: [
      {
        name: "Document workflows",
        description:
          "Define document types, statuses, and the transitions allowed between them.",
        includes: ["Document types", "Statuses", "Transitions"],
        href: "/master-data/operational/document-workflows",
      },
      {
        name: "Document numbering",
        description:
          "Control identifier rules and inspect daily allocation counters.",
        includes: ["Number rules", "Document IDs", "Daily counters"],
        href: "/master-data/operational/document-numbering",
      },
      {
        name: "Task workflows",
        description:
          "Configure task types, shared statuses, transitions, and execution priorities.",
        includes: ["Task types", "Statuses", "Transitions", "Priorities"],
        href: "/master-data/operational/task-workflows",
      },
      {
        name: "Picking configuration",
        description:
          "Define sort methods and scoped picking strategies with ordered rules.",
        includes: ["Sort methods", "Strategies", "Strategy rules"],
        href: "/master-data/operational/picking-configuration",
      },
      {
        name: "Putaway configuration",
        description:
          "Define scoped putaway strategies and the rules used to select destinations.",
        includes: ["Strategies", "Strategy rules"],
        href: "/master-data/operational/putaway-configuration",
      },
      {
        name: "Modules and workflow permissions",
        description:
          "Maintain application modules and select permission metadata for workflows.",
        includes: ["Application modules", "Permission lookup"],
        href: "/master-data/operational/modules-permissions",
      },
    ],
  },
] as const satisfies readonly MasterDataSection[];

export function isMasterDataSectionSlug(
  value: string,
): value is MasterDataSectionSlug {
  return masterDataSectionSlugs.some((slug) => slug === value);
}

export function getMasterDataSection(slug: MasterDataSectionSlug) {
  return masterDataSections.find((section) => section.slug === slug)!;
}
