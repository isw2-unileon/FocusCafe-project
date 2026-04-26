
import { UserStats } from '@/types/user';
import React, { createContext, useContext, useState, useEffect } from 'react';
import { loginWithEmail, loginWithGoogle as googleRedirect } from '@/services/auth_service';

interface AuthContextType{
    isAuthenticated: boolean;
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
    const [userStats, setUserStats] = useState<UserStats | null>(null);
    const [error, setError] = useState<string | null>(null);

    //Login
    const login = async (email: string, password: string) =>{
        try {
            setError(null);
            const token = await loginWithEmail(email, password);
            localStorage.setItem('token', token);
            setIsAuthenticated(true);
        } catch (err) {
            setError((err as Error).message);
            throw err;
        }
    }

    const loginWithGoogle = () => {
        googleRedirect();
    }

    //Logout
    const logout = () =>{
        localStorage.clear()
        setIsAuthenticated(false);
        setUserStats(null);
        setError(null);
    }

    //Effect if token expires
    useEffect(() => {
        const storedToken = localStorage.getItem('token');
        if (storedToken) {
            setIsAuthenticated(true);
        } else {
            setIsAuthenticated(false);
        }
    }, []);

    return (
        <AuthContext.Provider value={{ isAuthenticated, userStats, setUserStats, login, loginWithGoogle, logout, error }}>
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