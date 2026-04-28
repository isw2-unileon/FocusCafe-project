import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Zap, BookOpen, ChevronRight, Award } from 'lucide-react';
import { StatCard } from "../components/StatCard";
import { getRemoteUserStats } from "@/services/user_service";
import { OrderList } from "@/components/OrderList";
import { AvatarDashboard } from "@/components/AvatarDashboard";
import { useAuth } from "@/context/AuthContext";
import { UserStats } from "@/types/user";

const Home = () => {
    const {logout, setUserStats } = useAuth();
    const [userStats, setUserStatsHome] = useState<UserStats | null>(null);
    const [loading, setLoading] = useState(true);
    const navigate = useNavigate();
    
    useEffect(() => {
        getRemoteUserStats()
        .then((data) => {
            setUserStats(data);
            setUserStatsHome(data);
        })
        .catch((err) => {
            console.error("Error loading stats:", err);
            logout();
            navigate("/login");
        })
        .finally(() => {
            setLoading(false);
        });
    }, [navigate, setUserStats]);


    const handleLogout = () => {
        logout();
        navigate('/');
    };

    if (loading) return <div className="min-h-screen flex items-center justify-center">Loading Cafeteria...</div>;

    if (!userStats) return <div className="min-h-screen flex items-center justify-center text-center">
        Expired session.<br/>Please, login again.
    </div>;

    const handleStudySessionClick = () => {
        navigate("/study")
    }

    return (
        <div className="min-h-screen bg-stone-100 p-6">
            <div className="max-w-5xl mx-auto">
                {/* Header Section */}
                <div className="flex justify-between items-center mb-8">
                    <h2 className="text-2xl font-black text-stone-800 flex items-center gap-2">
                        ☕ Welcome back, {`${userStats.first_name}`}!
                        <span className="text-sm font-medium bg-white px-3 py-1 rounded-full border">
                            Level {`${userStats.level}`}
                        </span>
                    </h2>
                    <div className="flex items-center gap-4">
                        <Link 
                            to="/dashboard"
                            className="transition-transform hover:scale-105 active:scale-95"
                        >
                            <AvatarDashboard/>
                        </Link>

                        <button 
                            onClick={handleLogout} 
                            className="text-stone-400 hover:text-red-500 font-bold text-sm transition-colors"
                        >
                            Logout
                        </button>
                    </div>
                    
                </div>

                {/* Stats Grid */}
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8 items-stretch">
                    <StatCard 
                        title="Player Stats"
                        stats={[
                            {
                            icon: <Zap className="text-yellow-500" size={20}/>,
                            label: "Energy",
                            current: userStats.energy,
                            max: userStats.max_energy,
                            barColor: "bg-yellow-500"
                            },
                            {
                            icon: <Award className="text-blue-500" size={20}/>,
                            label: "Experience",
                            current: userStats.xp,
                            max: userStats.max_energy, // O el valor que corresponda a XP
                            barColor: "bg-blue-500"
                            }
                        ]}
                        color = "bg-yellow-50"
                    />
                    <section className="col-span-2 md:col-span-2">
                        <OrderList />
                    </section>
                    {/*<StatCard icon={<Trophy className="text-amber-500" size={20}/>} label="Ranking" value={`${user.ranking}`} color="bg-amber-50" />
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
                        onClick={handleStudySessionClick}
                        className="bg-white text-stone-900 px-12 py-6 rounded-2xl font-black text-xl shadow-2xl hover:bg-orange-600 hover:text-white transition-all flex items-center gap-3 active:scale-95 group"
                        >
                        <BookOpen size={28}/> BREW COFFEE (STUDY)
                        <ChevronRight className="group-hover:translate-x-1 transition-transform" />
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};

export default Home;
