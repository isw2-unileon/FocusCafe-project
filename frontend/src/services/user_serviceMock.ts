import { UserStats } from "@/types/user";
import { UserOrder } from "@/types/user-order";

const mockUserStats: UserStats = {
    id: 1, // Just a mock implementation, using token length as id
    first_name: "pepe",
    energy: 100,
    max_energy: 500,
    xp: 200,
    level: 5
};

let userOrders: UserOrder[] = [
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
        energy_cost: 800,
        reward_xp: 60,
        created_at: new Date().toISOString()
    }
];

export function getRemoteUserStats(): UserStats {
    return mockUserStats;
}

export async function getUserOrders(): Promise<UserOrder[]> {
    //console.log("Mock: Getting orders:", token);
    // Simulate network delay
    await new Promise(resolve => setTimeout(resolve, 10));
    return [...userOrders];
}

export async function completeOrder(orderId: number): Promise<UserStats> {
        // Simulate network delay
        userOrders = userOrders.filter(order => order.id !== orderId);

        await new Promise(resolve => setTimeout(resolve, 10));
        console.log("Mock: order completed, returning updated user stats");
        return mockUserStats;
    }

