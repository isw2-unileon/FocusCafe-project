
import { UserStats } from '@/types/user';
import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { loginWithEmail, loginWithGoogle as googleRedirect } from '@/services/auth_service';
import { getCurrentProfile } from '@/services/user_service';

interface AuthContextType{
    isAuthenticated: boolean;
    isAdmin: boolean;
    userStats: UserStats | null;
    login: (email: string, password: string) => Promise<void>;
    loginWithGoogle: () => void;
    logout: () => void;
    setUserStats : (user: UserStats | null) => void;
    error: string | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: React.ReactNode })=>{
    //Initial state
    const [isAuthenticated, setIsAuthenticated] = useState<boolean>(!!localStorage.getItem('token'));
    const [isAdmin, setIsAdmin] = useState<boolean>(localStorage.getItem('userRole') === 'admin');
    const [userStats, setUserStats] = useState<UserStats | null>(null);
    const [error, setError] = useState<string | null>(null);

    //Login
    const login = useCallback(async (email: string, password: string) =>{
        try {
            setError(null);
            const token = await loginWithEmail(email, password);
            localStorage.setItem('token', token);
            setIsAuthenticated(true);

            // Fetch real user profile to get the role from the database
            const profile = await getCurrentProfile();
            setIsAdmin(profile.role === 'admin');
            if (profile.role === 'admin') {
                localStorage.setItem('userRole', 'admin');
            } else {
                localStorage.removeItem('userRole');
            }
        } catch (err) {
            setError((err as Error).message);
            throw err;
        }
    }, []);

    const loginWithGoogle = useCallback(() => {
        googleRedirect();
    }, []);

    //Logout
    const logout = useCallback(() =>{
        localStorage.removeItem('token');
        localStorage.removeItem('userRole');
        setIsAuthenticated(false);
        setIsAdmin(false);
        setUserStats(null);
        setError(null);
    }, []);

    //Effect if token expires
    useEffect(() => {
        const storedToken = localStorage.getItem('token');
        if (storedToken) {
            setIsAuthenticated(true);
            setIsAdmin(localStorage.getItem('userRole') === 'admin');
        } else {
            setIsAuthenticated(false);
            setIsAdmin(false);
        }
    }, []);

    return (
        <AuthContext.Provider value={{ isAuthenticated, isAdmin, userStats, setUserStats, login, loginWithGoogle, logout, error }}>
            {children}
        </AuthContext.Provider>
    );
}

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error("useAuth debe usarse dentro de un AuthProvider");
    }
    return context;
};