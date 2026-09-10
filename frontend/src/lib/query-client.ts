import { QueryClient } from "@tanstack/react-query";

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        refetchOnWindowFocus: false,
        retry: (failureCount, error) => {
          const status =
            typeof error === "object" && error !== null && "status" in error
              ? Number(error.status)
              : 0;

          return status >= 400 && status < 500 ? false : failureCount < 2;
        },
      },
      mutations: {
        retry: false,
      },
    },
  });
}
