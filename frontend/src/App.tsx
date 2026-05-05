import { Navigate, Route, Routes } from "react-router-dom";
import { Layout } from "@/components/Layout";
import { RequireAuth } from "@/auth/RequireAuth";
import { CompanySelectPage } from "@/features/companies/CompanySelectPage";
import { CompanyDashboardPage } from "@/features/companies/CompanyDashboardPage";
import { SimulationPage } from "@/features/tax-simulation/SimulationPage";

export function App() {
  return (
    <Routes>
      <Route
        element={
          <RequireAuth>
            <Layout />
          </RequireAuth>
        }
      >
        <Route index element={<Navigate to="/companies" replace />} />
        <Route path="/companies" element={<CompanySelectPage />} />
        <Route path="/companies/:companyId" element={<CompanyDashboardPage />} />
        <Route
          path="/companies/:companyId/simulation"
          element={<SimulationPage />}
        />
        <Route path="*" element={<Navigate to="/companies" replace />} />
      </Route>
    </Routes>
  );
}
