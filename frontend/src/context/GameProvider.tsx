import {useState, useEffect, ReactNode, useCallback } from "react";
import { UserStats } from "@/types/user";
import { GameContext } from "./GameContext";
import { UserService } from "@/services/user_service";
//import { UserServiceMock } from "@/services/user_serviceMock";
import { UserOrder } from "@/types/user-order";

//TO-DO: Implement authentication context to get actual user ID and token instead of using localStorage and mock data
const nodeEnv = process.env.NODE_ENV || "development";
//const userServ = nodeEnv === "production" ? new UserService() : new UserServiceMock();
console.log("Using user service:", nodeEnv === "production" ? "UserService" : "UserServiceMock");

export const GameProvider = ({ children }: { children: ReactNode }) => {
    const [user, setUser] = useState<UserStats | null>(null);
    const [loading, setLoading] = useState<boolean>(true);
    const [orders, setOrders] = useState<UserOrder[]>([]);

    //Load user data from backend (or mock) when the provider mounts
    const loadUserData = useCallback(async ()=>{
        try{

            //TO-DO: Replace with actual user ID from authentication context
            const userId = Number(localStorage.getItem('userId')) || 123;

            /*const [stats, userOrders] = await Promise.all([
                userServ.getRemoteUserStats(`token_for_user_${userId}`),
                userServ.getUserOrders(`token_for_user_${userId}`)
            ]);
            setOrders(userOrders);
            setUser(stats);*/
        } catch (error) {
            console.error("Error loading user data:", error);
        } finally {
            setLoading(false);
        }
    }, []);

    //
    /*
    const handleCompleteOrder = async (orderId: number) => {
        try{
            await userServ.completeOrder(orderId, `token_for_user_${user?.id}`);
            
            await loadUserData(); // Refresh user data after completing the order
        }catch(error){
            console.error("Error completing order:", error);
        }
    };


    const addXP = async (amount: number) => {
        try {
            // Opción A: Llamada real al backend de Go [cite: 16, 21]
            //await UserService.updateXP(user.id, amount); 
            
            // Opción B: Actualización local para pruebas
            console.log(`Adding ${amount} XP to user...`);
            await loadUserData(); // refresh user data to get updated XP from backend (or mock) after adding XP
        } catch (error) {
            console.error("Error updating XP:", error);
        }
    };

    useEffect(() => {
        loadUserData()}, [loadUserData]);

    return (
        <GameContext.Provider value={{
            user, 
            orders,
            completeOrder: handleCompleteOrder,
            loading, 
            refresh: loadUserData,
            addXP}}>
            {children}
        </GameContext.Provider>
    );*/
};
