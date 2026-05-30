import React, { createContext, useContext, useEffect, useRef, useCallback } from 'react';
import { useAuth } from './AuthContext';


interface WebSocketContextType {
    sendMessage: (msg: unknown) => void;
    subscribe: (type: string, callback: (payload: Record<string, unknown>) => void) => () => void;
}

const WebSocketContext = createContext<WebSocketContextType | null>(null);

export const useWebSocket = () => {
    const context = useContext(WebSocketContext);
    if (!context) {
        throw new Error('useWebSocket must be used within a WebSocketProvider');
    }
    return context;
};

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const { isAuthenticated, userStats } = useAuth();
    const socketRef = useRef<WebSocket | null>(null);
    const subscribersRef = useRef<Record<string, ((payload: Record<string, unknown>) => void)[]>>({});
    
    // Captured group ID to trigger reconnection if it changes
    const groupID = userStats?.group?.id;

    const connect = useCallback(() => {
        const token = localStorage.getItem('token');
        if (!isAuthenticated || !token) return;

        const apiUrl = import.meta.env.VITE_API_URL;
        if(!apiUrl){
            console.error("No VITE_API_URL");
            return;
        }
        const wsUrl = `${apiUrl}/ws`;

        const socket = new WebSocket(wsUrl);

        socket.onopen = () => {
            console.log('WebSocket Connected');
            // Identify immediately after connection
            socket.send(JSON.stringify({ type: 'AUTH', payload: token }));
        };

        socket.onmessage = (event) => {            
            const message = JSON.parse(event.data);
            const type = message.type;
            const payload = message.payload;

            if (subscribersRef.current[type]) {
                subscribersRef.current[type].forEach(callback => callback(payload));
            }
        };

        socket.onclose = () => {
            console.log('WebSocket Disconnected. Reconnecting...');
            setTimeout(connect, 3000);
        };

        socketRef.current = socket;
    }, [isAuthenticated]);

    useEffect(() => {
        connect();
        return () => {
            socketRef.current?.close();
        };
    }, [connect, groupID]);

    const sendMessage = useCallback((msg: unknown) => {
        if (socketRef.current?.readyState === WebSocket.OPEN) {
            socketRef.current.send(JSON.stringify(msg));
        }
    }, []);

    const subscribe = useCallback((type: string, callback: (payload: Record<string, unknown>) => void) => {
        if (!subscribersRef.current[type]) {
            subscribersRef.current[type] = [];
        }
        subscribersRef.current[type].push(callback);

        return () => {
            if (subscribersRef.current[type]) {
                subscribersRef.current[type] = subscribersRef.current[type].filter(cb => cb !== callback);
            }
        };
    }, []);

    return (
        <WebSocketContext.Provider value={{ sendMessage, subscribe }}>
            {children}
        </WebSocketContext.Provider>
    );
};
