import { UserStats } from "@/types/user";
import { IUserService } from "@/services/interfaces/IUserService";
import { UserOrder } from "@/types/user-order";

const baseUrl = "/api/users";

export class UserService implements IUserService {
    async getRemoteUserStats(token: string): Promise<UserStats> {
        const headers: HeadersInit = {
        "Content-Type": "application/json",
        };

        //Add the authorization header with the token
        headers["Authorization"] = `Bearer ${token}`;

        const response = await fetch(
            `${baseUrl}/me/stats`,
            {headers}
        );
        if(!response.ok){
            throw new Error("Failed to fetch user stats");
        }
        return response.json();
    }

    async getUserOrders(token: string): Promise<UserOrder[]> {
        
        const headers: HeadersInit = {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
        };

        const response = await fetch(`${baseUrl}/me/orders/active`, {headers});

        if(!response.ok){
            throw new Error("Failed to obtain active orders");
        }
        return response.json();
    }

    async completeOrder(orderId: number, token: string): Promise<UserStats> {
        const headers: HeadersInit = {
            "Content-Type": "application/json",
            "Authorization": `Bearer ${token}`
        };

        const response = await fetch(`${baseUrl}/orders/${orderId}/complete`, {
            method: "POST",
            headers
        });
        if(!response.ok){
            throw new Error("Failed to complete order");
        }
        return response.json();
    }
}