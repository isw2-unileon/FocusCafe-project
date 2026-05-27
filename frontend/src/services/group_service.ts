import { apiFetch } from "@/services/api_client";
import { Group } from "@/types/user";

export interface GroupMember {
    id: string;
    first_name: string;
    last_name: string;
    email: string;
    level: number;
}

export interface GroupDetail {
    id: number;
    name: string;
    invite_code: string;
    leader_id: string;
    created_at: string;
    members: GroupMember[];
}

export async function createGroup(name: string): Promise<Group> {
    return apiFetch('/groups', {
        method: 'POST',
        body: JSON.stringify({ name }),
    });
}

export async function joinGroup(inviteCode: string): Promise<Group> {
    return apiFetch('/groups/join', {
        method: 'POST',
        body: JSON.stringify({ invite_code: inviteCode }),
    });
}

export async function leaveGroup(): Promise<{ message: string }> {
    return apiFetch('/groups/leave', {
        method: 'POST',
    });
}

export async function deleteGroup(): Promise<{ message: string }> {
    return apiFetch('/groups', {
        method: 'DELETE',
    });
}

export async function getAllGroups(): Promise<GroupDetail[]> {
    return apiFetch('/admin/groups');
}

export async function adminDeleteGroup(groupId: number): Promise<{ message: string }> {
    return apiFetch(`/admin/groups/${groupId}`, {
        method: 'DELETE',
    });
}
