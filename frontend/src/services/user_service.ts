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
    return apiFetch('/me', {
        method: 'PUT',
        body: JSON.stringify(data),
    });
}