export interface CafeOrder{
    id: string;
    name: string;
    description: string;
    category: string;
    energy_cost: number;
    reward_xp: number;
    required_level: number;
}

export interface UserOrder {
    id: number;
    user_id: number;
    cafe_order_id: number;
    status: 'pending' | 'completed';
    created_at: string;

    cafe_order?:CafeOrder
}