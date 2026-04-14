import { UserStats } from '@/types/user';
import { UserOrder } from '@/types/user-order';
import { createContext } from 'react';

interface GameContextType {
    user: UserStats | null;
    orders: UserOrder[];
    completeOrder: (orderId: number) => Promise<void>;
    loading: boolean;
    refresh: () => void; //function to refresh user data
    addXP: (amount: number) => Promise<void>;
}



export const GameContext = createContext<GameContextType | undefined>(undefined);
