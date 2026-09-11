"use client";

import * as SelectPrimitive from "@radix-ui/react-select";
import { Check, ChevronDown, ChevronUp } from "lucide-react";

import { cn } from "@/lib/utils";

export type SelectOption<TValue extends string> = {
  value: TValue;
  label: string;
  disabled?: boolean;
};

export type SelectProps<TValue extends string> = {
  value: TValue;
  options: readonly SelectOption<TValue>[];
  onValueChange: (value: TValue) => void;
  ariaLabel: string;
  id?: string;
  placeholder?: string;
  disabled?: boolean;
  invalid?: boolean;
  ariaDescribedBy?: string;
  className?: string;
};

export function Select<TValue extends string>({
  value,
  options,
  onValueChange,
  ariaLabel,
  id,
  placeholder,
  disabled,
  invalid,
  ariaDescribedBy,
  className,
}: SelectProps<TValue>) {
  return (
    <SelectPrimitive.Root
      value={value}
      disabled={disabled}
      onValueChange={(nextValue) => onValueChange(nextValue as TValue)}
    >
      <SelectPrimitive.Trigger
        id={id}
        aria-label={ariaLabel}
        aria-describedby={ariaDescribedBy}
        aria-invalid={invalid || undefined}
        className={cn(
          "flex h-11 w-full items-center justify-between gap-3 rounded-xl border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 shadow-sm transition-colors outline-none hover:border-slate-400 focus:border-cyan-500 focus:ring-3 focus:ring-cyan-100 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-400 data-[placeholder]:text-slate-400",
          invalid && "border-rose-400",
          className,
        )}
      >
        <SelectPrimitive.Value placeholder={placeholder} />
        <SelectPrimitive.Icon asChild>
          <ChevronDown className="size-4 shrink-0 text-slate-500" />
        </SelectPrimitive.Icon>
      </SelectPrimitive.Trigger>

      <SelectPrimitive.Portal>
        <SelectPrimitive.Content
          position="popper"
          sideOffset={5}
          collisionPadding={8}
          className="z-50 max-h-[min(18rem,var(--radix-select-content-available-height))] min-w-[var(--radix-select-trigger-width)] overflow-hidden rounded-xl border border-slate-200 bg-white p-1 shadow-xl shadow-slate-950/10"
        >
          <SelectPrimitive.ScrollUpButton className="flex h-7 items-center justify-center text-slate-500">
            <ChevronUp className="size-4" />
          </SelectPrimitive.ScrollUpButton>

          <SelectPrimitive.Viewport>
            {options.map((option) => (
              <SelectPrimitive.Item
                key={option.value}
                value={option.value}
                disabled={option.disabled}
                className="relative flex min-h-10 cursor-default items-center rounded-lg py-2 pr-9 pl-3 text-sm text-slate-700 outline-none select-none data-[disabled]:pointer-events-none data-[disabled]:opacity-40 data-[highlighted]:bg-cyan-50 data-[highlighted]:text-slate-950"
              >
                <SelectPrimitive.ItemText>
                  {option.label}
                </SelectPrimitive.ItemText>
                <SelectPrimitive.ItemIndicator className="absolute right-3 inline-flex items-center text-cyan-700">
                  <Check className="size-4" />
                </SelectPrimitive.ItemIndicator>
              </SelectPrimitive.Item>
            ))}
          </SelectPrimitive.Viewport>

          <SelectPrimitive.ScrollDownButton className="flex h-7 items-center justify-center text-slate-500">
            <ChevronDown className="size-4" />
          </SelectPrimitive.ScrollDownButton>
        </SelectPrimitive.Content>
      </SelectPrimitive.Portal>
    </SelectPrimitive.Root>
  );
}
