import { UserStats } from "@/types/user";
import { UserProfile } from "@/types/user-profile";
import { apiFetch } from "@/services/api_client";

//Prefix for all the routes
const PATH = "/users";

//Fetch remote user statistics
export async function getRemoteUserStats(): Promise<UserStats> {
    return apiFetch(`${PATH}/me`);
}

//Get the current user's profile details
export async function getCurrentProfile(): Promise<UserProfile> {
    return apiFetch(`${PATH}/me`);
}

//Update the user's profile information
export async function updateUserProfile(data: { first_name: string; last_name: string }): Promise<UserProfile> {
    return apiFetch(`${PATH}/me`, {
        method: 'PUT',
        body: JSON.stringify(data),
    });
}

// Fetch all users (admin only)
export async function getAllUsers(): Promise<UserProfile[]> {
    return apiFetch('/admin/users');
}

// Create a new user (admin only)
export async function createUser(data: {
    first_name: string;
    last_name: string;
    email: string;
    password: string;
    confirm_password: string;
    role: string;
}): Promise<void> {
    return apiFetch('/admin/users', {
        method: 'POST',
        body: JSON.stringify(data),
    });
}

// Delete a user by ID (admin only)
export async function deleteUser(id: string): Promise<void> {
    return apiFetch(`/admin/users/${id}`, {
        method: 'DELETE',
    });
}

// Fetch global leaderboard top 5
export async function getLeaderboard(): Promise<UserProfile[]> {
    return apiFetch('/leaderboard');
}

// Fetch current user's global rank and profile
export async function getLeaderboardMe(): Promise<{ rank: number; user: UserProfile }> {
    return apiFetch('/leaderboard/me');
}