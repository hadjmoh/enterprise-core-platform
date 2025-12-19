import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './contexts/AuthContext';
import { AlertProvider } from './contexts/AlertContext';
import { UIProvider } from './contexts/UIContext';
import { ProtectedRoute } from './components/auth/ProtectedRoute';
import { LoginPage } from './pages/LoginPage';
import { DashboardPage } from './pages/DashboardPage';
import { UserPage } from './pages/UserPage';
import { DataInputsPage } from './pages/DataInputsPage';
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
                {/* Future routes will go here */}
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
