import type { HTMLAttributes } from "react";

import { cn } from "@/lib/utils";

export function Panel({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <section
      className={cn(
        "rounded-2xl border border-slate-200/80 bg-white shadow-[0_1px_2px_rgba(15,23,42,0.03)]",
        className,
      )}
      {...props}
    />
  );
}
