import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft, Mail, User, Calendar, Zap, Flame } from 'lucide-react';
import { getCurrentProfile } from '@/services/user_service';
import { UserProfile } from '@/types/user-profile';

const Dashboard = () => {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    getCurrentProfile()
      .then(setProfile)
      .catch((err) => {
        console.error('Error loading profile:', err);
        setError('No se pudo cargar el perfil');
      })
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <div className="min-h-screen bg-stone-100 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-orange-600 mx-auto mb-4"></div>
          <p className="text-stone-600 font-semibold">Cargando perfil...</p>
        </div>
      </div>
    );
  }

  if (error || !profile) {
    return (
      <div className="min-h-screen bg-stone-100 flex items-center justify-center p-6">
        <div className="bg-white rounded-2xl p-8 shadow-lg text-center max-w-md">
          <h2 className="text-2xl font-black text-stone-800 mb-4">Oops!</h2>
          <p className="text-stone-600 mb-6">{error || 'No se encontró el perfil'}</p>
          <button
            onClick={() => navigate('/home')}
            className="bg-orange-600 text-white px-6 py-3 rounded-xl font-bold hover:bg-orange-700 transition-colors"
          >
            Volver al inicio
          </button>
        </div>
      </div>
    );
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('es-ES', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  return (
    <div className="min-h-screen bg-stone-100 p-6">
      <div className="max-w-3xl mx-auto">
        {/* Header */}
        <div className="flex items-center gap-4 mb-8">
          <button
            onClick={() => navigate('/home')}
            className="p-3 bg-white rounded-xl shadow-sm hover:bg-stone-50 transition-colors"
          >
            <ArrowLeft className="text-stone-600" size={24} />
          </button>
          <h1 className="text-3xl font-black text-stone-800">Mi Perfil</h1>
        </div>

        {/* Profile Card */}
        <div className="bg-white rounded-3xl shadow-xl overflow-hidden border border-stone-200 mb-6">
          {/* Header Background */}
          <div className="h-32 bg-gradient-to-r from-orange-500 to-orange-600"></div>

          {/* Profile Content */}
          <div className="px-8 pb-8">
            {/* Avatar Section */}
            <div className="flex flex-col items-center -mt-16 mb-8">
              <div className="w-32 h-32 bg-white rounded-2xl shadow-lg border-4 border-stone-100 flex items-center justify-center">
                <User className="text-orange-600" size={64} strokeWidth={1.5} />
              </div>
            </div>

            {/* Name Section */}
            <div className="text-center mb-8">
              <h2 className="text-3xl font-black text-stone-800 mb-2">
                {profile.first_name} {profile.last_name}
              </h2>
              <p className="text-lg text-orange-600 font-semibold">@{profile.username}</p>
            </div>

            {/* Stats Grid */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
              {/* Energy Card */}
              {profile.progress && (
                <div className="bg-yellow-50 rounded-2xl p-4 text-center border border-yellow-100">
                  <div className="flex justify-center mb-3">
                    <div className="p-2 bg-yellow-100 rounded-lg">
                      <Zap className="text-yellow-600" size={24} />
                    </div>
                  </div>
                  <p className="text-xs font-bold uppercase text-gray-500 tracking-tight mb-1">Energía</p>
                  <p className="text-2xl font-black text-yellow-600">
                    {profile.progress.energy}
                  </p>
                </div>
              )}

              {/* Level Card */}
              {profile.progress && (
                <div className="bg-red-50 rounded-2xl p-4 text-center border border-red-100">
                  <div className="flex justify-center mb-3">
                    <div className="p-2 bg-red-100 rounded-lg">
                      <Flame className="text-red-600" size={24} />
                    </div>
                  </div>
                  <p className="text-xs font-bold uppercase text-gray-500 tracking-tight mb-1">Nivel</p>
                  <p className="text-2xl font-black text-red-600">{profile.progress.level}</p>
                </div>
              )}

              {/* Role Card */}
              <div className="bg-blue-50 rounded-2xl p-4 text-center border border-blue-100">
                <div className="flex justify-center mb-3">
                  <div className="p-2 bg-blue-100 rounded-lg">
                    <User className="text-blue-600" size={24} />
                  </div>
                </div>
                <p className="text-xs font-bold uppercase text-gray-500 tracking-tight mb-1">Rol</p>
                <p className="text-xl font-black text-blue-600 capitalize">{profile.role}</p>
              </div>

              {/* Join Date Card */}
              <div className="bg-green-50 rounded-2xl p-4 text-center border border-green-100">
                <div className="flex justify-center mb-3">
                  <div className="p-2 bg-green-100 rounded-lg">
                    <Calendar className="text-green-600" size={24} />
                  </div>
                </div>
                <p className="text-xs font-bold uppercase text-gray-500 tracking-tight mb-1">Miembro desde</p>
                <p className="text-sm font-black text-green-600">{formatDate(profile.created_at)}</p>
              </div>
            </div>

            {/* Divider */}
            <div className="border-t border-stone-200 my-8"></div>

            {/* Progress Section */}
            <div className="mb-8">
              <h3 className="text-xl font-black text-stone-800 mb-4">📊 Tu Progreso</h3>
              
              {profile.progress ? (
                <div className="bg-white rounded-2xl p-6 border border-stone-200 shadow-sm">
                  {/* Energy Bar */}
                  <div className="mb-6">
                    <p className="text-sm font-semibold text-stone-600 mb-3">Energía</p>
                    <div className="w-full bg-stone-200 rounded-full h-4 overflow-hidden">
                      <div 
                        className="bg-gradient-to-r from-orange-500 to-orange-600 h-4 rounded-full transition-all duration-500"
                        style={{ width: `${(profile.progress.energy / 500) * 100}%` }}
                      ></div>
                    </div>
                    <p className="text-xs text-stone-500 mt-2">
                      {profile.progress.energy} / 500
                    </p>
                  </div>

                  {/* Level */}
                  <div className="bg-gradient-to-r from-red-50 to-orange-50 rounded-xl p-4 border border-red-100">
                    <p className="text-sm font-semibold text-stone-600 mb-2">Nivel Actual</p>
                    <p className="text-3xl font-black text-red-600">{profile.progress.level}</p>
                  </div>
                </div>
              ) : (
                <div className="bg-orange-50 rounded-2xl p-6 border border-orange-100 shadow-sm">
                  <p className="text-lg font-semibold text-orange-900 mb-2">¡Todavía no has empezado!</p>
                  <p className="text-stone-600">Empieza a estudiar para aumentar tu progreso</p>
                </div>
              )}
            </div>

            {/* Divider */}
            <div className="border-t border-stone-200 my-8"></div>
            <div className="space-y-4">
              <h3 className="text-xl font-black text-stone-800 mb-4">Información de Contacto</h3>

              {/* Email */}
              <div className="flex items-center gap-4 p-4 bg-stone-50 rounded-2xl border border-stone-200">
                <div className="p-3 bg-white rounded-lg">
                  <Mail className="text-stone-600" size={24} />
                </div>
                <div>
                  <p className="text-xs font-bold uppercase text-gray-500 tracking-tight">Email</p>
                  <p className="text-lg font-semibold text-stone-800">{profile.email}</p>
                </div>
              </div>

              {/* Username */}
              <div className="flex items-center gap-4 p-4 bg-stone-50 rounded-2xl border border-stone-200">
                <div className="p-3 bg-white rounded-lg">
                  <User className="text-stone-600" size={24} />
                </div>
                <div>
                  <p className="text-xs font-bold uppercase text-gray-500 tracking-tight">Usuario</p>
                  <p className="text-lg font-semibold text-stone-800">@{profile.username}</p>
                </div>
              </div>
            </div>

            {/* Divider */}
            <div className="border-t border-stone-200 my-8"></div>

            {/* Metadata */}
            <div className="grid grid-cols-2 gap-4 p-4 bg-stone-50 rounded-2xl">
              <div>
                <p className="text-xs font-bold uppercase text-gray-500 tracking-tight mb-1">Último actualizado</p>
                <p className="text-sm font-semibold text-stone-700">{formatDate(profile.updated_at)}</p>
              </div>
              <div>
                <p className="text-xs font-bold uppercase text-gray-500 tracking-tight mb-1">ID de Usuario</p>
                <p className="text-xs font-mono text-stone-600 break-all">{profile.id.slice(0, 12)}...</p>
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
            Volver al Inicio
          </button>
          <button
            onClick={() => alert('Editar perfil - Próximamente')}
            className="flex-1 bg-orange-600 text-white px-6 py-4 rounded-2xl font-bold hover:bg-orange-700 transition-colors"
          >
            Editar Perfil
          </button>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
