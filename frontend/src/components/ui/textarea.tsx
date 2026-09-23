import { forwardRef, type TextareaHTMLAttributes } from "react";
import { cn } from "@/lib/utils";

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  invalid?: boolean;
}
export const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, invalid, ...props }, ref) => (
    <textarea
      ref={ref}
      aria-invalid={invalid || undefined}
      className={cn(
        "mt-2 min-h-24 w-full resize-y rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm text-slate-950 transition outline-none placeholder:text-slate-400 focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100 disabled:bg-slate-100 disabled:text-slate-500",
        invalid && "border-rose-400",
        className,
      )}
      {...props}
    />
  ),
);
Textarea.displayName = "Textarea";
