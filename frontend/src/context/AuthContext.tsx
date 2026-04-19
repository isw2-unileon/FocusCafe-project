
import { UserStats } from '@/types/user';
import React, { createContext, useContext, useState, useEffect } from 'react';

interface AuthContextType{
    isAuthenticated: boolean;
    userStats: UserStats | null;
    login: (token: string) =>void;
    logout: () => void;
    setUserStats : (user: UserStats | null) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: React.ReactNode })=>{
    //Initial state
    const [isAuthenticated, setIsAuthenticated] = useState<boolean>(!!localStorage.getItem('token'));
    const [userStats, setUserStats] = useState<UserStats | null>(null);

    //Login
    const login = (newToken: string) =>{
        localStorage.setItem('token', newToken);
        setIsAuthenticated(true);
        console.log(localStorage.getItem('token'))

    }

    //Logout
    const logout = () =>{
        localStorage.removeItem('token');
        setIsAuthenticated(false);
    }

    //Effect if token expires
    useEffect(() => {
        const storedToken = localStorage.getItem('token');
        if (storedToken) {
            setIsAuthenticated(true);
        }
    }, []);

    return (
        <AuthContext.Provider value={{ isAuthenticated, userStats, setUserStats, login, logout }}>
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