# ADR-002: Lazy Order Generation (N=3)

## Status

Accepted

## Date

2025-05-15

## Context

The application needs to keep each user's order list continuously active to maintain the game loop, without overwhelming the database or the user interface. When a user completes all pending orders, the system must automatically replenish the list so the gamification cycle never breaks.

Assigning every unlocked cafe order at once (e.g. 50+ orders for a high-level user) would create an unmanageable UI list and pollute the database with rows the player cannot realistically process. Conversely, pre-generating all orders at account creation would quickly become stale as the user levels up.

## Decision

We will implement a **lazy filling** strategy limited to **3 orders at a time**:

1. Orders are generated on-demand inside `GetUserOrders` when the pending count for a user (personal) or group reaches zero.
2. The system queries the user's current level (or the highest level in the group), then selects **3 random `CafeOrder` rows** whose `required_level <= that level`.
3. These 3 rows are inserted into `user_orders` with `status = "pending"`.
4. The same mechanism applies to both **personal orders** (`group_id IS NULL`) and **group orders** (`group_id = <group_id>`).

This approach applies equally to both individual users and collaborative groups, ensuring that team members always see a fresh, level-appropriate set of shared orders.

## Consequences

### Positive

- **UX & Gamification**: A short list of 3 orders reduces cognitive load and creates a sense of "new customers arriving at the cafe". It incentivises the player to finish the current batch to see what comes next.
- **Database Efficiency**: We avoid creating hundreds of `pending` rows that the user never sees. `SELECT` queries on `user_orders` stay small and fast.
- **Relevance**: Because the level is checked at every refill, a player who jumps from level 1 to level 10 will immediately receive level-10 orders instead of being stuck with a backlog of low-level items.
- **Self-Healing**: If a user or group somehow ends up with zero orders, the next `GetUserOrders` call automatically fixes the situation without manual intervention.

### Negative

- **Extra Queries on Empty List**: When the list is exhausted, `GetUserOrders` performs an additional `SELECT` against `user_progress` and `cafe_orders` to determine the 3 replacements. The impact is negligible because it only happens when the list is already empty.
- **No Upfront Visibility**: Users cannot see all future orders; they only discover the next batch after completing the current one.

## Alternatives Considered

- **Pre-generate all unlocked orders at once**: Rejected because it creates an unmanageable UI and leaves stale low-level orders after rapid level-ups.
- **Generate orders on user creation only**: Rejected because the initial set would become obsolete as the user progresses.
- **Fixed daily quota (e.g. 10 orders/day)**: Rejected because it breaks the real-time game loop; players should be able to keep playing as long as they have energy.
