# ADR-005: WebSockets for Real-Time Collaborative Orders

## Status

Accepted

## Date

2026-06-03

## Context

FocusCafe features a **collaborative group mode** where multiple users join a team and share group orders. When one member completes a shared order, the order must disappear from all other members' screens immediately. Without a real-time channel, users would only see updates after manually refreshing the page, leading to:
- Race conditions where two members try to complete the same order simultaneously.
- Stale UI state where deleted groups or completed orders remain visible.
- A poor sense of teamwork because actions feel isolated rather than shared.

## Decision

We will use **WebSockets** to push real-time events from the backend to connected clients:

1. **Backend Hub**: A single `Hub` struct (`backend/internal/ws/hub.go`) maintains a registry of authenticated `Client` connections, each tagged with a `UserID` and optional `GroupID`.
2. **Broadcast Model**: The Hub exposes `BroadcastToGroup(groupID, message)` and `SendToUser(userID, message)` to target specific audiences.
3. **Event Types**:
   - `ORDERS_UPDATED`: Sent after a user completes any order. All group members receive it and refresh their order list.
   - `GROUP_DELETED`: Sent when a group leader (or admin) deletes the team. All connected members clear their local group state immediately.
4. **Frontend Subscription**: `WebSocketContext.tsx` wraps the native `WebSocket` in a React Context, offering a `subscribe(type, callback)` API that components (e.g. `OrderList.tsx`, `Home.tsx`) use to listen for specific events.

## Consequences

### Positive

- **Real-Time Collaboration**: Group members see order completions and deletions instantly, creating a shared gameplay experience.
- **Reduced Polling**: No need for aggressive HTTP polling; the server pushes updates only when they happen.
- **Race Condition Mitigation**: When `ORDERS_UPDATED` arrives, the frontend removes the completed order locally. This reduces (though does not eliminate) the chance of a second user clicking "Complete" on the same order.

### Negative

- **Connection Complexity**: WebSocket connections require authentication handshake, reconnection logic (3-second retry), and careful cleanup on unmount to avoid memory leaks.
- **Offline Members**: Users who are not connected at the moment of broadcast miss the event. They rely on the next `GetUserOrders` HTTP call to synchronise state.
- **Sticky GroupID**: The backend caches each client's `GroupID` at connection time. If a user joins or leaves a group without reconnecting, the server may broadcast to the wrong audience until the next WebSocket reconnect.

## Alternatives Considered

- **HTTP Long Polling**: Rejected because it is less efficient for frequent, small messages and creates unnecessary server load.
- **Server-Sent Events (SSE)**: Rejected because SSE is unidirectional (server → client). We wanted the flexibility of future bidirectional features (e.g. typing indicators or live leaderboards).
- **Supabase Realtime (Postgres changes)**: Rejected to avoid additional Supabase dependency lock-in and to keep full control over the event payload shape and filtering logic.
