import { forwardRef, type InputHTMLAttributes } from "react";

import { cn } from "@/lib/utils";

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  invalid?: boolean;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, invalid, ...props }, ref) => (
    <input
      ref={ref}
      aria-invalid={invalid || undefined}
      className={cn(
        "mt-2 h-11 w-full rounded-xl border border-slate-300 bg-white px-3 text-sm text-slate-950 transition outline-none placeholder:text-slate-400 read-only:bg-slate-100 read-only:text-slate-500 focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100 disabled:bg-slate-100 disabled:text-slate-500",
        invalid && "border-rose-400",
        className,
      )}
      {...props}
    />
  ),
);

Input.displayName = "Input";
