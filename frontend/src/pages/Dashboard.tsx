import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Mail, User, Calendar, Zap, Award, LogOut } from 'lucide-react';
import { getCurrentProfile } from '@/services/user_service';
import { UserProfile } from '@/types/user-profile';
import { AvatarDashboard } from '@/components/AvatarDashboard';
import { useAuth } from '@/context/AuthContext';

function getRankInfo(level: number) {
  const ranks = [
    { title: 'Coffee Novice', color: 'bg-stone-400' },
    { title: 'Focus Apprentice', color: 'bg-emerald-500' },
    { title: 'Concentration Expert', color: 'bg-amber-500' },
    { title: 'Flow Master', color: 'bg-fuchsia-600' },
    { title: 'Zen Grandmaster', color: 'bg-purple-600' },
  ] as const;
  const index = Math.min(Math.floor((level - 1) / 3), ranks.length - 1);
  return ranks[index]!;
}

const Dashboard = () => {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();
  const { logout } = useAuth();

  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(() => {
    getCurrentProfile()
      .then(setProfile)
      .catch((err) => {
        console.error('Error loading profile:', err);
        setError('Could not load profile');
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-stone-100 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-orange-600 mx-auto mb-4"></div>
          <p className="text-stone-600 font-semibold">Loading profile...</p>
        </div>
      </div>
    );
  }

  if (error || !profile) {
    return (
      <div className="min-h-screen bg-stone-100 flex items-center justify-center p-6">
        <div className="bg-white rounded-2xl p-8 shadow-lg text-center max-w-md">
          <h2 className="text-2xl font-black text-stone-800 mb-4">Oops!</h2>
          <p className="text-stone-600 mb-6">{error || 'Profile not found'}</p>
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

  const handleLogout = () => {
    logout();
    navigate('/');
  };

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  return (
    <div className="min-h-screen bg-stone-100 p-6">
      <div className="max-w-5xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate('/home')}
              className="p-3 bg-white rounded-xl shadow-sm hover:bg-stone-50 transition-colors"
            >
              <ArrowLeft className="text-stone-600" size={24} />
            </button>
            <h1 className="text-3xl font-black text-stone-800 flex items-center gap-3">
              ☕ My Profile
              <span className="text-sm font-medium bg-white px-3 py-1 rounded-full border" data-testid="profile-level">
                Level {profile.level}
              </span>
            </h1>
          </div>
          <button
            onClick={handleLogout}
            className="text-stone-400 hover:text-red-500 font-bold text-sm transition-colors flex items-center gap-2"
          >
            <LogOut size={18} />
            Logout
          </button>
        </div>

        {/* Profile Card */}
        <div className="bg-white rounded-[3rem] shadow-2xl overflow-hidden border-8 border-white mb-6">
          {/* User Section: same color as Pending Orders card */}
          <div className="bg-orange-50 p-10">
            <div className="flex flex-col items-center text-center">
              {/* Avatar */}
              <div className="mb-5">
                <AvatarDashboard />
              </div>

              {/* Name */}
              <h2 className="text-3xl font-black text-stone-800 mb-1" data-testid="profile-full-name">
                {profile.first_name} {profile.last_name}
              </h2>
              {/* Username */}
              <p className="text-lg text-orange-600/80 font-semibold mb-4" data-testid="profile-username">
                @{profile.username}
              </p>

              {/* Rank Badge */}
              {(() => {
                const rank = getRankInfo(profile.level || 1);
                return (
                  <span className={`px-4 py-1.5 ${rank.color} text-white rounded-full text-sm font-bold shadow-sm`} data-testid="profile-rank">
                    {rank.title}
                  </span>
                );
              })()}

            </div>
          </div>

          {/* Stats & Info Section */}
          <div className="p-8">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Player Stats (inline, same style as Home's StatCard) */}
              <div className="bg-yellow-50 p-7 rounded-2xl flex flex-col">
                <div className="pt-4">
                  <h2 className="text-lg font-bold text-stone-800 uppercase tracking-widest mb-10">
                    Player Stats
                  </h2>
                </div>
                <div className="flex flex-col gap-6">
                  {/* Energy */}
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-gray-50 rounded-xl shadow-sm border border-black/5">
                      <Zap className="text-yellow-500" size={20} />
                    </div>
                    <div className="flex-1">
                      <div className="flex justify-between items-end mb-1">
                        <p className="text-[11px] font-bold uppercase text-gray-500">Energy</p>
                        <p className="text-sm font-black text-gray-800" data-testid="stat-energy-value">
                          {profile.progress?.energy ?? profile.energy ?? 0}{' '}
                          <span className="text-gray-400 font-medium">/ 500</span>
                        </p>
                      </div>
                      <div className="w-full h-2.5 bg-gray-100 rounded-full overflow-hidden" data-testid="stat-energy-bar">
                        <div
                          className="h-full bg-yellow-500 transition-all duration-500"
                          style={{
                            width: `${Math.min(
                              (((profile.progress?.energy ?? profile.energy ?? 0) / 500) * 100),
                              100
                            )}%`,
                          }}
                        />
                      </div>
                    </div>
                  </div>
                  {/* Experience */}
                  <div className="flex items-center gap-4">
                    <div className="p-3 bg-gray-50 rounded-xl shadow-sm border border-black/5">
                      <Award className="text-blue-500" size={20} />
                    </div>
                    <div className="flex-1">
                      <div className="flex justify-between items-center">
                        <p className="text-[11px] font-bold uppercase text-gray-500">Experience</p>
                        <p className="text-sm font-black text-gray-800" data-testid="stat-experience-value">
                          {profile.xp ?? 0} <span className="text-gray-400 font-medium">XP</span>
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Quick Info Card */}
              <div className="bg-stone-50 rounded-2xl border border-stone-100 p-7 flex flex-col justify-center gap-4">
                <div className="flex items-center gap-4 p-3 bg-white rounded-2xl shadow-sm">
                  <div className="p-3 bg-stone-50 rounded-lg">
                    <Mail className="text-stone-500" size={22} />
                  </div>
                  <div>
                    <p className="text-xs font-bold uppercase text-gray-400 tracking-tight">Email</p>
                    <p className="text-base font-semibold text-stone-800" data-testid="profile-email">{profile.email}</p>
                  </div>
                </div>
                <div className="flex items-center gap-4 p-3 bg-white rounded-2xl shadow-sm">
                  <div className="p-3 bg-stone-50 rounded-lg">
                    <User className="text-stone-500" size={22} />
                  </div>
                  <div>
                    <p className="text-xs font-bold uppercase text-gray-400 tracking-tight">Username</p>
                    <p className="text-base font-semibold text-stone-800" data-testid="profile-username-card">@{profile.username}</p>
                  </div>
                </div>
                <div className="flex items-center gap-4 p-3 bg-white rounded-2xl shadow-sm">
                  <div className="p-3 bg-stone-50 rounded-lg">
                    <Calendar className="text-stone-500" size={22} />
                  </div>
                  <div>
                    <p className="text-xs font-bold uppercase text-gray-400 tracking-tight">Member Since</p>
                    <p className="text-base font-semibold text-stone-800" data-testid="profile-member-since">{formatDate(profile.created_at)}</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex gap-4">
          <button
            onClick={() => navigate('/home')}
            className="flex-1 bg-white text-stone-800 px-6 py-4 rounded-2xl font-bold border-2 border-stone-200 hover:bg-stone-50 transition-colors"
          >
            Back to Home
          </button>
          <button
            onClick={() => navigate('/edit-profile')}
            className="flex-1 bg-orange-600 text-white px-6 py-4 rounded-2xl font-bold hover:bg-orange-700 transition-colors"
          >
            Edit Profile
          </button>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
