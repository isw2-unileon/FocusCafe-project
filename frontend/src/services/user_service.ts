import { UserStats } from "@/types/user";

export const getRemoteUserStats = async (userId: number): Promise<UserStats> => {
    const response = await fetch(`/api/user/${userId}/stats`);
    if(!response.ok){
        throw new Error("Failed to fetch user stats");
    }
    return response.json();
}