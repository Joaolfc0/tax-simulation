import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";

export function useUploads(companyId: string | undefined) {
  return useQuery({
    queryKey: ["uploads", companyId],
    queryFn: () => apiClient.tax.ListUploads(companyId!),
    enabled: Boolean(companyId),
    select: (data) => data.uploads,
  });
}
