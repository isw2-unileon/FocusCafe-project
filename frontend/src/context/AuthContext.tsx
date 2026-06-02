
import { UserStats } from '@/types/user';
import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { loginWithEmail, loginWithGoogle as googleRedirect } from '@/services/auth_service';
import { getCurrentProfile } from '@/services/user_service';

interface AuthContextType{
    isAuthenticated: boolean;
    isAdmin: boolean;
    isLoading: boolean;
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
    const [isAdmin, setIsAdmin] = useState<boolean>(false);
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [userId, setUserId] = useState<string | null>(localStorage.getItem('userId'));
    const [userStats, setUserStats] = useState<UserStats | null>(null);
    const [error, setError] = useState<string | null>(null);

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

    // Helper to process profile after getting a token
    const processProfile = useCallback(async () => {
        try {
            const profile = await getCurrentProfile();
            const isAdminUser = profile.role === 'admin';
            setIsAdmin(isAdminUser);
            setUserId(profile.id);
            localStorage.setItem('userId', profile.id);
            return isAdminUser;
        } catch (err) {
            console.error("Error fetching profile:", err);
            logout();
            return false;
        }
    }, [logout]);

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
    }, [processProfile]);

    const handleOAuthToken = useCallback(async (token: string) => {
        localStorage.setItem('token', token);
        setIsAuthenticated(true);
        await processProfile();
    }, [processProfile]);

    const loginWithGoogle = useCallback(() => {
        googleRedirect();
    }, []);

    //Effect if token expires
    useEffect(() => {
        const initAuth = async () => {
            const storedToken = localStorage.getItem('token');
            if (storedToken) {
                setIsAuthenticated(true);
                await processProfile();
            } else {
                setIsAuthenticated(false);
                setIsAdmin(false);
                setUserId(null);
            }
            setIsLoading(false);
        };
        initAuth();
    }, [processProfile]);

    return (
        <AuthContext.Provider value={{ isAuthenticated, isAdmin, isLoading, userId, userStats, setUserStats, login, loginWithGoogle, handleOAuthToken, logout, error }}>
            {children}
        </AuthContext.Provider>
    );
}

export const useAuth = () => {
    const context = useContext(AuthContext);
    if (!context) {
        throw new Error("useAuth must be used within an AuthProvider");
    }
    return context;
};