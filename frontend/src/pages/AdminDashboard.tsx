import React, { useState } from 'react';
import { useNavigate } from "react-router-dom";
import { UserPlus, Trash2, Shield, Search, ArrowLeft, Mail, ChevronRight } from 'lucide-react';
import { useAuth } from "@/context/AuthContext";

const AdminDashboard = () => {
    const { logout } = useAuth();
    const navigate = useNavigate();
    const [searchTerm, setSearchTerm] = useState("");

    const [users, setUsers] = useState([
        { id: 1, first_name: "Admin", email: "admin@cafe.com", role: "ADMIN", level: 99 },
        { id: 2, first_name: "Pepa", email: "pepa@estudiante.com", role: "USER", level: 3 },
    ]);

    const handleLogout = () => {
        logout();
        navigate('/');
    };

    return (
        <div className="min-h-screen bg-stone-100 p-6">
            <div className="max-w-5xl mx-auto">
                
                <div className="flex justify-between items-center mb-8">
                    <h2 className="text-2xl font-black text-stone-800 flex items-center gap-2">
                        <Shield className="text-orange-600" size={28}/> Staff Management
                    </h2>
                    
                    <div className="flex items-center gap-4">

                        <button 
                            onClick={handleLogout} 
                            className="text-stone-400 hover:text-red-500 font-bold text-sm transition-colors"
                        >
                            Logout
                        </button>
                    </div>
                </div>

                {/* Finder*/}
                <div className="bg-[#fdfaf7] rounded-[2rem] p-5 mb-8 border-4 border-white shadow-xl flex items-center gap-4">
                    <Search className="text-stone-300" size={24} />
                    <input 
                        type="text" 
                        placeholder="Search by name or email..." 
                        className="bg-transparent w-full outline-none font-bold text-lg text-stone-700"
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>

                {/* User List */}
                <div className="grid gap-4 mb-12">
                    {users.map((user) => (
                        <div key={user.id} className="bg-[#fdfaf7] rounded-[2.5rem] p-6 border-4 border-white shadow-lg flex items-center justify-between">
                            <div className="flex items-center gap-5">
                                <div className="w-16 h-16 bg-stone-200 rounded-2xl flex items-center justify-center font-black text-stone-500 text-2xl border-2 border-white shadow-inner">
                                    {user.first_name[0]}
                                </div>
                                <div>
                                    <div className="flex items-center gap-2">
                                        <h3 className="text-xl font-black text-stone-800">{user.first_name}</h3>
                                            <span className={`text-[10px] font-black px-2 py-0.5 rounded-full border ${user.role === 'ADMIN' ? 'bg-orange-50 text-orange-600 border-orange-200' : 'bg-white text-stone-400 border-stone-200'}`}>
                                                {user.role}
                                        </span>
                                    </div>
                                    <p className="text-stone-400 font-bold text-sm flex items-center gap-1">
                                        <Mail size={14}/> {user.email}
                                    </p>
                                </div>
                            </div>

                            <div className="flex items-center gap-6">
                                <div className="text-right">
                                    <p className="text-[10px] font-black text-stone-300 uppercase tracking-widest text-center">Level</p>
                                    <p className="text-2xl font-black text-stone-700">{user.level}</p>
                                </div>
                                <button className="bg-white p-4 rounded-2xl text-stone-200 hover:text-red-500 border border-stone-100 shadow-sm transition-all active:scale-90">
                                    <Trash2 size={24} />
                                </button>
                            </div>
                        </div>
                    ))}
                </div>

                <div className="flex justify-center mt-10">
                    <button 
                        className="bg-white text-stone-900 px-12 py-6 rounded-2xl font-black text-xl shadow-2xl hover:bg-orange-600 hover:text-white transition-all flex items-center gap-3 active:scale-95 group border-8 border-white"
                    >
                        <UserPlus size={28}/> HIRE NEW STAFF
                        <ChevronRight className="group-hover:translate-x-1 transition-transform" />
                    </button>
                </div>

            </div>
        </div>
    );
};

export default AdminDashboard;