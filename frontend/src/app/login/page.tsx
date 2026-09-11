import type { Metadata } from "next";
import { Boxes, CheckCircle2, ScanLine, ShieldCheck } from "lucide-react";
import { redirect } from "next/navigation";

import { LoginForm } from "@/features/auth/login-form";
import { getServerSession } from "@/lib/auth/session";

export const metadata: Metadata = { title: "Sign in" };

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ reason?: string }>;
}) {
  const session = await getServerSession();
  if (session.status === "authenticated") {
    redirect("/");
  }

  const { reason } = await searchParams;

  return (
    <main className="min-h-screen bg-slate-950 lg:grid lg:grid-cols-[minmax(0,1.1fr)_minmax(28rem,0.9fr)]">
      <section className="relative hidden min-h-screen overflow-hidden p-10 lg:flex lg:flex-col xl:p-14">
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_20%_15%,rgba(34,211,238,0.2),transparent_34%),radial-gradient(circle_at_85%_80%,rgba(14,116,144,0.22),transparent_38%)]" />
        <div className="surface-grid absolute inset-0 opacity-20" />

        <div className="relative flex items-center gap-3">
          <div className="grid size-11 place-items-center rounded-2xl bg-cyan-400 text-lg font-black text-slate-950 shadow-[0_0_0_6px_rgba(34,211,238,0.1)]">
            W
          </div>
          <div>
            <p className="font-bold tracking-tight text-white">NEXA WMS</p>
            <p className="text-xs text-slate-400">Operations control</p>
          </div>
        </div>

        <div className="relative my-auto max-w-2xl py-16">
          <p className="text-sm font-bold tracking-[0.2em] text-cyan-300 uppercase">
            Warehouse intelligence
          </p>
          <h1 className="mt-5 text-4xl leading-tight font-bold tracking-[-0.035em] text-white xl:text-6xl">
            One workspace for every movement.
          </h1>
          <p className="mt-6 max-w-xl text-base leading-7 text-slate-300 xl:text-lg">
            Coordinate receiving, stock control, fulfillment, and reporting with
            a traceable operational record from dock to dispatch.
          </p>
          <div className="mt-10 grid max-w-xl grid-cols-3 gap-3">
            {[
              { icon: ScanLine, label: "Traceable" },
              { icon: Boxes, label: "Inventory-led" },
              { icon: ShieldCheck, label: "Role-aware" },
            ].map(({ icon: Icon, label }) => (
              <div
                key={label}
                className="rounded-2xl border border-white/10 bg-white/[0.05] p-4 backdrop-blur"
              >
                <Icon className="size-5 text-cyan-300" />
                <p className="mt-3 text-sm font-semibold text-white">{label}</p>
              </div>
            ))}
          </div>
        </div>

        <p className="relative text-xs text-slate-500">
          Access is logged and governed by your assigned WMS permissions.
        </p>
      </section>

      <section className="flex min-h-screen items-center bg-slate-50 px-4 py-10 sm:px-8 lg:px-12 xl:px-20">
        <div className="mx-auto w-full max-w-md">
          <div className="mb-10 flex items-center gap-3 lg:hidden">
            <div className="grid size-10 place-items-center rounded-xl bg-cyan-400 font-black text-slate-950">
              W
            </div>
            <div>
              <p className="font-bold text-slate-950">NEXA WMS</p>
              <p className="text-xs text-slate-500">Operations control</p>
            </div>
          </div>

          <div className="flex size-11 items-center justify-center rounded-2xl border border-cyan-200 bg-cyan-50 text-cyan-700">
            <CheckCircle2 className="size-5" />
          </div>
          <p className="mt-6 text-sm font-bold tracking-[0.16em] text-cyan-700 uppercase">
            Secure access
          </p>
          <h2 className="mt-2 text-3xl font-bold tracking-tight text-slate-950">
            Welcome back
          </h2>
          <p className="mt-3 text-sm leading-6 text-slate-600">
            Sign in with the WMS account issued by your administrator.
          </p>

          {session.status === "unavailable" ? (
            <div className="mt-6 rounded-xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-900">
              {session.message} You can retry signing in when it is available.
            </div>
          ) : null}

          <LoginForm sessionExpired={reason === "session-expired"} />

          <p className="mt-8 text-center text-xs leading-5 text-slate-500">
            Your session is stored in a secure HttpOnly cookie and is never
            exposed to browser JavaScript.
          </p>
        </div>
      </section>
    </main>
  );
}
