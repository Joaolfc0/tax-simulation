import { useMutation, useQueryClient } from "@tanstack/react-query";
import { uploadInvoiceFile, type UploadProcessResult } from "@/lib/api/upload";

export function useUploadFile(companyId: string | undefined) {
  const queryClient = useQueryClient();
  return useMutation<UploadProcessResult, Error, File>({
    mutationFn: (file: File) => {
      if (!companyId) {
        return Promise.reject(new Error("companyId is required"));
      }
      return uploadInvoiceFile(companyId, file);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["uploads", companyId] });
      queryClient.invalidateQueries({ queryKey: ["monthReport", companyId] });
    },
  });
}
