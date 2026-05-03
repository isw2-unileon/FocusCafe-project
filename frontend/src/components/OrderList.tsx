import { completeOrder, getUserOrders } from '@/services/user_order_service';
import { getRemoteUserStats } from '@/services/user_service';
import {UserOrder} from '@/types/user-order';
import { Coffee, Zap } from 'lucide-react';
import { useEffect, useState } from 'react';

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Card } from "@/components/ui/card";
import { useAuth } from '@/context/AuthContext';
import { showLevelUpModal, showOrderServedToast, showXpToast } from '@/lib/notifications';



export const OrderList = () => {
    const { userStats, setUserStats } = useAuth();
    const [orders, setOrders] = useState<UserOrder[]>([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchOrders = async () => {
            try {
                const fetchedOrders = await getUserOrders();
                setOrders(fetchedOrders);
            } catch (error) {
                console.error('Failed to fetch orders:', error);
            } finally {
                setLoading(false);
            }
        };
        fetchOrders();
    }, []);

    const handleComplete = async (order: UserOrder) =>{
        try{
            const orderId = order.id;
            const levelBefore = userStats?.level ?? 0
            await completeOrder(orderId);
            
            showOrderServedToast();

            const stats = await getRemoteUserStats();
            setUserStats(stats);
            setOrders(prev => prev.filter(o=> o.id !== orderId));

            if (stats.level > levelBefore) {
                showLevelUpModal(stats.level);
            } else {
                // Si no subió de nivel, solo mostramos el brindis (toast) de XP
                showXpToast(order.cafe_order?.reward_xp ?? 0);
            }
            
            console.log(`Order ${orderId} completed`);
        }catch(error){
            alert("Error completing order")
            console.error(error)
        }
    };

    if (loading) return <div className="p-4 text-center">Loading orders...</div>;

    const userEnergy = userStats?.energy ?? 0;

    return (
        <Card className="w-full bg-orange-50 h-full overflow-y-auto">
            <Accordion type="single" collapsible defaultValue="orders-section">
                <AccordionItem value="orders-section">
                    
                    {/* Header */}
                    <AccordionTrigger className="hover:no-underline px-6 py-4 flex items-center">
                        <div className="flex items-center gap-3 w-full pr-4">
                            <div className="p-4 bg-white rounded-xl shadow-inner">
                                <Coffee className="text-orange-500" size={22} />
                            </div>
                            <h2 className="text-lg font-bold text-stone-800 uppercase tracking-tight">Pending orders</h2>
                        </div>
                    </AccordionTrigger>

                    {/* Order list */}
                    <AccordionContent className="px-6 pb-6 pt-2">
                        <div className="space-y-3 pt-4 border-t border-stone-100">
                            {orders.length === 0 ? (
                                <p className="text-stone-400 italic text-center py-4 text-sm">No orders yet!</p>
                            ) : (
                                orders.slice(0,3).map((order) => {
                                    const canAfford = userEnergy >= (order.cafe_order?.energy_cost ?? 0);
                                    return (
                                        <div 
                                            key={order.id} 
                                            className={`flex items-center justify-between p-4 rounded-xl border border-stone-50 bg-stone-50/30 ${!canAfford && 'opacity-60'}`}
                                        >
                                            <div className="flex flex-col gap-1">
                                                <span className="font-semibold text-stone-700">{order.cafe_order?.name}</span>
                                                <span className="text-[10px] font-semibold text-stone-500">{order.cafe_order?.description}</span>
                                                <span className="text-[10px] font-bold text-orange-500 uppercase">+{order.cafe_order?.reward_xp} XP</span>
                                            </div>

                                            <div className="flex items-center gap-4">
                                                <div className="flex items-center gap-1 text-stone-600">
                                                    <span className="text-sm font-bold">{order.cafe_order?.energy_cost}</span>
                                                    <Zap size={14} className={"text-amber-500"} />
                                                </div>

                                                <button
                                                    onClick={() => handleComplete(order)}
                                                    disabled={!canAfford}
                                                    className={`px-4 py-2 rounded-lg text-xs font-bold transition-all ${
                                                        canAfford 
                                                            ? 'bg-white text-stone-900 border-2 hover:bg-orange-600 active:scale-95' 
                                                            : 'bg-white-500 border-2 text-stone-80 cursor-not-allowed'
                                                    }`}
                                                >
                                                    {canAfford ? 'Complete' : 'No energy'}
                                                </button>
                                            </div>
                                        </div>
                                    )
                                })
                            )}
                        </div>
                    </AccordionContent>
                </AccordionItem>
            </Accordion>
        </Card>
    
    );
};

