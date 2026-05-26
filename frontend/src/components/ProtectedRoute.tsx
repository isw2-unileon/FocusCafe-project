import { Navigate, Outlet } from "react-router-dom";
import { useAuth } from "@/context/AuthContext";

export const ProtectedRoute = () => {
    const { isAuthenticated, isAdmin } = useAuth();

    if (!isAuthenticated) {
        return <Navigate to="/" replace />;
    }

    if (isAdmin) {
        return <Navigate to="/adminDashboard" replace />;
    }

    return <Outlet />;
};