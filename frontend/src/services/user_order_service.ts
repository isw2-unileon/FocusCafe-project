import { UserStats } from "@/types/user";
import { UserOrder } from "@/types/user-order";
import { apiFetch } from "@/services/api_client";

//Prefix for all the routes
const PATH = "/users/me/orders";


export async function getUserOrders(): Promise<UserOrder[]> {
    return apiFetch(`${PATH}`);
}

export async function completeOrder(orderId: number): Promise<UserStats> {
    return apiFetch(`${PATH}/${orderId}/complete`,{
         method: 'POST',
    })
}
