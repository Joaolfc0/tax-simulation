import type { ReactNode } from "react";

/**
 * A feature attached to a company. New features just need to:
 *   1. Add an entry here.
 *   2. Register their route(s) in `App.tsx`.
 *   3. The CompanyDashboardPage will pick them up automatically.
 */
export interface FeatureDescriptor {
  /** Stable id, used as React key and for analytics. */
  id: string;
  /** Short, action-oriented label shown on the dashboard tile. */
  label: string;
  /** One-line subtitle for the tile. */
  description: string;
  /** Builds the route to navigate into the feature for a given company. */
  to: (companyId: string) => string;
  /** Optional badge text (e.g. "beta", "novo"). */
  badge?: string;
  /** Optional icon node to render in the tile. */
  icon?: ReactNode;
  /** Hide the feature from the dashboard without removing it. */
  enabled?: boolean;
}

export const features: FeatureDescriptor[] = [
  {
    id: "tax-simulation",
    label: "Simulação de impostos",
    description:
      "Importe planilhas e gere relatórios mensais com destaque para fornecedores fora do Simples.",
    to: (id) => `/companies/${id}/simulation`,
    enabled: true,
  },
];

export function listEnabledFeatures(): FeatureDescriptor[] {
  return features.filter((f) => f.enabled !== false);
}
