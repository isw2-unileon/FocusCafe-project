import { UserStats } from "@/types/user";

export const getRemoteUserStats = async (userId: number): Promise<UserStats> => {
    if(!response.ok){
        throw new Error("Failed to fetch user stats");
    }
    return response.json();
    /*return {
        id: userId,
        name: "Pepe",
        energy: 150,
        max_energy: 500,
        level: 3
    };*/
}