import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const statusVariants = cva(
  "inline-flex items-center rounded-full px-2.5 py-1 text-xs font-semibold",
  {
    variants: {
      tone: {
        neutral: "bg-slate-100 text-slate-700",
        info: "bg-cyan-50 text-cyan-800 ring-1 ring-inset ring-cyan-200",
        success:
          "bg-emerald-50 text-emerald-800 ring-1 ring-inset ring-emerald-200",
        warning: "bg-amber-50 text-amber-800 ring-1 ring-inset ring-amber-200",
        danger: "bg-rose-50 text-rose-800 ring-1 ring-inset ring-rose-200",
      },
    },
    defaultVariants: { tone: "neutral" },
  },
);

interface StatusBadgeProps extends VariantProps<typeof statusVariants> {
  children: React.ReactNode;
  className?: string;
}

export function StatusBadge({ children, className, tone }: StatusBadgeProps) {
  return (
    <span className={cn(statusVariants({ tone }), className)}>{children}</span>
  );
}
