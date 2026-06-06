# ADR-002: Lazy Order Generation with Frontend Limit (N=3)

## Status

Accepted

## Date

2026-05-15

## Context

The application needs to keep each user's order list continuously active to maintain the game loop, without overwhelming the user interface. When a user completes all pending orders, the system must automatically replenish the list so the gamification cycle never breaks.

Assigning every unlocked cafe order at once (e.g. 50+ orders for a high-level user) would create an unmanageable UI list. Conversely, pre-generating all orders at account creation would quickly become stale as the user levels up.

## Decision

We will implement a **two-part strategy** combining lazy backend filling with a frontend display limit:

### 1. Lazy Filling (Backend)

The backend generates orders **on-demand** inside `GetUserOrders` when the pending count reaches zero:

1. For **personal orders** (`group_id IS NULL`): the system queries the user's current level and generates **up to 3** random `CafeOrder` rows matching that level.
2. For **group orders** (`group_id = <group_id>`): the system queries the **highest level in the group** and generates **up to 3** random matching rows.
3. These rows are inserted into `user_orders` with `status = "pending"`.

The backend does **not** enforce a hard cap on how many total pending orders may exist in the database; its responsibility is only to ensure the list is never empty.

### 2. Display Limit (Frontend)

The `OrderList` component renders **at most 3 pending orders** using `orders.slice(0, 3)`. This keeps the UI compact and avoids cognitive overload, regardless of how many pending rows the database might hold.

This approach applies equally to both individual users and collaborative groups.

## Consequences

### Positive

- **UX & Gamification**: A short list of 3 orders reduces cognitive load and creates a sense of "new customers arriving at the cafe". It incentivises the player to finish the current batch to see what comes next.
- **Database Efficiency**: We avoid creating hundreds of `pending` rows that the user never sees. `SELECT` queries on `user_orders` stay small and fast because regeneration only happens when the list is exhausted.
- **Relevance**: Because the level is checked at every refill, a player who jumps from level 1 to level 10 will immediately receive level-10 orders instead of being stuck with a backlog of low-level items.
- **Self-Healing**: If a user or group somehow ends up with zero orders, the next `GetUserOrders` call automatically fixes the situation without manual intervention.

### Negative

- **Extra Queries on Empty List**: When the list is exhausted, `GetUserOrders` performs an additional `SELECT` against `user_progress` and `cafe_orders` to determine the replacements. The impact is negligible because it only happens when the list is already empty.
- **No Upfront Visibility**: Users cannot see all future orders; they only discover the next batch after completing the current one.

## Alternatives Considered

- **Pre-generate all unlocked orders at once**: Rejected because it creates an unmanageable UI and leaves stale low-level orders after rapid level-ups.
- **Generate orders on user creation only**: Rejected because the initial set would become obsolete as the user progresses.
- **Fixed daily quota (e.g. 10 orders/day)**: Rejected because it breaks the real-time game loop; players should be able to keep playing as long as they have energy.
- **Backend hard-caps the pending count**: Rejected because it adds unnecessary complexity; the frontend slice is sufficient to control UX while the backend remains focused on regeneration logic.
