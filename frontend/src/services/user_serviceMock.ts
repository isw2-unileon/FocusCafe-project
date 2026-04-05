import { UserStats } from "@/types/user";
import { IUserService } from "@/services/interfaces/IUserService";
import { UserOrder } from "@/types/user-order";


export class UserServiceMock implements IUserService {
    private mockUserStats: UserStats = {
            id: 1, // Just a mock implementation, using token length as id
            name: "pepe",
            energy: 100,
            max_energy: 500,
            level: 5
    };

    private userOrders: UserOrder[] = [
        {
            id: 1,
            user_id: 1,
            order_id: 1,
            status: 'pending',
            name: "Order 1",
            category: "Category A",
            energy_cost: 50,
            reward_xp: 100,
            created_at: new Date().toISOString()
        },

        {
            id: 2,
            user_id: 1,
            order_id: 2,
            category: "Category B",
            status: 'pending',
            name: "Order 2",
            energy_cost: 30,
            reward_xp: 60,
            created_at: new Date().toISOString()
        }
    ];

    async getRemoteUserStats(token: string): Promise<UserStats> {
        return this.mockUserStats;
    }

    async getUserOrders(token: string): Promise<UserOrder[]> {
        //console.log("Mock: Getting orders:", token);
        // Simulate network delay
        //await new Promise(resolve => setTimeout(resolve, 10));
        return [...this.userOrders];
    }

    async completeOrder(orderId: number, token: string): Promise<UserStats> {
        console.log(`Mock: Order completed with id: ${orderId} and token: ${token}`);
        // Simulate network delay
        this.userOrders = this.userOrders.filter(order => order.id !== orderId);

        await new Promise(resolve => setTimeout(resolve, 10));
        console.log("Mock: order completed, returning updated user stats");
        return this.mockUserStats;
    }
}