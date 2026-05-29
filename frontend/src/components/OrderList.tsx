import { completeOrder, getUserOrders } from '@/services/user_order_service';
import { getRemoteUserStats } from '@/services/user_service';
import {UserOrder} from '@/types/user-order';
import { Coffee, Users, Zap } from 'lucide-react';
import { useCallback, useEffect, useState } from 'react';

import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Card } from "@/components/ui/card";
import { useAuth } from '@/context/AuthContext';
import { useWebSocket } from '@/context/WebSocketContext';
import { showLevelUpModal, showOrderServedToast, showXpToast } from '@/lib/notifications';



export const OrderList = ({ inGroup = false }: { inGroup?: boolean }) => {
    const { userStats, setUserStats } = useAuth();
    const { subscribe } = useWebSocket();
    const [orders, setOrders] = useState<UserOrder[]>([]);
    const [loading, setLoading] = useState(true);

    const fetchOrders = useCallback(async (showLoading = true) => {
        if (showLoading) setLoading(true);
        try {
            const fetchedOrders = await getUserOrders();
            setOrders(fetchedOrders);
            
            // Also refresh stats when orders are updated to reflect potential XP/Energy changes from collaborative work
            const stats = await getRemoteUserStats();
            setUserStats(stats);
        } catch (error) {
            console.error('Failed to fetch orders:', error);
        } finally {
            if (showLoading) setLoading(false);
        }
    }, [setUserStats]);

    useEffect(() => {
        fetchOrders(true);

        // Subscribe to real-time updates
        const unsubscribe = subscribe('ORDERS_UPDATED', (payload) => {
            console.log('Real-time update received: Triggering smart refresh');
            
            if (payload && payload.order_id) {
                showOrderServedToast();
                setOrders((currentOrders) => {
                    const updatedOrders = currentOrders.filter(order => order.id !== payload.order_id);
                    
                    if (inGroup) {
                        const remainingGroup = updatedOrders.filter(o => !!o.group_id);
                        if (remainingGroup.length === 0) {
                            fetchOrders(false)
                        }
                    } else {
                        const remainingIndividual = updatedOrders.filter(o => !o.group_id);
                        if (remainingIndividual.length === 0) {
                            fetchOrders(false);
                        }
                    }

                    return updatedOrders; 
                });

            } else {
                fetchOrders(false);
            }
    });

        return () => unsubscribe();
    }, [subscribe, fetchOrders]);

    /*const removeOrderFromUI = useCallback((orderId: number) => {
    setOrders((currentOrders) => {
        const updatedOrders = currentOrders.filter(order => order.id !== orderId);
        
        if (inGroup) {
            const remainingGroup = updatedOrders.filter(o => !!o.group_id);
            if (remainingGroup.length === 0) fetchOrders(false);
        } else {
            const remainingIndividual = updatedOrders.filter(o => !o.group_id);
            if (remainingIndividual.length === 0) fetchOrders(false);
        }
        
        return updatedOrders;
    });
}, [inGroup, fetchOrders]);*/

    const handleComplete = async (order: UserOrder) =>{
        try{
            const orderId = order.id;
            const levelBefore = userStats?.level ?? 0
            await completeOrder(orderId);
            
            showOrderServedToast();

            const stats = await getRemoteUserStats();
            setUserStats(stats);
            
            const remainingOrders = orders.filter(o => o.id !== orderId);
            setOrders(remainingOrders);

            if (inGroup) {
                const remainingGroupOrders = remainingOrders.filter(o => !!o.group_id);
                if (remainingGroupOrders.length === 0) {
                    await fetchOrders();
                }
            } else {
                const remainingIndividualOrders = remainingOrders.filter(o => !o.group_id);
                if (remainingIndividualOrders.length === 0) {
                    console.log("¡Se han acabado los pedidos individuales! Recargando...");
                    await fetchOrders();
                }
            }

            if (stats.level > levelBefore) {
                showLevelUpModal(stats.level);
            } else {
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

    const filteredOrders = orders.filter(order =>
        inGroup
            ? !!order.group_id
            : !order.group_id
    );

    /**
     * UI configuration depending on mode
     */
    const isGroup = inGroup;

    const title = isGroup ? "Group orders" : "Pending orders";
    const Icon = isGroup ? Users : Coffee;
    const iconColor = isGroup ? "text-blue-500" : "text-orange-500";

    return (
        <Card className="w-full bg-orange-50 h-full overflow-y-auto">
            <Accordion type="single" collapsible defaultValue="orders-section">
                <AccordionItem value="orders-section">
                    
                    {/* Header */}
                    <AccordionTrigger className="hover:no-underline px-6 py-4 flex items-center">
                        <div className="flex items-center gap-3 w-full pr-4">
                            <div className={"p-4 rounded-xl shadow-inner bg-white"}>
                                <Icon className={iconColor} size={22} />
                            </div>
                            <h2 className="text-lg font-bold text-stone-800 uppercase tracking-tight">{title}</h2>
                        </div>
                    </AccordionTrigger>

                    {/* Order list */}
                    <AccordionContent className="px-6 pb-6 pt-2">
                        <div className="space-y-3 pt-4 border-t border-stone-100">
                            {filteredOrders.length === 0 ? (
                                <p className="text-stone-400 italic text-center py-4 text-sm">No orders yet!</p>
                            ) : (
                                filteredOrders.slice(0,3).map((order) => {
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

