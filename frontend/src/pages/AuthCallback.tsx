import { useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { syncGoogleUser, supabase } from '../services/auth_service';
import { useAuth } from '@/context/AuthContext';

export default function AuthCallback() {
  const navigate = useNavigate();
  const { handleOAuthToken } = useAuth();
  const called = useRef(false);

  useEffect(() => {
    if (called.current) return;
    called.current = true;

    const sync = async () => {
      const { token, synced } = await syncGoogleUser();
      
      // Obtener el intent (login o register) de localStorage
      const intent = localStorage.getItem('oauth_intent');
      console.log('DEBUG OAuth:', { intent, synced, hasToken: !!token });
      localStorage.removeItem('oauth_intent');

      if (token) {
        // Si el usuario intentaba registrarse pero ya existía en la base de datos
        if (intent === 'register' && !synced) {
          console.log('Duplicate detected, aborting login...');
          await supabase.auth.signOut();
          localStorage.removeItem('token'); // Aseguramos que no quede rastro del token
          navigate('/register?error=already_exists', { replace: true });
          return;
        }

        await handleOAuthToken(token);
        navigate('/home', { replace: true });
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