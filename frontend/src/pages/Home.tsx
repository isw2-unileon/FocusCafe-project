import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Zap, BookOpen, ChevronRight, Award, Shield, LogIn, PlusCircle } from 'lucide-react';
import { StatCard } from "../components/StatCard";
import { getRemoteUserStats } from "@/services/user_service";
import { OrderList } from "@/components/OrderList";
import { AvatarDashboard } from "@/components/AvatarDashboard";
import { useAuth } from "@/context/AuthContext";

const Home = () => {
    const { logout, userStats, setUserStats, isAdmin } = useAuth();
    const [loading, setLoading] = useState(true);
    const [inviteCode, setInviteCode] = useState("");
    const [newGroupName, setNewGroupName] = useState("");
    const navigate = useNavigate();
    
    useEffect(() => {
        getRemoteUserStats()
        .then((data) => {
            setUserStats(data);
        })
        .catch((err) => {
            console.error("Error loading stats:", err);
            logout();
            navigate("/login");
        })
        .finally(() => {
            setLoading(false);
        });
    }, [navigate, setUserStats, logout]);


    const handleLogout = () => {
        logout();
        navigate('/');
    };

    const handleCreateGroup = async () => {
        // TODO: API Call createGroup(newGroupName)
        console.log("Creating group:", newGroupName);
    };

    const handleJoinGroup = async () => {
        // TODO: API Call joinGroup(inviteCode)
        console.log("Joining with code:", inviteCode);
    };

    const handleStudySessionClick = () => {
        navigate("/study")
    }

    if (loading) return <div className="min-h-screen flex items-center justify-center">Loading Cafeteria...</div>;

    if (!userStats) return <div className="min-h-screen flex items-center justify-center text-center">
        Expired session.<br/>Please, login again.
    </div>;

    return (
        <div className="min-h-screen bg-stone-100 p-6">
            <div className="max-w-6xl mx-auto">
                
                {/* --- UPDATED HEADER SECTION --- */}
                <div className="flex flex-col md:flex-row md:items-center justify-between mb-10 gap-4">
                    <div>
                        <h2 className="text-3xl font-black text-stone-800 flex items-center gap-3">
                            ☕ {userStats.first_name}
                            <span className="text-xs font-bold bg-orange-100 text-orange-700 px-3 py-1 rounded-full border border-orange-200 uppercase">
                                Lvl {userStats.level}
                            </span>
                        </h2>
                    </div>

                    <div className="flex flex-wrap items-center gap-4 bg-white/50 p-2 rounded-3xl border border-white backdrop-blur-sm">
                        
                        {/* Group Logic in Header */}
                        {!userStats.group ? (
                            <div className="flex items-center gap-2 px-2">
                                <div className="flex gap-1 border-r pr-4 border-stone-200">
                                    <input 
                                        className="bg-white border-none text-xs rounded-lg px-3 py-2 w-28 focus:ring-1 focus:ring-orange-500 outline-none shadow-sm"
                                        placeholder="Invite Code"
                                        value={inviteCode}
                                        onChange={(e) => setInviteCode(e.target.value.toUpperCase())}
                                    />
                                    <button onClick={handleJoinGroup} className="p-2 hover:bg-stone-800 hover:text-white rounded-lg transition-colors" title="Join Team">
                                        <LogIn size={18} />
                                    </button>
                                </div>
                                <div className="flex gap-1">
                                    <input 
                                        className="bg-white border-none text-xs rounded-lg px-3 py-2 w-28 focus:ring-1 focus:ring-orange-700 outline-none shadow-sm"
                                        placeholder="New Team"
                                        value={newGroupName}
                                        onChange={(e) => setNewGroupName(e.target.value)}
                                    />
                                    <button onClick={handleCreateGroup} className="p-2 hover:bg-orange-600 hover:text-white rounded-lg transition-colors text-orange-700" title="Create Team">
                                        <PlusCircle size={18} />
                                    </button>
                                </div>
                            </div>
                        ) : (
                            <div className="flex items-center gap-4 px-4 py-1">
                                <div className="flex flex-col">
                                    <span className="text-[10px] font-bold text-stone-400 uppercase leading-none">Team</span>
                                    <span className="text-sm font-black text-stone-700">{userStats.group.name}</span>
                                </div>
                                <div className="bg-orange-100 text-stone-600 px-3 py-1 rounded-lg">
                                    <span className="text-[10px] block opacity-70 leading-none">CODE</span>
                                    <span className="text-xs font-mono font-bold">{userStats.group.invite_code}</span>
                                </div>
                            </div>
                        )}

                        <div className="h-8 w-[1px] bg-stone-200 mx-1 hidden md:block" />

                        {/* Profile & Admin Icons */}
                        <div className="flex items-center gap-3 pr-2">
                            {isAdmin && (
                                <Link to="/adminDashboard" className="text-stone-400 hover:text-orange-600 transition-colors">
                                    <Shield size={22} />
                                </Link>
                            )}
                            <Link to="/dashboard" className="transition-transform hover:scale-110 active:scale-95">
                                <AvatarDashboard />
                            </Link>
                            <button onClick={handleLogout} className="text-stone-400 hover:text-red-500 font-bold text-xs uppercase tracking-tighter">
                                Logout
                            </button>
                        </div>
                    </div>
                </div>
                {/* Stats Grid */}
                <div className="grid grid-cols-2 md:grid-cols-3 gap-4 mb-8 items-stretch">
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
                                icon: <Award className="text-blue-500" size={20} />,
                                label: "Experience",
                                current: userStats.xp,
                                barColor: "bg-blue-500",
                            }
                        ]}
                        color = "bg-yellow-50"
                    />
                    <OrderList inGroup={false}/>
                    <OrderList inGroup={true}/>
                    {/* <OrderList/> */}
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
