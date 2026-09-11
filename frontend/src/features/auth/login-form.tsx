"use client";

import { useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import {
  Eye,
  EyeOff,
  LoaderCircle,
  LockKeyhole,
  UserRound,
} from "lucide-react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";

import { Button } from "@/components/ui/button";
import { loginSchema, type LoginInput } from "@/features/auth/login-schema";
import type { ApiResponse } from "@/lib/api/types";
import type { AuthenticatedUser } from "@/lib/auth/schemas";
import { cn } from "@/lib/utils";

async function submitLogin(credentials: LoginInput) {
  const response = await fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(credentials),
  });
  const payload = (await response.json()) as ApiResponse<AuthenticatedUser>;

  if (!response.ok || !payload.success || payload.data === undefined) {
    throw new Error(payload.message || "Unable to sign in.");
  }

  return payload.data;
}

export function LoginForm({ sessionExpired }: { sessionExpired: boolean }) {
  const router = useRouter();
  const [showPassword, setShowPassword] = useState(false);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginInput>({
    resolver: zodResolver(loginSchema),
    defaultValues: { identifier: "", password: "" },
  });
  const login = useMutation({
    mutationFn: submitLogin,
    onSuccess: () => {
      router.replace("/");
      router.refresh();
    },
  });

  return (
    <form
      noValidate
      className="mt-8 space-y-5"
      onSubmit={handleSubmit((values) => login.mutate(values))}
    >
      {sessionExpired ? (
        <div
          role="status"
          className="rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900"
        >
          Your session expired. Sign in again to continue.
        </div>
      ) : null}

      {login.error ? (
        <div
          role="alert"
          className="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-900"
        >
          <p className="font-semibold">{login.error.message}</p>
        </div>
      ) : null}

      <div>
        <label
          htmlFor="identifier"
          className="text-sm font-semibold text-slate-800"
        >
          Username or email
        </label>
        <div className="relative mt-2">
          <UserRound className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
          <input
            id="identifier"
            autoComplete="username"
            autoFocus
            aria-invalid={errors.identifier ? true : undefined}
            aria-describedby={
              errors.identifier ? "identifier-error" : undefined
            }
            className={cn(
              "h-12 w-full rounded-xl border bg-white pr-4 pl-10 text-sm text-slate-950 transition outline-none placeholder:text-slate-400 focus:ring-3",
              errors.identifier
                ? "border-rose-400 focus:border-rose-500 focus:ring-rose-100"
                : "border-slate-300 focus:border-cyan-500 focus:ring-cyan-100",
            )}
            placeholder="your.name or name@company.com"
            {...register("identifier")}
          />
        </div>
        {errors.identifier ? (
          <p id="identifier-error" className="mt-1.5 text-xs text-rose-700">
            {errors.identifier.message}
          </p>
        ) : null}
      </div>

      <div>
        <label
          htmlFor="password"
          className="text-sm font-semibold text-slate-800"
        >
          Password
        </label>
        <div className="relative mt-2">
          <LockKeyhole className="pointer-events-none absolute top-1/2 left-3.5 size-4 -translate-y-1/2 text-slate-400" />
          <input
            id="password"
            type={showPassword ? "text" : "password"}
            autoComplete="current-password"
            aria-invalid={errors.password ? true : undefined}
            aria-describedby={errors.password ? "password-error" : undefined}
            className={cn(
              "h-12 w-full rounded-xl border bg-white pr-12 pl-10 text-sm text-slate-950 transition outline-none placeholder:text-slate-400 focus:ring-3",
              errors.password
                ? "border-rose-400 focus:border-rose-500 focus:ring-rose-100"
                : "border-slate-300 focus:border-cyan-500 focus:ring-cyan-100",
            )}
            placeholder="Enter your password"
            {...register("password")}
          />
          <button
            type="button"
            onClick={() => setShowPassword((visible) => !visible)}
            className="absolute top-1/2 right-2 grid size-9 -translate-y-1/2 place-items-center rounded-lg text-slate-500 hover:bg-slate-100 hover:text-slate-800"
            aria-label={showPassword ? "Hide password" : "Show password"}
          >
            {showPassword ? (
              <EyeOff className="size-4" />
            ) : (
              <Eye className="size-4" />
            )}
          </button>
        </div>
        {errors.password ? (
          <p id="password-error" className="mt-1.5 text-xs text-rose-700">
            {errors.password.message}
          </p>
        ) : null}
      </div>

      <Button type="submit" className="h-12 w-full" disabled={login.isPending}>
        {login.isPending ? (
          <LoaderCircle className="size-4 animate-spin" />
        ) : (
          <LockKeyhole className="size-4" />
        )}
        {login.isPending ? "Signing in…" : "Sign in to workspace"}
      </Button>
    </form>
  );
}
