import { createContext, useContext, useState, useEffect, ReactNode } from "react";
import { UserStats } from "@/types/user";
import { getRemoteUserStats } from "@/services/user_service";

interface GameContextType {
    user: UserStats | null;
    loading: boolean;
    refresh: () => void; //function to refresh user data
}

const GameContext = createContext<GameContextType | undefined>(undefined);

export const GameProvider = ({ children }: { children: ReactNode }) => {
    const [user, setUser] = useState<UserStats | null>(null);
    const [loading, setLoading] = useState<boolean>(true);

    const loadUserData = async ()=>{
        try{

            //TO-DO: Replace with actual user ID from authentication context
            const userId = Number(localStorage.getItem('userId')) || 123;

            const userData = await getRemoteUserStats(userId); 
            setUser(userData);
        } catch (error) {
            console.error("Error loading user data:", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {loadUserData()}, []);

    return (
        <GameContext.Provider value={{user, loading, refresh: loadUserData}}>
            {children}
        </GameContext.Provider>
    );
};

export const useGame = () => {
    const context = useContext(GameContext);
    if(!context){
        throw new Error("useGame must be used within a GameProvider");
    }
    return context;
}