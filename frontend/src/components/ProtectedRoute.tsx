import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";

export const ProtectedRoute = () => {
    const { isAuthenticated, isAdmin, isLoading } = useAuth();

    if (isLoading) {
        return <div className="flex items-center justify-center h-screen">Loading...</div>;
    }

    // Read token directly from localStorage to avoid race conditions
    // where React state hasn't propagated yet after login
    const hasToken = !!localStorage.getItem('token');

    if (!hasToken) {
        return <Navigate to="/" replace />;
    }

    if (isAdmin) {
        return <Navigate to="/adminDashboard" replace />;
    }

    return <Outlet />;
};