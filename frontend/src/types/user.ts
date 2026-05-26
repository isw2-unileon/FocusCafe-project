export interface Group {
    id: string;          // UUID
    name: string;
    invite_code: string;
    leader_id: string;
}

export interface UserStats {
    id: number;
    first_name: string;
    energy: number;
    max_energy: number;
    xp: number;
    level: number;
    
    group?: Group;
}