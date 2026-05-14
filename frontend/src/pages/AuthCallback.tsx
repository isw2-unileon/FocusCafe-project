import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { syncGoogleUser } from '../services/auth_service';
import { useAuth } from '@/context/AuthContext';

export default function AuthCallback() {
  const navigate = useNavigate();
  const { handleOAuthToken } = useAuth();

  useEffect(() => {
    const sync = async () => {
      const token = await syncGoogleUser();
      if (token) {
        await handleOAuthToken(token);
        navigate('/home');
      } else {
        navigate('/login');
      }
    };
    sync();
  }, [navigate, handleOAuthToken]);

  
  return (
    <div className="min-h-screen bg-orange-50 flex flex-col items-center justify-center">
      <div className="flex flex-col items-center gap-4">
        <div className="w-12 h-12 border-4 border-orange-200 border-t-orange-600 rounded-full animate-spin"></div>
        <p className="text-orange-900 font-bold text-xl animate-pulse">
          Preparing your Café...
        </p>
      </div>                                             
    </div>
  );
}