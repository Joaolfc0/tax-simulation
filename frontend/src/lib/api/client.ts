import Client, { Local, type BaseURL } from "@/lib/client.gen";

const target: BaseURL = (import.meta.env.VITE_API_URL as BaseURL) ?? Local;

export const apiClient = new Client(target, {
  // Future: requestInit: { headers: { Authorization: `Bearer ${token}` } }
});

export { target as apiBaseURL };

export interface ApiError extends Error {
  status?: number;
  code?: string;
}

export function toApiError(err: unknown): ApiError {
  if (err instanceof Error) {
    return err as ApiError;
  }
  return new Error(String(err));
}
