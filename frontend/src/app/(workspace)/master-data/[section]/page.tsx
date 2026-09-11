import type { Metadata } from "next";
import { notFound } from "next/navigation";

import {
  getMasterDataSection,
  isMasterDataSectionSlug,
  masterDataSectionSlugs,
} from "@/features/master-data/master-data-definitions";
import { MasterDataOverview } from "@/features/master-data/master-data-overview";
import { hasPermission, PERMISSIONS } from "@/lib/auth/permissions";
import { getServerSession } from "@/lib/auth/session";

interface MasterDataSectionPageProps {
  params: Promise<{ section: string }>;
}

export function generateStaticParams() {
  return masterDataSectionSlugs.map((section) => ({ section }));
}

export async function generateMetadata({
  params,
}: MasterDataSectionPageProps): Promise<Metadata> {
  const { section } = await params;
  if (!isMasterDataSectionSlug(section)) return {};
  return { title: getMasterDataSection(section).name };
}

export default async function MasterDataSectionPage({
  params,
}: MasterDataSectionPageProps) {
  const { section: sectionSlug } = await params;
  if (!isMasterDataSectionSlug(sectionSlug)) notFound();

  const session = await getServerSession();
  if (session.status !== "authenticated") return null;

  const permissions = session.user.permissions;
  const canView = hasPermission(permissions, PERMISSIONS.MASTER.READ);

  if (!canView) notFound();

  const canWrite = hasPermission(permissions, PERMISSIONS.MASTER.WRITE);

  return (
    <MasterDataOverview
      section={getMasterDataSection(sectionSlug)}
      canWrite={canWrite}
    />
  );
}
