# Data Models and Database Schema

The persistence layer of FocusCafe relies on a relational SQL model managed through GORM (Object-Relational Mapping). This structure enforces referential integrity, strong data typing, and strict relational constraints across all core features.

Below is the Entity-Relationship (ER) representation highlighting how the internal authentication records link directly to our custom application tables (`users`, `groups`, `study_sessions`)

## Table `users`

| Name | Type | Constraints |
|------|------|-------------|
| `first_name` | `text` |  |
| `last_name` | `text` |  |
| `username` | `text` |  |
| `email` | `text` |  Unique |
| `role` | `text` |  Nullable |
| `id` | `uuid` | Primary |
| `created_at` | `timestamptz` |  Nullable |
| `group_id` | `int8` |  Nullable |

## Table `user_progress`

| Name | Type | Constraints |
|------|------|-------------|
| `user_id` | `uuid` | Primary |
| `energy` | `int4` |  Nullable |
| `level` | `int4` |  Nullable |
| `xp` | `int4` |  Nullable |

## Table `study_materials`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `user_id` | `uuid` |  |
| `title` | `text` |  |
| `subject_name` | `text` |  |
| `file_path` | `text` |  |
| `upload_date` | `timestamp` |  Nullable |
| `content` | `text` |  Nullable |

## Table `study_sessions`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `user_id` | `uuid` |  |
| `material_id` | `int8` |  |
| `duration_minutes` | `int8` |  |
| `start_time` | `timestamp` |  |
| `end_time` | `timestamp` |  Nullable |
| `status` | `text` |  |

## Table `cafe_orders`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `name` | `text` |  |
| `category` | `text` |  Nullable |
| `energy_cost` | `int8` |  |
| `reward_xp` | `int8` |  |
| `description` | `text` |  Nullable |
| `required_level` | `int8` |  Nullable |

## Table `quizzes`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `session_id` | `int8` |  |
| `generated_at` | `timestamp` |  Nullable |

## Table `questions`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `quiz_id` | `int8` |  |
| `question_text` | `text` |  |
| `option_a` | `text` |  |
| `option_b` | `text` |  |
| `option_c` | `text` |  |
| `option_d` | `text` |  |
| `correct_answer` | `bpchar` |  |
| `explanation` | `text` |  Nullable |

## Table `user_orders`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `cafe_order_id` | `int8` |  |
| `status` | `text` |  Nullable |
| `created_at` | `timestamp` |  Nullable |
| `user_id` | `uuid` |  Nullable |
| `group_id` | `int8` |  Nullable |

## Table `groups`

| Name | Type | Constraints |
|------|------|-------------|
| `id` | `int8` | Primary Identity |
| `created_at` | `timestamptz` |  |
| `name` | `text` |  |
| `invite_code` | `text` |  Unique |
| `leader_id` | `uuid` |  |
