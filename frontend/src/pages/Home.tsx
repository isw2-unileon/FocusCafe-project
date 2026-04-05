import { useNavigate } from "react-router-dom";
import { Zap, Coffee, Trophy, Users, BookOpen, ChevronRight } from 'lucide-react';
import { StatCard } from "../components/StatCard";
import { useGame } from "@/context/GameContext";

const Home = () => {
    const {user, loading} = useGame();
    const navigate = useNavigate();
    
    if(loading) return <div className="min-h-screen flex items-center justify-center">Loading Caffe Salon...</div>;
    if(!user) return <div className="min-h-screen flex items-center justify-center">User data not found. Please log in again.</div>;


    //TO-DO: Add an icon for the Dashboard interface, maybe a graduation cap or a book?
    return (
        <div className="min-h-screen bg-stone-100 p-6">
            <div className="max-w-5xl mx-auto">
                {/* Header Section */}
                <div className="flex justify-between items-center mb-8">
                    <h2 className="text-2xl font-black text-stone-800 flex items-center gap-2">
                        ☕ Welcome back, {`${user.name}`}!
                        <span className="text-sm font-medium bg-white px-3 py-1 rounded-full border">
                            Level {`${user.level}`}
                        </span>
                    </h2>
                    <button 
                        onClick={() => navigate('/')} 
                        className="text-stone-400 hover:text-red-500 font-bold text-sm transition-colors"
                    >
                        Logout
                    </button>
                </div>

                {/* Stats Grid */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
                    <StatCard icon={<Zap className="text-yellow-500" size={20}/>} label="Energy" value={`${user.energy} / ${user.max_energy}`} color="bg-yellow-50" />
                    {/*<StatCard icon={<Coffee className="text-orange-700" size={20}/>} label="Orders" value={ordersReady} color="bg-orange-50" />
                    <StatCard icon={<Trophy className="text-amber-500" size={20}/>} label="Ranking" value={ranking} color="bg-amber-50" />
                    <StatCard icon={<Users className="text-blue-500" size={20}/>} label="Online" value={onlineUsers} color="bg-blue-50" />*/}
                </div>

                {/* Main Action Area */}
                <div className="bg-white rounded-[3rem] shadow-2xl overflow-hidden border-8 border-white relative group min-h-[400px]">
                    <img 
                        src="https://images.unsplash.com/photo-1554118811-1e0d58224f24?auto=format&fit=crop&q=80&w=1200" 
                        className="absolute inset-0 object-cover w-full h-full opacity-40 group-hover:scale-105 transition-transform duration-[3000ms]"
                        alt="Cafe Interior"
                    />
                    <div className="absolute inset-0 bg-gradient-to-t from-stone-900/60 via-transparent to-transparent" />
                    
                    <div className="absolute bottom-12 left-0 right-0 flex flex-col items-center">
                        <button 
                        onClick={() => navigate('/study')}
                        className="bg-white text-stone-900 px-12 py-6 rounded-2xl font-black text-xl shadow-2xl hover:bg-orange-600 hover:text-white transition-all flex items-center gap-3 active:scale-95 group"
                        >
                        <BookOpen size={28} /> BREW COFFEE (STUDY)
                        <ChevronRight className="group-hover:translate-x-1 transition-transform" />
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Home;
