import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";
import type { tax } from "@/lib/client.gen";

export function useCompanies() {
  return useQuery({
    queryKey: ["companies"],
    queryFn: () => apiClient.tax.ListCompanies(),
    select: (data) => data.companies,
  });
}

export function useCompany(id: string | undefined) {
  return useQuery({
    queryKey: ["company", id],
    queryFn: () => apiClient.tax.GetCompany(id!),
    enabled: Boolean(id),
  });
}

export function useCreateCompany() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (params: tax.CreateCompanyParams) =>
      apiClient.tax.CreateCompany(params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["companies"] });
    },
  });
}
