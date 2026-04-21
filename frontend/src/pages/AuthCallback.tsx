import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { syncGoogleUser } from '../services/auth_service';

export default function AuthCallback() {
  const navigate = useNavigate();

  useEffect(() => {
    const sync = async () => {
      const ok = await syncGoogleUser();
      navigate(ok ? '/home' : '/login');
    };
    sync();
  }, [navigate]);

  return <p>Loading...</p>;
}