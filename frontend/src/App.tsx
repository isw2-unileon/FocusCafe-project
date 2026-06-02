import { BrowserRouter, Routes, Route, Navigate, Outlet } from "react-router-dom";
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { Toaster } from 'react-hot-toast';
import { useAuth } from "./context/AuthContext";

import Home from '@/pages/Home';
import Login from '@/pages/Login';
import Register from '@/pages/Register';
import AuthCallback from './pages/AuthCallback';
import Dashboard from '@/pages/Dashboard';
import EditProfile from '@/pages/EditProfile';
import StudySession from '@/pages/StudySession';
import { AuthProvider } from "./context/AuthContext";
import { WebSocketProvider } from "./context/WebSocketContext";
import AdminDashboard from "./pages/AdminDashboard";

const AdminRoute = () => {
  const { isAuthenticated, isAdmin, isLoading } = useAuth();
  if (isLoading) return <div className="flex items-center justify-center h-screen">Loading...</div>;
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  if (!isAdmin) return <Navigate to="/home" replace />;
  return <Outlet />;
};

export default function App() {
  return (
    <AuthProvider>
      <WebSocketProvider>
        <BrowserRouter>
          <Toaster position="top-center" />
          <Routes>
            {/*Public routes*/}
            <Route path="/" element={<Navigate to="/login" replace />} /> 
            <Route path="/login" element={<Login />} />
            <Route path="/register" element={<Register />} />
            <Route path="/auth/callback" element={<AuthCallback />} />
            
            {/*Private routes*/}
                <Route element={<ProtectedRoute />}>
                  <Route path="/home" element={<Home />} />
                  <Route path="/study" element={<StudySession />} />
                  <Route path="/dashboard" element={<Dashboard />} />
                  <Route path="/edit-profile" element={<EditProfile />} />
                </Route>

                {/* Admin routes */}
                <Route element={<AdminRoute />}>
                  <Route path="/adminDashboard" element={<AdminDashboard />} />
                </Route>
            <Route path="*" element={<Navigate to="/" />} />
          </Routes>
        </BrowserRouter>
      </WebSocketProvider>
    </AuthProvider>
  );
}