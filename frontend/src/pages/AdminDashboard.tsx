import { useState, useEffect, useMemo } from 'react';
import { useNavigate } from "react-router-dom";
import { UserPlus, Trash2, Shield, Search, Mail, ChevronRight, ArrowLeft, X } from 'lucide-react';
import { useAuth } from "@/context/AuthContext";
import { getAllUsers, createUser } from "@/services/user_service";
import { UserProfile } from "@/types/user-profile";

const AVATAR_COLORS = [
    { bg: 'bg-orange-200', text: 'text-orange-700' },
    { bg: 'bg-yellow-200', text: 'text-yellow-700' },
    { bg: 'bg-red-200', text: 'text-red-700' },
    { bg: 'bg-blue-200', text: 'text-blue-700' },
    { bg: 'bg-green-200', text: 'text-green-700' },
    { bg: 'bg-stone-300', text: 'text-stone-700' },
];

function getAvatarColor(seed: string): { bg: string; text: string } {
    if (!seed) return AVATAR_COLORS[0]!;
    const index = seed.charCodeAt(0) % AVATAR_COLORS.length;
    return AVATAR_COLORS[index]!;
}

const AdminDashboard = () => {
    const { logout } = useAuth();
    const navigate = useNavigate();
    const [users, setUsers] = useState<UserProfile[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [searchQuery, setSearchQuery] = useState("");

    const filteredUsers = useMemo(() => {
        if (!searchQuery.trim()) return users;
        const q = searchQuery.toLowerCase();
        return users.filter(u => {
            const fullName = `${u.first_name || ''} ${u.last_name || ''}`.trim().toLowerCase();
            return (
                u.first_name?.toLowerCase().includes(q) ||
                u.last_name?.toLowerCase().includes(q) ||
                u.email?.toLowerCase().includes(q) ||
                fullName.includes(q)
            );
        });
    }, [users, searchQuery]);

    // Modal state
    const [showModal, setShowModal] = useState(false);
    const [firstName, setFirstName] = useState('');
    const [lastName, setLastName] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [formError, setFormError] = useState<string | null>(null);
    const [formLoading, setFormLoading] = useState(false);

    useEffect(() => {
        getAllUsers()
            .then((data) => {
                setUsers(data);
                setError(null);
            })
            .catch((err) => {
                console.error("Error fetching users:", err);
                setError("Failed to load users.");
            })
            .finally(() => setLoading(false));
    }, []);

    const handleLogout = () => {
        logout();
        navigate('/');
    };

    const resetForm = () => {
        setFirstName('');
        setLastName('');
        setEmail('');
        setPassword('');
        setConfirmPassword('');
        setFormError(null);
    };

    const handleCreateUser = async () => {
        setFormError(null);

        if (!firstName.trim() || !lastName.trim()) {
            setFormError("First name and last name are required.");
            return;
        }
        if (!email.trim()) {
            setFormError("Email is required.");
            return;
        }
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!emailRegex.test(email.trim())) {
            setFormError("Please enter a valid email address.");
            return;
        }
        if (password.length < 6) {
            setFormError("Password must be at least 6 characters long.");
            return;
        }
        if (password !== confirmPassword) {
            setFormError("Passwords do not match.");
            return;
        }

        setFormLoading(true);
        try {
            await createUser({
                first_name: firstName.trim(),
                last_name: lastName.trim(),
                email: email.trim(),
                password,
                confirm_password: confirmPassword,
            });

            // Refresh user list
            const updatedUsers = await getAllUsers();
            setUsers(updatedUsers);

            // Close modal and reset
            setShowModal(false);
            resetForm();
        } catch (err) {
            setFormError((err as Error).message || "Failed to create user.");
        } finally {
            setFormLoading(false);
        }
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-stone-100 flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-orange-600 mx-auto mb-4"></div>
                    <p className="text-stone-600 font-semibold">Loading staff...</p>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="min-h-screen bg-stone-100 flex items-center justify-center p-6">
                <div className="bg-white rounded-2xl p-8 shadow-lg text-center max-w-md">
                    <h2 className="text-2xl font-black text-stone-800 mb-4">Oops!</h2>
                    <p className="text-stone-600 mb-6">{error}</p>
                    <button
                        onClick={() => navigate('/home')}
                        className="bg-orange-600 text-white px-6 py-3 rounded-xl font-bold hover:bg-orange-700 transition-colors"
                    >
                        Back to Home
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-stone-100 p-6">
            <div className="max-w-5xl mx-auto">
                
                <div className="flex justify-between items-center mb-8">
                    <div className="flex items-center gap-4">
                        <button
                            onClick={() => navigate('/home')}
                            className="p-3 bg-white rounded-xl shadow-sm hover:bg-stone-50 transition-colors"
                        >
                            <ArrowLeft className="text-stone-600" size={24} />
                        </button>
                        <h2 className="text-2xl font-black text-stone-800 flex items-center gap-2">
                            <Shield className="text-orange-600" size={28}/> Staff Management
                        </h2>
                    </div>
                    
                    <button 
                        onClick={handleLogout} 
                        className="text-stone-400 hover:text-red-500 font-bold text-sm transition-colors"
                    >
                        Logout
                    </button>
                </div>

                {/* Finder*/}
                <div className="bg-[#fdfaf7] rounded-[2rem] p-5 mb-8 border-4 border-white shadow-xl flex items-center gap-4">
                    <Search className="text-stone-300" size={24} />
                    <input 
                        type="text" 
                        placeholder="Search by name or email..." 
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        className="bg-transparent w-full outline-none font-bold text-lg text-stone-700"
                    />
                </div>

                {/* User List */}
                <div className="grid gap-4 mb-12">
                    {filteredUsers.length === 0 && searchQuery.trim() && (
                        <div className="bg-white rounded-2xl p-8 shadow-lg text-center">
                            <p className="text-stone-400 font-bold">No users found matching "{searchQuery}"</p>
                        </div>
                    )}
                    {filteredUsers.map((user) => {
                        const color = getAvatarColor(user.first_name);
                        const displayName = `${user.first_name} ${user.last_name || ''}`.trim();
                        return (
                            <div key={user.id} className="bg-[#fdfaf7] rounded-[2.5rem] p-6 border-4 border-white shadow-lg flex items-center justify-between">
                                <div className="flex items-center gap-5">
                                    <div className={`w-16 h-16 ${color.bg} rounded-2xl flex items-center justify-center font-black ${color.text} text-2xl border-2 border-white shadow-inner`}>
                                        {user.first_name[0]}
                                    </div>
                                    <div>
                                        <div className="flex items-center gap-2">
                                            <h3 className="text-xl font-black text-stone-800">{displayName}</h3>
                                                <span className={`text-[10px] font-black px-2 py-0.5 rounded-full border ${user.role === 'admin' ? 'bg-orange-50 text-orange-600 border-orange-200' : 'bg-white text-stone-400 border-stone-200'}`}>
                                                    {user.role?.toUpperCase()}
                                            </span>
                                        </div>
                                        <p className="text-stone-400 font-bold text-sm flex items-center gap-1">
                                            <Mail size={14}/> {user.email}
                                        </p>
                                    </div>
                                </div>

                                <div className="flex items-center gap-6">
                                    <div className="hidden md:flex items-center gap-4">
                                        <div className="text-right">
                                            <p className="text-[10px] font-black text-stone-300 uppercase tracking-widest text-center">Energy</p>
                                            <p className="text-xl font-black text-yellow-600">{user.energy ?? 0}</p>
                                        </div>
                                        <div className="text-right">
                                            <p className="text-[10px] font-black text-stone-300 uppercase tracking-widest text-center">Member Since</p>
                                            <p className="text-sm font-black text-stone-600">{user.created_at ? new Date(user.created_at).toLocaleDateString('en-US', { month: 'short', year: 'numeric' }) : 'N/A'}</p>
                                        </div>
                                    </div>
                                    <div className="text-right">
                                        <p className="text-[10px] font-black text-stone-300 uppercase tracking-widest text-center">Level</p>
                                        <p className="text-2xl font-black text-stone-700">{user.level ?? 1}</p>
                                    </div>
                                    <button className="bg-white p-4 rounded-2xl text-stone-200 hover:text-red-500 border border-stone-100 shadow-sm transition-all active:scale-90">
                                        <Trash2 size={24} />
                                    </button>
                                </div>
                            </div>
                        );
                    })}
                </div>

                <div className="flex justify-center mt-10">
                    <button 
                        onClick={() => {
                            resetForm();
                            setShowModal(true);
                        }}
                        className="bg-white text-stone-900 px-12 py-6 rounded-2xl font-black text-xl shadow-2xl hover:bg-orange-600 hover:text-white transition-all flex items-center gap-3 active:scale-95 group border-8 border-white"
                    >
                        <UserPlus size={28}/> HIRE NEW STAFF
                        <ChevronRight className="group-hover:translate-x-1 transition-transform" />
                    </button>
                </div>

            </div>

            {/* Create User Modal */}
            {showModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={() => setShowModal(false)}>
                    <div 
                        className="bg-white rounded-[2.5rem] shadow-2xl p-10 max-w-md w-full border-b-8 border-orange-200"
                        style={{ animation: 'dropIn 0.3s ease-out' }}
                        onClick={(e) => e.stopPropagation()}
                    >
                        <div className="flex justify-between items-center mb-8">
                            <h2 className="text-2xl font-black text-stone-800">Create New Staff</h2>
                            <button
                                onClick={() => setShowModal(false)}
                                className="p-2 hover:bg-stone-100 rounded-xl transition-colors"
                            >
                                <X size={20} className="text-stone-400" />
                            </button>
                        </div>

                        <div className="space-y-4">
                            <div className="flex gap-3">
                                <input
                                    type="text"
                                    placeholder="First Name"
                                    value={firstName}
                                    onChange={(e) => setFirstName(e.target.value)}
                                    className="w-1/2 p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all"
                                />
                                <input
                                    type="text"
                                    placeholder="Last Name"
                                    value={lastName}
                                    onChange={(e) => setLastName(e.target.value)}
                                    className="w-1/2 p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all"
                                />
                            </div>
                            <input
                                type="email"
                                placeholder="Email"
                                value={email}
                                onChange={(e) => setEmail(e.target.value)}
                                className="w-full p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all"
                            />
                            <input
                                type="password"
                                placeholder="Password"
                                value={password}
                                onChange={(e) => setPassword(e.target.value)}
                                className="w-full p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all"
                            />
                            <input
                                type="password"
                                placeholder="Confirm Password"
                                value={confirmPassword}
                                onChange={(e) => setConfirmPassword(e.target.value)}
                                className="w-full p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all"
                            />

                            {formError && (
                                <p className="text-red-500 text-sm text-center font-medium">{formError}</p>
                            )}

                            <div className="flex gap-3 pt-4">
                                <button
                                    type="button"
                                    onClick={() => setShowModal(false)}
                                    className="flex-1 bg-white text-stone-800 px-6 py-4 rounded-2xl font-bold border-2 border-stone-200 hover:bg-stone-50 transition-colors"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="button"
                                    onClick={handleCreateUser}
                                    disabled={formLoading}
                                    className="flex-1 bg-orange-600 text-white px-6 py-4 rounded-2xl font-bold hover:bg-orange-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {formLoading ? 'Creating...' : 'Create User'}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default AdminDashboard;
