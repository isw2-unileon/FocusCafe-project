import { useGame } from '@/context/useGame';
import { UserOrder } from '@/types/user-order';
import { ClipboardClock } from 'lucide-react';


export const OrderList = () => {
    const {user, orders, completeOrder, loading} = useGame();

    if(loading) return <div className="p-4 text-center">Loading orders...</div>;
    if (orders.length === 0) return <div className="p-4 text-gray-400 italic">Wow! There are no pending orders</div>;

    return (
        <div className="grid gap-4 p-4">
            <div className="flex items-center gap-2">
                <ClipboardClock className="text-orange-500" size={24} />
                <h2 className="text-xl font-bold text-brown-800">Pending orders</h2>
            </div>
            {orders.map((order: UserOrder) => (
                <OrderCard 
                key={order.id} 
                order={order} 
                userEnergy={user?.energy || 0}
                onComplete={completeOrder}
                />
            ))}
        </div>
    );
};

const OrderCard = ({
    order, 
    userEnergy, 
    onComplete
}: {
    order: UserOrder, 
    userEnergy: number, 
    onComplete: (orderId: number) => void
}) => {
    const canAfford = userEnergy >= order.energy_cost;

    return (
        <div className={`flex justify-between items-center p-4 border rounded-lg shadow-sm bg-white ${!canAfford ? 'opacity-70' : ''}`}>
            <div className="flex flex-col">
                <span className="text-lg font-semibold">{order.name}</span>
                <div className="flex gap-3 text-sm">
                <span className="text-blue-600 font-medium">⚡ {order.energy_cost}</span>
                <span className="text-green-600 font-medium">✨ +{order.reward_xp} XP</span>
                </div>
            </div>

            <button
                onClick={() => onComplete(order.id)}
                disabled={!canAfford}
                className={`px-4 py-2 rounded-full font-bold transition-colors ${
                canAfford 
                    ? 'bg-orange-500 text-white hover:bg-orange-600' 
                    : 'bg-gray-200 text-gray-500 cursor-not-allowed'
                }`}
            >
                {canAfford ? 'Complete' : 'Without energy'}
            </button>
        </div>
    )
}
