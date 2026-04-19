
import React, { createContext, useContext, useState, useEffect } from 'react';

interface AuthContextType{
    isAuthenticated: boolean;
    login: (token: string) =>void;
    logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider = ({ children }: { children: React.ReactNode })=>{
    //Initial state
    const [isAuthenticated, setIsAuthenticated] = useState<boolean>(!!localStorage.getItem('token'));
    console.log(localStorage.getItem('token'))

    //Login
    const login = (newToken: string) =>{
        localStorage.setItem('token', newToken);
        setIsAuthenticated(true);
    }

    //Logout
    const logout = () =>{
        localStorage.removeItem('token');
        setIsAuthenticated(false);
    }

    //Efecto if token expires
    useEffect(() => {
        const storedToken = localStorage.getItem('token');
        if (storedToken) {
            setIsAuthenticated(true);
        }
    }, []);

    return (
        <AuthContext.Provider value={{ isAuthenticated, login, logout }}>
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