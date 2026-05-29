import { useState, useEffect, useMemo } from 'react';
import { useNavigate } from "react-router-dom";
import toast from 'react-hot-toast';
import { UserPlus, Trash2, Shield, Search, Mail, ChevronRight, X, AlertTriangle, LogOut, ChevronDown, Users, UsersRound, Crown } from 'lucide-react';
import { useAuth } from "@/context/AuthContext";
import { getAllUsers, createUser, deleteUser } from "@/services/user_service";
import { getAllGroups, adminDeleteGroup, GroupDetail } from "@/services/group_service";
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

type Tab = 'users' | 'groups';

const AdminDashboard = () => {
    const { logout, userId } = useAuth();
    const navigate = useNavigate();
    const [activeTab, setActiveTab] = useState<Tab>('users');

    // Users state
    const [users, setUsers] = useState<UserProfile[]>([]);
    const [usersLoading, setUsersLoading] = useState(true);
    const [usersError, setUsersError] = useState<string | null>(null);
    const [searchQuery, setSearchQuery] = useState("");

    // Groups state
    const [groups, setGroups] = useState<GroupDetail[]>([]);
    const [groupsLoading, setGroupsLoading] = useState(false);
    const [groupsError, setGroupsError] = useState<string | null>(null);
    const [expandedGroup, setExpandedGroup] = useState<number | null>(null);
    const [groupSearchQuery, setGroupSearchQuery] = useState("");

    // Modal state
    const [showModal, setShowModal] = useState(false);
    const [firstName, setFirstName] = useState('');
    const [lastName, setLastName] = useState('');
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [role, setRole] = useState('');
    const [formError, setFormError] = useState<string | null>(null);
    const [formLoading, setFormLoading] = useState(false);

    // Delete modal state
    const [showDeleteModal, setShowDeleteModal] = useState(false);
    const [userToDelete, setUserToDelete] = useState<UserProfile | null>(null);
    const [groupToDelete, setGroupToDelete] = useState<GroupDetail | null>(null);
    const [deleteLoading, setDeleteLoading] = useState(false);

    // eslint-disable-next-line react-hooks/exhaustive-deps
    useEffect(() => {
        setUsersLoading(true);
        getAllUsers()
            .then((data) => {
                setUsers(data);
                setUsersError(null);
            })
            .catch((err) => {
                console.error("Error fetching users:", err);
                setUsersError("Failed to load users.");
            })
            .finally(() => setUsersLoading(false));
    }, []);

    // eslint-disable-next-line react-hooks/exhaustive-deps
    useEffect(() => {
        if (activeTab === 'groups') {
            setGroupsLoading(true);
            getAllGroups()
                .then((data) => {
                    setGroups(data);
                    setGroupsError(null);
                })
                .catch((err) => {
                    console.error("Error fetching groups:", err);
                    setGroupsError("Failed to load groups.");
                })
                .finally(() => setGroupsLoading(false));
        }
    }, [activeTab]);

    const filteredUsers = useMemo(() => {
        let result = users;
        if (searchQuery.trim()) {
            const q = searchQuery.toLowerCase();
            result = users.filter(u => {
                const fullName = `${u.first_name || ''} ${u.last_name || ''}`.trim().toLowerCase();
                return (
                    u.first_name?.toLowerCase().includes(q) ||
                    u.last_name?.toLowerCase().includes(q) ||
                    u.email?.toLowerCase().includes(q) ||
                    fullName.includes(q)
                );
            });
        }
        return result.sort((a, b) => {
            if (a.id === userId) return -1;
            if (b.id === userId) return 1;
            return 0;
        });
    }, [users, searchQuery, userId]);

    const filteredGroups = useMemo(() => {
        const safeGroups = Array.isArray(groups) ? groups : [];
        if (!groupSearchQuery.trim()) {
            return safeGroups;
        }
        const q = groupSearchQuery.toLowerCase();
        return safeGroups.filter(g =>
            g.name.toLowerCase().includes(q) ||
            g.invite_code.toLowerCase().includes(q)
        );
    }, [groups, groupSearchQuery]);

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
        setRole('');
        setFormError(null);
    };

    const handleDeleteUser = async () => {
        if (!userToDelete) return;
        setDeleteLoading(true);
        try {
            await deleteUser(userToDelete.id);
            const updatedUsers = await getAllUsers();
            setUsers(updatedUsers);
            toast.success('User deleted successfully');
            setShowDeleteModal(false);
            setUserToDelete(null);
        } catch (err) {
            console.error("Error deleting user:", err);
            toast.error('Failed to delete user. Please try again.');
            setShowDeleteModal(false);
            setUserToDelete(null);
        } finally {
            setDeleteLoading(false);
        }
    };

    const handleDeleteGroup = async () => {
        if (!groupToDelete) return;
        setDeleteLoading(true);
        try {
            await adminDeleteGroup(groupToDelete.id);
            const updatedGroups = await getAllGroups();
            setGroups(updatedGroups);
            toast.success('Group deleted successfully');
            setShowDeleteModal(false);
            setGroupToDelete(null);
        } catch (err) {
            console.error("Error deleting group:", err);
            toast.error('Failed to delete group. Please try again.');
            setShowDeleteModal(false);
            setGroupToDelete(null);
        } finally {
            setDeleteLoading(false);
        }
    };

    const openDeleteUserModal = (user: UserProfile) => {
        if (user.id === userId) {
            toast.error('You cannot delete your own account');
            return;
        }
        setUserToDelete(user);
        setGroupToDelete(null);
        setShowDeleteModal(true);
    };

    const openDeleteGroupModal = (group: GroupDetail) => {
        setGroupToDelete(group);
        setUserToDelete(null);
        setShowDeleteModal(true);
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
                role,
            });

            const updatedUsers = await getAllUsers();
            setUsers(updatedUsers);

            setShowModal(false);
            resetForm();
            toast.success('User created successfully');
        } catch (err) {
            setFormError((err as Error).message || "Failed to create user.");
        } finally {
            setFormLoading(false);
        }
    };

    const isLoading = activeTab === 'users' ? usersLoading : groupsLoading;
    const error = activeTab === 'users' ? usersError : groupsError;

    if (isLoading) {
        return (
            <div className="min-h-screen bg-stone-100 flex items-center justify-center">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-orange-600 mx-auto mb-4"></div>
                    <p className="text-stone-600 font-semibold">Loading...</p>
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
                        <h2 className="text-2xl font-black text-stone-800 flex items-center gap-2">
                            <Shield className="text-orange-600" size={28}/> Staff Management
                        </h2>
                    </div>
                    
                    <button
                        onClick={handleLogout}
                        className="text-stone-400 hover:text-red-500 font-bold text-sm transition-colors flex items-center gap-2"
                    >
                        <LogOut size={18} />
                        Logout
                    </button>
                </div>

                {/* Tabs */}
                <div className="flex gap-2 mb-8">
                    <button
                        onClick={() => setActiveTab('users')}
                        className={`flex items-center gap-2 px-6 py-3 rounded-2xl font-bold transition-all ${
                            activeTab === 'users' 
                                ? 'bg-orange-600 text-white shadow-lg' 
                                : 'bg-white text-stone-600 hover:bg-stone-50'
                        }`}
                    >
                        <Users size={20} />
                        Staff
                    </button>
                    <button
                        onClick={() => setActiveTab('groups')}
                        className={`flex items-center gap-2 px-6 py-3 rounded-2xl font-bold transition-all ${
                            activeTab === 'groups' 
                                ? 'bg-orange-600 text-white shadow-lg' 
                                : 'bg-white text-stone-600 hover:bg-stone-50'
                        }`}
                    >
                        <UsersRound size={20} />
                        Teams
                    </button>
                </div>

                {/* Search */}
                <div className="bg-[#fdfaf7] rounded-[2rem] p-5 mb-8 border-4 border-white shadow-xl flex items-center gap-4">
                    <Search className="text-stone-300" size={24} />
                    <input 
                        type="text" 
                        placeholder={activeTab === 'users' ? "Search by name or email..." : "Search by team name or code..."}
                        value={activeTab === 'users' ? searchQuery : groupSearchQuery}
                        onChange={(e) => activeTab === 'users' ? setSearchQuery(e.target.value) : setGroupSearchQuery(e.target.value)}
                        className="bg-transparent w-full outline-none font-bold text-lg text-stone-700"
                    />
                </div>

                {/* Content */}
                {activeTab === 'users' ? (
                    <>
                        {/* User List */}
                        <div className="grid gap-4 mb-12">
                            {filteredUsers.length === 0 && searchQuery.trim() && (
                                <div className="bg-white rounded-2xl p-8 shadow-lg text-center">
                                    <p className="text-stone-400 font-bold">No users found matching "{searchQuery}"</p>
                                </div>
                            )}
                            {filteredUsers.length === 0 && !searchQuery.trim() && (
                                <div className="bg-white rounded-2xl p-10 shadow-lg text-center">
                                    <p className="text-stone-800 font-black text-xl mb-2">No staff yet</p>
                                    <p className="text-stone-400 font-medium">Hire your first team member to get started.</p>
                                </div>
                            )}
                            {filteredUsers.map((user) => {
                                const color = getAvatarColor(user.first_name);
                                const displayName = `${user.first_name} ${user.last_name || ''}`.trim();
                                return (
                                    <div key={user.id} className="bg-[#fdfaf7] rounded-[2.5rem] p-6 border-4 border-white shadow-lg flex items-center justify-between">
                                        <div className="flex items-center gap-5">
                                            <div className={`w-16 h-16 ${color.bg} rounded-2xl flex items-center justify-center font-black ${color.text} text-2xl border-2 border-white shadow-inner`}>
                                                {user.first_name?.[0] ?? '?'}
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
                                            <button
                                                data-testid={`delete-user-${user.email}`}
                                                onClick={() => openDeleteUserModal(user)}
                                                className="bg-white p-4 rounded-2xl text-stone-200 hover:text-red-500 border border-stone-100 shadow-sm transition-all active:scale-90"
                                            >
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
                    </>
                ) : (
                    <>
                        {/* Group List */}
                        <div className="grid gap-4 mb-12">
                            {filteredGroups.length === 0 ? (
                                groupSearchQuery.trim() ? (
                                    <div className="bg-white rounded-2xl p-8 shadow-lg text-center">
                                        <p className="text-stone-400 font-bold">No teams found matching "{groupSearchQuery}"</p>
                                    </div>
                                ) : (
                                    <div className="bg-white rounded-2xl p-10 shadow-lg text-center">
                                        <p className="text-stone-800 font-black text-xl mb-2">No teams yet</p>
                                        <p className="text-stone-400 font-medium">Teams will appear here when users create them.</p>
                                    </div>
                                )
                            ) : filteredGroups.map((group) => {
                                const isExpanded = expandedGroup === group.id;
                                const leader = group.members.find(m => m.id === group.leader_id);
                                return (
                                    <div key={group.id} className="bg-[#fdfaf7] rounded-[2.5rem] p-6 border-4 border-white shadow-lg">
                                        <div className="flex items-center justify-between cursor-pointer" onClick={() => setExpandedGroup(isExpanded ? null : group.id)}>
                                            <div className="flex items-center gap-5">
                                                <div className="w-16 h-16 bg-blue-200 rounded-2xl flex items-center justify-center font-black text-blue-700 text-2xl border-2 border-white shadow-inner">
                                                    <UsersRound size={28} />
                                                </div>
                                                <div>
                                                    <div className="flex items-center gap-2">
                                                        <h3 className="text-xl font-black text-stone-800">{group.name}</h3>
                                                        <span className="text-[10px] font-black px-2 py-0.5 rounded-full border bg-blue-50 text-blue-600 border-blue-200">
                                                            {group.members.length} MEMBERS
                                                        </span>
                                                    </div>
                                                    <p className="text-stone-400 font-bold text-sm">
                                                        Code: <span className="font-mono text-stone-600">{group.invite_code}</span>
                                                    </p>
                                                    {leader && (
                                                        <p className="text-stone-400 text-xs font-medium">
                                                            Leader: {leader.first_name} {leader.last_name}
                                                        </p>
                                                    )}
                                                </div>
                                            </div>

                                            <div className="flex items-center gap-4">
                                                <button
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        openDeleteGroupModal(group);
                                                    }}
                                                    className="bg-white p-4 rounded-2xl text-stone-200 hover:text-red-500 border border-stone-100 shadow-sm transition-all active:scale-90"
                                                >
                                                    <Trash2 size={24} />
                                                </button>
                                                <ChevronDown 
                                                    size={24} 
                                                    className={`text-stone-400 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                                                />
                                            </div>
                                        </div>

                                        {/* Expanded Members */}
                                        {isExpanded && (
                                            <div className="mt-4 pt-4 border-t border-stone-200">
                                                <p className="text-[10px] font-black text-stone-400 uppercase tracking-widest mb-3">Members</p>
                                                <div className="grid gap-2">
                                                    {group.members.map((member) => {
                                                        const isLeader = member.id === group.leader_id;
                                                        const color = getAvatarColor(member.first_name);
                                                        return (
                                                            <div key={member.id} className={`flex items-center gap-3 p-3 rounded-xl ${isLeader ? 'bg-orange-50 border border-orange-200' : 'bg-white border border-stone-100'}`}>
                                                                <div className={`w-10 h-10 ${color.bg} rounded-xl flex items-center justify-center font-black ${color.text} text-lg`}>
                                                                    {member.first_name?.[0] ?? '?'}
                                                                </div>
                                                                <div className="flex-1">
                                                                    <p className="font-bold text-stone-700 text-sm">
                                                                        {member.first_name} {member.last_name}
                                                                        {isLeader && (
                                                                            <span className="ml-2 text-[10px] font-black bg-orange-100 text-orange-600 px-2 py-0.5 rounded-full border border-orange-200">
                                                                                <Crown size={10} className="inline mr-1" />
                                                                                LEADER
                                                                            </span>
                                                                        )}
                                                                    </p>
                                                                    <p className="text-stone-400 text-xs">{member.email}</p>
                                                                </div>
                                                                <div className="text-right">
                                                                    <p className="text-[10px] font-black text-stone-400 uppercase">Level</p>
                                                                    <p className="text-lg font-black text-stone-700">{member.level}</p>
                                                                </div>
                                                            </div>
                                                        );
                                                    })}
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                );
                            })}
                        </div>
                    </>
                )}
            </div>

            {/* Delete Confirmation Modal */}
            {showDeleteModal && (userToDelete || groupToDelete) && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={() => setShowDeleteModal(false)}>
                    <div 
                        className="bg-white rounded-[2.5rem] shadow-2xl p-10 max-w-md w-full border-b-8 border-red-200"
                        onClick={(e) => e.stopPropagation()}
                    >
                        <div className="flex justify-between items-center mb-8">
                            <h2 className="text-2xl font-black text-stone-800 flex items-center gap-2">
                                <AlertTriangle className="text-red-500" size={28} /> 
                                {userToDelete ? 'Remove Staff' : 'Delete Team'}
                            </h2>
                            <button
                                onClick={() => setShowDeleteModal(false)}
                                className="p-2 hover:bg-stone-100 rounded-xl transition-colors"
                            >
                                <X size={20} className="text-stone-400" />
                            </button>
                        </div>

                        <div className="space-y-4">
                            <p className="text-stone-600 font-medium text-center">
                                Are you sure you want to delete{' '}
                                <span className="font-black text-stone-800">
                                    {userToDelete ? `${userToDelete.first_name} ${userToDelete.last_name}` : groupToDelete?.name}
                                </span>?
                            </p>
                            {groupToDelete && (
                                <p className="text-stone-400 text-sm text-center">
                                    This will remove all {groupToDelete.members.length} members from the team.
                                </p>
                            )}
                            <p className="text-stone-400 text-sm text-center">
                                This action cannot be undone.
                            </p>

                            <div className="flex gap-3 pt-4">
                                <button
                                    type="button"
                                    onClick={() => setShowDeleteModal(false)}
                                    className="flex-1 bg-white text-stone-800 px-6 py-4 rounded-2xl font-bold border-2 border-stone-200 hover:bg-stone-50 transition-colors"
                                >
                                    Cancel
                                </button>
                                <button
                                    type="button"
                                    onClick={userToDelete ? handleDeleteUser : handleDeleteGroup}
                                    disabled={deleteLoading}
                                    className="flex-1 bg-red-500 text-white px-6 py-4 rounded-2xl font-bold hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                    {deleteLoading ? 'Deleting...' : 'Delete'}
                                </button>
                            </div>
                        </div>
                    </div>
                </div>
            )}

            {/* Create User Modal */}
            {showModal && (
                <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={() => setShowModal(false)}>
                    <div 
                        className="bg-white rounded-[2.5rem] shadow-2xl p-10 max-w-md w-full border-b-8 border-orange-200"
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
                                data-testid="create-user-email"
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

                            <div className="relative">
                                <select
                                    value={role}
                                    onChange={(e) => setRole(e.target.value)}
                                    className="w-full p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all font-bold text-stone-700 appearance-none cursor-pointer pr-10"
                                >
                                    <option value="" disabled>User type</option>
                                    <option value="user">User</option>
                                    <option value="admin">Administrator</option>
                                </select>
                                <ChevronDown className="absolute right-4 top-1/2 -translate-y-1/2 text-stone-400 pointer-events-none" size={20} />
                            </div>

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
