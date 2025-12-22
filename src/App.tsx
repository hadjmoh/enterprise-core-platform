import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { AlertProvider } from './contexts/AlertContext';
import { UIProvider } from './contexts/UIContext';
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { LoginPage } from './pages/LoginPage';
import { DashboardPage } from './pages/DashboardPage';
import { UserPage } from './pages/UserPage';
import { DataInputsPage } from './pages/DataInputsPage';
import { SecurityDashboardPage } from './pages/SecurityDashboardPage';
import { IncidentReviewPage } from './pages/IncidentReviewPage';
import { EntityProfilePage } from './pages/EntityProfilePage';
import { ThreatIntelPage } from './pages/ThreatIntelPage';
import { SecurityRulesPage } from './pages/SecurityRulesPage';
import { CloudAssetsPage } from './pages/CloudAssetsPage';
import { CompliancePage } from './pages/CompliancePage';
import { ClusterManagerPage } from './pages/ClusterManagerPage';
import { TrustGraphPage } from './pages/TrustGraphPage';
import { SimulationLabPage } from './pages/SimulationLabPage';
import { ControlCenterPage } from './pages/ControlCenterPage';
import AppDashboardPage from '@/pages/AppDashboardPage';
import AppStorePage from '@/pages/AppStorePage';
import { DashboardLayout } from './components/layout/DashboardLayout';

function App() {
  return (
    <AuthProvider>
      <UIProvider>
        <AlertProvider>
          <BrowserRouter>
            <Routes>
              <Route path="/login" element={<LoginPage />} />

              <Route element={<ProtectedRoute />}>
                <Route
                  path="/"
                  element={
                    <DashboardLayout>
                      <DashboardPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/users"
                  element={
                    <DashboardLayout>
                      <UserPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/data-inputs"
                  element={
                    <DashboardLayout>
                      <DataInputsPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/security/dashboard"
                  element={
                    <DashboardLayout>
                      <SecurityDashboardPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/security/incidents"
                  element={
                    <DashboardLayout>
                      <IncidentReviewPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/entity/:entityId"
                  element={
                    <DashboardLayout>
                      <EntityProfilePage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/threat-intel"
                  element={
                    <DashboardLayout>
                      <ThreatIntelPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/security/rules"
                  element={
                    <DashboardLayout>
                      <SecurityRulesPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/cloud/assets"
                  element={
                    <DashboardLayout>
                      <CloudAssetsPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/compliance"
                  element={
                    <DashboardLayout>
                      <CompliancePage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/cluster"
                  element={
                    <DashboardLayout>
                      <ClusterManagerPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/trust-graph"
                  element={
                    <DashboardLayout>
                      <TrustGraphPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/simulation-lab"
                  element={
                    <DashboardLayout>
                      <SimulationLabPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/control-center"
                  element={
                    <DashboardLayout>
                      <ControlCenterPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/apps/:appName/dashboards/:dashboardId"
                  element={
                    <DashboardLayout>
                      <AppDashboardPage />
                    </DashboardLayout>
                  }
                />
                <Route
                  path="/apps/store"
                  element={
                    <DashboardLayout>
                      <AppStorePage />
                    </DashboardLayout>
                  }
                />
              </Route>

              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </BrowserRouter>
        </AlertProvider>
      </UIProvider>
    </AuthProvider>
  );
}

export default App;
