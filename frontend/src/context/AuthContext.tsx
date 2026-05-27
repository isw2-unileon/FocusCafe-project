
import { UserStats } from '@/types/user';
import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { loginWithEmail, loginWithGoogle as googleRedirect } from '@/services/auth_service';
import { getCurrentProfile } from '@/services/user_service';

interface AuthContextType{
    isAuthenticated: boolean;
    isAdmin: boolean;
    userId: string | null;
    userStats: UserStats | null;
    login: (email: string, password: string) => Promise<boolean>;
    loginWithGoogle: () => void;
    handleOAuthToken: (token: string) => Promise<void>;
    logout: () => void;
    setUserStats: React.Dispatch<React.SetStateAction<UserStats | null>>;
    error: string | null;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: React.ReactNode })=>{
    //Initial state
    const [isAuthenticated, setIsAuthenticated] = useState<boolean>(!!localStorage.getItem('token'));
    const [isAdmin, setIsAdmin] = useState<boolean>(localStorage.getItem('userRole') === 'admin');
    const [userId, setUserId] = useState<string | null>(localStorage.getItem('userId'));
    const [userStats, setUserStats] = useState<UserStats | null>(null);
    const [error, setError] = useState<string | null>(null);

    // Helper to process profile after getting a token
    const processProfile = async () => {
        const profile = await getCurrentProfile();
        const isAdminUser = profile.role === 'admin';
        setIsAdmin(isAdminUser);
        setUserId(profile.id);
        localStorage.setItem('userId', profile.id);
        if (isAdminUser) {
            localStorage.setItem('userRole', 'admin');
        } else {
            localStorage.removeItem('userRole');
        }
        return isAdminUser;
    };

    //Login
    const login = useCallback(async (email: string, password: string): Promise<boolean> => {
        try {
            setError(null);
            const token = await loginWithEmail(email, password);
            localStorage.setItem('token', token);
            setIsAuthenticated(true);

            return await processProfile();
        } catch (err) {
            setError((err as Error).message);
            throw err;
        }
    }, []);

    const handleOAuthToken = useCallback(async (token: string) => {
        localStorage.setItem('token', token);
        setIsAuthenticated(true);
        await processProfile();
    }, []);

    const loginWithGoogle = useCallback(() => {
        googleRedirect();
    }, []);

    //Logout
    const logout = useCallback(() =>{
        localStorage.removeItem('token');
        localStorage.removeItem('userRole');
        localStorage.removeItem('userId');
        setIsAuthenticated(false);
        setIsAdmin(false);
        setUserId(null);
        setUserStats(null);
        setError(null);
    }, []);

    //Effect if token expires
    useEffect(() => {
        const storedToken = localStorage.getItem('token');
        if (storedToken) {
            setIsAuthenticated(true);
            setIsAdmin(localStorage.getItem('userRole') === 'admin');
            setUserId(localStorage.getItem('userId'));
        } else {
            setIsAuthenticated(false);
            setIsAdmin(false);
            setUserId(null);
        }
    }, []);

    return (
        <AuthContext.Provider value={{ isAuthenticated, isAdmin, userId, userStats, setUserStats, login, loginWithGoogle, handleOAuthToken, logout, error }}>
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