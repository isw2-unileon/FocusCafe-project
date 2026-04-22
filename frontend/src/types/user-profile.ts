export interface UserProgress {
  user_id: string;
  energy: number;
  level: number;
}

export interface UserProfile {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  username: string;
  role: string;
  created_at: string;
  updated_at: string;
  progress?: UserProgress;
}
