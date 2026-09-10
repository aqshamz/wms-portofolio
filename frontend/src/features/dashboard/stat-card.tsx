import type { LucideIcon } from "lucide-react";
import { ArrowDownRight, ArrowUpRight } from "lucide-react";

import { Panel } from "@/components/ui/panel";
import { cn } from "@/lib/utils";

interface StatCardProps {
  label: string;
  value: string;
  detail: string;
  trend?: "up" | "down";
  icon: LucideIcon;
  tone: "cyan" | "indigo" | "amber" | "emerald";
}

const tones: Record<StatCardProps["tone"], string> = {
  cyan: "bg-cyan-50 text-cyan-700",
  indigo: "bg-indigo-50 text-indigo-700",
  amber: "bg-amber-50 text-amber-700",
  emerald: "bg-emerald-50 text-emerald-700",
};

export function StatCard({
  label,
  value,
  detail,
  trend,
  icon: Icon,
  tone,
}: StatCardProps) {
  const TrendIcon = trend === "down" ? ArrowDownRight : ArrowUpRight;

  return (
    <Panel className="p-4 sm:p-5">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="text-sm font-medium text-slate-500">{label}</p>
          <p className="mt-2 text-2xl font-bold tracking-tight text-slate-950 sm:text-[1.75rem]">
            {value}
          </p>
        </div>
        <div
          className={cn(
            "grid size-10 place-items-center rounded-xl",
            tones[tone],
          )}
        >
          <Icon className="size-5" />
        </div>
      </div>
      <div className="mt-4 flex items-center gap-1.5 text-xs text-slate-500">
        {trend ? (
          <TrendIcon
            className={cn(
              "size-4",
              trend === "up" ? "text-emerald-600" : "text-rose-600",
            )}
          />
        ) : null}
        <span>{detail}</span>
      </div>
    </Panel>
  );
}
