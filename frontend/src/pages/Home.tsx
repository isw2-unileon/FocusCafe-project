import { useEffect, useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Zap, BookOpen, ChevronRight, Award, Shield, LogIn, PlusCircle, Copy, Check, Trash2, LogOut } from 'lucide-react';
import toast from 'react-hot-toast';
import { StatCard } from "../components/StatCard";
import { getRemoteUserStats } from "@/services/user_service";
import { createGroup, joinGroup, leaveGroup, deleteGroup } from "@/services/group_service";
import { OrderList } from "@/components/OrderList";
import { AvatarDashboard } from "@/components/AvatarDashboard";
import { useAuth } from "@/context/AuthContext";

const Home = () => {
    const { logout, userStats, setUserStats, isAdmin, userId } = useAuth();
    const [loading, setLoading] = useState(true);
    const [inviteCode, setInviteCode] = useState("");
    const [newGroupName, setNewGroupName] = useState("");
    const [isGroupLoading, setIsGroupLoading] = useState(false);
    const [showInviteModal, setShowInviteModal] = useState(false);
    const [createdGroupName, setCreatedGroupName] = useState("");
    const [createdGroupCode, setCreatedGroupCode] = useState("");
    const [copied, setCopied] = useState(false);
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
        if (!newGroupName.trim()) {
            toast.error("Team name is required");
            return;
        }
        setIsGroupLoading(true);
        try {
            const group = await createGroup(newGroupName.trim());
            setUserStats(prev => prev ? { ...prev, group } : prev);
            setCreatedGroupName(group.name);
            setCreatedGroupCode(group.invite_code);
            setShowInviteModal(true);
            setNewGroupName("");
            toast.success(`Team "${group.name}" created!`);
        } catch (err) {
            toast.error(err instanceof Error ? err.message : "Error creating team");
        } finally {
            setIsGroupLoading(false);
        }
    };

    const handleJoinGroup = async () => {
        if (!inviteCode.trim()) {
            toast.error("Invite code is required");
            return;
        }
        setIsGroupLoading(true);
        try {
            const group = await joinGroup(inviteCode.trim());
            setUserStats(prev => prev ? { ...prev, group } : prev);
            setInviteCode("");
            toast.success(`Joined team "${group.name}"!`);
        } catch (err) {
            toast.error(err instanceof Error ? err.message : "Invalid invite code");
        } finally {
            setIsGroupLoading(false);
        }
    };

    const handleLeaveTeam = async () => {
        if (!confirm("Are you sure you want to leave this team?")) return;
        setIsGroupLoading(true);
        try {
            await leaveGroup();
            setUserStats(prev => prev ? { ...prev, group: undefined } : prev);
            toast.success("You left the team");
        } catch (err) {
            toast.error(err instanceof Error ? err.message : "Error leaving team");
        } finally {
            setIsGroupLoading(false);
        }
    };

    const handleDeleteTeam = async () => {
        if (!confirm("Are you sure you want to delete this team? All members will be removed.")) return;
        setIsGroupLoading(true);
        try {
            await deleteGroup();
            setUserStats(prev => prev ? { ...prev, group: undefined } : prev);
            toast.success("Team deleted");
        } catch (err) {
            toast.error(err instanceof Error ? err.message : "Error deleting team");
        } finally {
            setIsGroupLoading(false);
        }
    };

    const handleCopyCode = async () => {
        try {
            await navigator.clipboard.writeText(createdGroupCode);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch {
            const textArea = document.createElement("textarea");
            textArea.value = createdGroupCode;
            document.body.appendChild(textArea);
            textArea.select();
            document.execCommand("copy");
            document.body.removeChild(textArea);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        }
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
                                <div className="flex flex-col border-r pr-4 border-stone-200">
                                    <div className="flex gap-1">
                                        <input
                                            className="bg-white border-none text-xs rounded-lg px-3 py-2 w-28 focus:ring-1 focus:ring-orange-500 outline-none shadow-sm disabled:opacity-50"
                                            placeholder="Invite Code"
                                            value={inviteCode}
                                            disabled={isGroupLoading}
                                            onChange={(e) => setInviteCode(e.target.value.toUpperCase())}
                                            onKeyDown={(e) => e.key === 'Enter' && handleJoinGroup()}
                                            data-testid="group-invite-input"
                                        />
                                        <button
                                            onClick={handleJoinGroup}
                                            disabled={isGroupLoading}
                                            className="p-2 hover:bg-stone-800 hover:text-white rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
                                            title="Join Team"
                                            data-testid="group-join-button"
                                        >
                                            <LogIn size={18} />
                                        </button>
                                    </div>
                                </div>
                                <div className="flex gap-1">
                                    <input
                                        className="bg-white border-none text-xs rounded-lg px-3 py-2 w-28 focus:ring-1 focus:ring-orange-700 outline-none shadow-sm disabled:opacity-50"
                                        placeholder="New Team"
                                        value={newGroupName}
                                        disabled={isGroupLoading}
                                        onChange={(e) => setNewGroupName(e.target.value)}
                                        onKeyDown={(e) => e.key === 'Enter' && handleCreateGroup()}
                                        data-testid="group-name-input"
                                    />
                                    <button
                                        onClick={handleCreateGroup}
                                        disabled={isGroupLoading}
                                        className="p-2 hover:bg-orange-600 hover:text-white rounded-lg transition-colors text-orange-700 disabled:opacity-40 disabled:cursor-not-allowed"
                                        title="Create Team"
                                        data-testid="group-create-button"
                                    >
                                        <PlusCircle size={18} />
                                    </button>
                                </div>
                            </div>
                        ) : (
                            <div className="flex items-center gap-4 px-4 py-1" data-testid="group-display">
                                <div className="flex flex-col">
                                    <span className="text-[10px] font-bold text-stone-400 uppercase leading-none">Team</span>
                                    <span className="text-sm font-black text-stone-700" data-testid="group-name">{userStats.group.name}</span>
                                </div>
                                <div className="bg-orange-100 text-stone-600 px-3 py-1 rounded-lg">
                                    <span className="text-[10px] block opacity-70 leading-none">CODE</span>
                                    <span className="text-xs font-mono font-bold" data-testid="group-code">{userStats.group.invite_code}</span>
                                </div>
                                {userId === userStats.group.leader_id ? (
                                    <button
                                        onClick={handleDeleteTeam}
                                        disabled={isGroupLoading}
                                        className="p-2 hover:bg-red-600 hover:text-white rounded-lg transition-colors text-red-500 disabled:opacity-40 disabled:cursor-not-allowed"
                                        title="Delete Team"
                                    >
                                        <Trash2 size={16} />
                                    </button>
                                ) : (
                                    <button
                                        onClick={handleLeaveTeam}
                                        disabled={isGroupLoading}
                                        className="p-2 hover:bg-stone-800 hover:text-white rounded-lg transition-colors text-stone-400 disabled:opacity-40 disabled:cursor-not-allowed"
                                        title="Leave Team"
                                    >
                                        <LogOut size={16} />
                                    </button>
                                )}
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

            {/* Success Modal for Created Group */}
            {showInviteModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
                    <div className="bg-white rounded-3xl p-8 max-w-md w-full mx-4 text-center shadow-2xl">
                        <h3 className="text-2xl font-black text-stone-800 mb-1">Team Created!</h3>
                        <p className="text-lg font-bold text-orange-600 mb-4">{createdGroupName}</p>
                        <p className="text-stone-600 mb-6">Share this code with your friends:</p>
                        
                        <div
                            className="bg-orange-50 border-2 border-orange-200 rounded-2xl p-6 mb-6 cursor-pointer hover:bg-orange-100 transition-colors relative"
                            onClick={handleCopyCode}
                            title="Click to copy"
                            data-testid="group-invite-code-display"
                        >
                            <span className="text-4xl font-mono font-black text-orange-600 tracking-widest block" data-testid="group-invite-code">
                                {createdGroupCode}
                            </span>
                            <div className="absolute top-2 right-2 text-orange-400">
                                {copied ? <Check size={20} /> : <Copy size={20} />}
                            </div>
                            {copied && (
                                <span className="text-xs text-orange-600 font-bold absolute bottom-1 left-0 right-0">Copied!</span>
                            )}
                        </div>
                        
                        <button
                            onClick={() => setShowInviteModal(false)}
                            className="bg-orange-600 text-white px-8 py-3 rounded-xl font-black text-lg hover:bg-orange-700 transition-colors"
                            data-testid="group-modal-done"
                        >
                            Done!
                        </button>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Home;
