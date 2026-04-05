export interface UserOrder {
    id: number;
    user_id: number;
    order_id: number;
    status: 'pending' | 'completed';

    name: string;
    category: string;
    energy_cost: number;
    reward_xp: number;
    created_at: string;
}