import type { LucideIcon } from "lucide-react";
import {
  ArrowLeft,
  ArrowRight,
  Boxes,
  Building2,
  Check,
  PencilLine,
  Workflow,
} from "lucide-react";
import Link from "next/link";

import { Panel } from "@/components/ui/panel";
import {
  masterDataSections,
  type MasterDataSection,
  type MasterDataSectionSlug,
} from "@/features/master-data/master-data-definitions";

const sectionIcons: Record<MasterDataSectionSlug, LucideIcon> = {
  core: Building2,
  catalog: Boxes,
  operational: Workflow,
};

function ResourceGrid({ section }: { section: MasterDataSection }) {
  return (
    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
      {section.resources.map((resource) => {
        const card = (
          <Panel className="flex min-h-52 flex-col p-5 transition group-hover:border-cyan-300 group-hover:shadow-md">
            <h2 className="text-base font-bold text-slate-950">
              {resource.name}
            </h2>
            <p className="mt-2 text-sm leading-6 text-slate-600">
              {resource.description}
            </p>
            <ul
              className="mt-5 flex flex-wrap gap-2"
              aria-label="Included data"
            >
              {resource.includes.map((item) => (
                <li
                  key={item}
                  className="inline-flex items-center gap-1.5 rounded-full bg-slate-100 px-2.5 py-1 text-xs font-semibold text-slate-600"
                >
                  <Check className="size-3 text-cyan-700" />
                  {item}
                </li>
              ))}
            </ul>
            {resource.href ? (
              <span className="mt-5 inline-flex items-center gap-2 text-sm font-bold text-cyan-800">
                Manage organizations
                <ArrowRight className="size-4 transition-transform group-hover:translate-x-1" />
              </span>
            ) : null}
          </Panel>
        );

        return resource.href ? (
          <Link
            key={resource.name}
            href={resource.href}
            className="group rounded-2xl focus-visible:ring-2 focus-visible:ring-cyan-500 focus-visible:ring-offset-2 focus-visible:outline-none"
          >
            {card}
          </Link>
        ) : (
          <div key={resource.name}>{card}</div>
        );
      })}
    </div>
  );
}

export function MasterDataOverview({
  section,
  canWrite,
}: {
  section?: MasterDataSection;
  canWrite: boolean;
}) {
  if (section) {
    const Icon = sectionIcons[section.slug];

    return (
      <div className="space-y-7">
        <header>
          <Link
            href="/master-data"
            className="inline-flex min-h-10 items-center gap-2 rounded-lg text-sm font-semibold text-slate-600 hover:text-slate-950 focus-visible:ring-2 focus-visible:ring-cyan-500 focus-visible:outline-none"
          >
            <ArrowLeft className="size-4" />
            All master data
          </Link>
          <div className="mt-4 flex items-start gap-4">
            <div className="grid size-12 shrink-0 place-items-center rounded-2xl bg-slate-950 text-cyan-300">
              <Icon className="size-6" />
            </div>
            <div>
              <div className="flex flex-wrap items-center gap-2">
                <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
                  {section.name}
                </h1>
                <span className="inline-flex items-center gap-1.5 rounded-full bg-cyan-50 px-2.5 py-1 text-xs font-bold text-cyan-800 ring-1 ring-cyan-200">
                  {canWrite ? <PencilLine className="size-3" /> : null}
                  {canWrite ? "Edit access" : "View only"}
                </span>
              </div>
              <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-600 sm:text-base">
                {section.description}
              </p>
            </div>
          </div>
        </header>

        <ResourceGrid section={section} />
      </div>
    );
  }

  return (
    <div className="space-y-7">
      <header>
        <p className="text-sm font-semibold text-cyan-700">Administration</p>
        <div className="mt-1 flex flex-wrap items-center gap-3">
          <h1 className="text-2xl font-bold tracking-tight text-slate-950 sm:text-3xl">
            Master data
          </h1>
          <span className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-600">
            {canWrite ? "Edit access" : "View only"}
          </span>
        </div>
        <p className="mt-2 max-w-3xl text-sm leading-6 text-slate-600 sm:text-base">
          Choose the type of master data you need to review or maintain. The
          groups follow the same boundaries used by the WMS API.
        </p>
      </header>

      <div className="grid gap-4 lg:grid-cols-3">
        {masterDataSections.map((item) => {
          const Icon = sectionIcons[item.slug];

          return (
            <Link
              key={item.slug}
              href={`/master-data/${item.slug}`}
              className="group rounded-2xl focus-visible:ring-2 focus-visible:ring-cyan-500 focus-visible:ring-offset-2 focus-visible:outline-none"
            >
              <Panel className="flex h-full min-h-72 flex-col p-5 transition group-hover:-translate-y-0.5 group-hover:border-cyan-300 group-hover:shadow-lg sm:p-6">
                <div className="grid size-11 place-items-center rounded-xl bg-slate-950 text-cyan-300">
                  <Icon className="size-5" />
                </div>
                <p className="mt-6 text-xs font-bold tracking-[0.14em] text-slate-500 uppercase">
                  {item.resources.length} management areas
                </p>
                <h2 className="mt-2 text-xl font-bold text-slate-950">
                  {item.name}
                </h2>
                <p className="mt-2 flex-1 text-sm leading-6 text-slate-600">
                  {item.description}
                </p>
                <span className="mt-5 inline-flex min-h-10 items-center gap-2 text-sm font-bold text-cyan-800">
                  Browse master data
                  <ArrowRight className="size-4 transition-transform group-hover:translate-x-1" />
                </span>
              </Panel>
            </Link>
          );
        })}
      </div>
    </div>
  );
}
