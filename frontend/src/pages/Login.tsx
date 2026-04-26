import { useState, useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { Coffee, LogIn } from 'lucide-react';
import { useAuth } from '@/context/AuthContext';

export default function Login() {
  const [email, setEmail] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const { login, loginWithGoogle, error, isAuthenticated } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (isAuthenticated) {
      navigate('/home');
    }
  }, [isAuthenticated, navigate]);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await login(email, password);
    } catch (_err) {
      // Error is handled by AuthContext
    }
  };

  return (
    <div className="min-h-screen bg-orange-50 flex items-center justify-center p-4">
      <div className="max-w-md w-full bg-white rounded-[2.5rem] shadow-2xl p-10 border-b-8 border-orange-200">

        <div className="text-center mb-8">
          <div className="bg-orange-100 w-20 h-20 rounded-full flex items-center justify-center mx-auto mb-4">
            <Coffee size={40} className="text-orange-600" />
          </div>
          <h1 className="text-4xl font-black text-orange-900 tracking-tight">FocusCafe</h1>
          <p className="text-orange-600 font-medium mt-1">Where studying pays off</p>
        </div>

        <form onSubmit={handleLogin} className="space-y-4">
          {error && (
            <p className="text-red-500 text-sm text-center font-medium">{error}</p>
          )}
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={e => setEmail(e.target.value)}
            className="w-full p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all"
          />
          <input
            type="password"
            placeholder="Password"
            value={password}
            onChange={e => setPassword(e.target.value)}
            className="w-full p-4 bg-gray-50 rounded-2xl border-2 border-transparent focus:border-orange-400 outline-none transition-all"
          />
          <button
            type="submit"
            className="w-full bg-orange-600 text-white font-bold py-4 rounded-2xl shadow-lg hover:bg-orange-700 transition-all flex items-center justify-center gap-2"
          >
            <LogIn size={20} />
            Enter the Café
          </button>

          <div className="flex items-center gap-3">
            <div className="flex-1 h-px bg-gray-200" />
            <span className="text-gray-400 text-sm">or continue with</span>
            <div className="flex-1 h-px bg-gray-200" />
          </div>

          <button
            type="button"
            onClick={loginWithGoogle}
            className="w-full flex items-center justify-center gap-3 py-4 rounded-2xl border-2 border-gray-200 bg-white hover:bg-gray-50 transition-all font-bold text-gray-700"
          >
            <svg width="20" height="20" viewBox="0 0 18 18">
              <path fill="#4285F4" d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844c-.209 1.125-.843 2.078-1.796 2.717v2.258h2.908c1.702-1.567 2.684-3.874 2.684-6.615Z"/>
              <path fill="#34A853" d="M9 18c2.43 0 4.467-.806 5.956-2.184l-2.908-2.258c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 0 0 9 18Z"/>
              <path fill="#FBBC05" d="M3.964 10.707A5.41 5.41 0 0 1 3.682 9c0-.593.102-1.17.282-1.707V4.961H.957A8.996 8.996 0 0 0 0 9c0 1.452.348 2.827.957 4.039l3.007-2.332Z"/>
              <path fill="#EA4335" d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 0 0 .957 4.961L3.964 7.293C4.672 5.166 6.656 3.58 9 3.58Z"/>
            </svg>
            Sign in with Google
          </button>

          <Link
            to="/register"
            className="block w-full text-center text-orange-800 text-sm font-bold opacity-60 hover:opacity-100 transition-opacity"
          >
            New here? Sign up
          </Link>
        </form>

      </div>
    </div>
  );
}