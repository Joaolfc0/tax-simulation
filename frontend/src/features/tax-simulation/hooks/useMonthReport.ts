import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";

export function useMonthReport(
  companyId: string | undefined,
  yearMonth: string | undefined,
) {
  return useQuery({
    queryKey: ["monthReport", companyId, yearMonth],
    queryFn: () => apiClient.tax.GetMonthReport(companyId!, yearMonth!),
    enabled: Boolean(companyId && yearMonth),
  });
}
