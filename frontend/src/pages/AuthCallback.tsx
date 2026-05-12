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

  return <p>Loading...</p>;
}