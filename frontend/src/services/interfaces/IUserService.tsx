import {UserStats} from "@/types/user";
import { UserOrder } from "@/types/user-order";

export interface IUserService {
    getRemoteUserStats(token: string): Promise<UserStats>;
    getUserOrders (token: string): Promise<UserOrder[]>;
    completeOrder (orderId: number, token: string): Promise<UserStats>;
}