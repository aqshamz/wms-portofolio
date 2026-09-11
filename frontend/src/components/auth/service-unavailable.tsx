"use client";

import { RefreshCw, ServerOff } from "lucide-react";
import { useRouter } from "next/navigation";

import { Button } from "@/components/ui/button";

export function ServiceUnavailable({ message }: { message: string }) {
  const router = useRouter();

  return (
    <main className="surface-grid grid min-h-screen place-items-center bg-slate-50 px-4 py-10">
      <section className="w-full max-w-lg rounded-3xl border border-slate-200 bg-white p-6 shadow-xl shadow-slate-950/5 sm:p-8">
        <div className="grid size-12 place-items-center rounded-2xl bg-amber-100 text-amber-700">
          <ServerOff className="size-6" />
        </div>
        <p className="mt-6 text-xs font-bold tracking-[0.18em] text-amber-700 uppercase">
          Service interruption
        </p>
        <h1 className="mt-2 text-2xl font-bold tracking-tight text-slate-950">
          The workspace cannot be loaded yet
        </h1>
        <p className="mt-3 text-sm leading-6 text-slate-600">{message}</p>
        <div className="mt-5 rounded-xl border border-cyan-200 bg-cyan-50 px-4 py-3 text-sm text-cyan-900">
          Your browser session has been kept. You will not be signed out because
          of a temporary API or database outage.
        </div>
        <Button
          className="mt-6 w-full sm:w-auto"
          onClick={() => router.refresh()}
        >
          <RefreshCw className="size-4" />
          Try again
        </Button>
      </section>
    </main>
  );
}
