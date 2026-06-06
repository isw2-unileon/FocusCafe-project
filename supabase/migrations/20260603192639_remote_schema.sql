-- =====================================================================
-- FOCUSCAFE - DATABASE SCHEMA & RLS POLICIES (UNIVERSAL LOCAL & CI/CD)
-- =====================================================================

-- =====================================================================
-- 1. ENABLING ROW LEVEL SECURITY
-- =====================================================================
alter table "public"."cafe_orders" enable row level security;
alter table "public"."groups" enable row level security;
alter table "public"."questions" enable row level security;
alter table "public"."quizzes" enable row level security;
alter table "public"."study_materials" enable row level security;
alter table "public"."study_sessions" enable row level security;
alter table "public"."user_orders" enable row level security;
alter table "public"."user_progress" enable row level security;
alter table "public"."users" enable row level security;

-- =====================================================================
-- 2. FOREIGN KEYS & CONSTRAINTS
-- =====================================================================
alter table "public"."questions" add constraint "questions_quiz_id_fkey1" FOREIGN KEY (quiz_id) REFERENCES public.quizzes(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;
alter table "public"."questions" validate constraint "questions_quiz_id_fkey1";

alter table "public"."quizzes" add constraint "quizzes_session_id_fkey1" FOREIGN KEY (session_id) REFERENCES public.study_sessions(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;
alter table "public"."quizzes" validate constraint "quizzes_session_id_fkey1";

alter table "public"."study_materials" add constraint "study_materials_user_id_fkey1" FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;
alter table "public"."study_materials" validate constraint "study_materials_user_id_fkey1";

alter table "public"."study_sessions" add constraint "study_sessions_material_id_fkey1" FOREIGN KEY (material_id) REFERENCES public.study_materials(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;
alter table "public"."study_sessions" validate constraint "study_sessions_material_id_fkey1";

alter table "public"."study_sessions" add constraint "study_sessions_user_id_fkey1" FOREIGN KEY (user_id) REFERENCES public.users(id) ON UPDATE CASCADE ON DELETE CASCADE not valid;
alter table "public"."study_sessions" validate constraint "study_sessions_user_id_fkey1";

-- -- =====================================================================
-- -- 3. RLS POLICIES: CAFE ORDERS & CATALOGS
-- -- =====================================================================
create policy "All users see the catalog"
on "public"."cafe_orders"
as permissive
for select
to authenticated
using (true);

create policy "Admins can manage the catalog"
on "public"."cafe_orders"
as permissive
for all
to authenticated
using (((auth.jwt() ->> 'role'::text) = 'admin'::text));

-- =====================================================================
-- 4. RLS POLICIES: GROUPS
-- =====================================================================
create policy "Users can view their own groups or as admin"
on "public"."groups"
as permissive
for select
to authenticated
using (((id IN ( SELECT users.group_id
    FROM public.users
   WHERE (users.id = auth.uid()))) OR (EXISTS ( SELECT 1
    FROM public.users
   WHERE ((users.id = auth.uid()) AND (users.role = 'admin'::text))))));

create policy "Users can create groups"
on "public"."groups"
as permissive
for insert
to authenticated
with check (true);

create policy "Leaders can update their groups"
on "public"."groups"
as permissive
for update
to authenticated
using ((leader_id = auth.uid()))
with check ((leader_id = auth.uid()));

create policy "Leaders or admins can delete groups"
on "public"."groups"
as permissive
for delete
to authenticated
using (((leader_id = auth.uid()) OR (EXISTS ( SELECT 1
    FROM public.users
   WHERE ((users.id = auth.uid()) AND (users.role = 'admin'::text))))));

-- =====================================================================
-- 5. RLS POLICIES: PUBLIC READ TABLES (QUESTIONS & QUIZZES)
-- =====================================================================
create policy "Allow public read access to questions"
on "public"."questions"
as permissive
for select
to public
using (true);

create policy "System can manage questions"
on "public"."questions"
as permissive
for all
to service_role, postgres
using (true)
with check (true);

create policy "Allow public read access to quizzes"
on "public"."quizzes"
as permissive
for select
to public
using (true);

create policy "System can manage quizzes"
on "public"."quizzes"
as permissive
for all
to service_role, postgres
using (true)
with check (true);

-- =====================================================================
-- 6. RLS POLICIES: STUDY MATERIALS & SESSIONS
-- =====================================================================
create policy "Users can manage their own study materials"
on "public"."study_materials"
as permissive
for all
to authenticated
using ((auth.uid() = user_id))
with check ((auth.uid() = user_id));

create policy "Users can manage their own study sessions"
on "public"."study_sessions"
as permissive
for all
to authenticated
using ((auth.uid() = user_id))
with check ((auth.uid() = user_id));

-- =====================================================================
-- 7. RLS POLICIES: USER ORDERS
-- =====================================================================
create policy "Users can view their own orders or group orders"
on "public"."user_orders"
as permissive
for select
to authenticated
using ((auth.uid() = user_id) OR (group_id IN ( SELECT users.group_id
    FROM public.users
   WHERE (users.id = auth.uid()))));

create policy "Users can update their own orders or group orders"
on "public"."user_orders"
as permissive
for update
to authenticated
using ((auth.uid() = user_id) OR (group_id IN ( SELECT users.group_id
    FROM public.users
   WHERE (users.id = auth.uid()))))
with check ((auth.uid() = user_id) OR (group_id IN ( SELECT users.group_id
    FROM public.users
   WHERE (users.id = auth.uid()))));

create policy "System can manage user orders"
on "public"."user_orders"
as permissive
for all
to service_role, postgres
using (true)
with check (true);

-- =====================================================================
-- 8. RLS POLICIES: USER PROGRESS
-- =====================================================================
-- Protected Read: Prevents users from reading other students' progress directly via the public API.
create policy "Users can read their own progress"
on "public"."user_progress"
as permissive
for select
to authenticated
using ((auth.uid() = user_id));

create policy "Users can insert their initial progress"
on "public"."user_progress"
as permissive
for insert
to public
with check (true);

create policy "Users can update their own progress"
on "public"."user_progress"
as permissive
for update
to authenticated
using ((auth.uid() = user_id))
with check ((auth.uid() = user_id));

-- -- System: It allows administrative operations
create policy "System can manage user progress"
on "public"."user_progress"
as permissive
for all
to service_role, postgres
using (true)
with check (true);

-- =====================================================================
-- 9. RLS POLICIES: USERS / PROFILES (CORREGIDO EL BUCLE INFINITO)
-- =====================================================================
create policy "Admins can view all profiles"
on "public"."users"
as permissive
for select
to authenticated
using (((auth.jwt() ->> 'role'::text) = 'admin'::text));

create policy "Users can view own profile"
on "public"."users"
as permissive
for select
to authenticated
using ((auth.uid() = id));

create policy "Users can update their own profile"
on "public"."users"
as permissive
for update
to authenticated
using ((auth.uid() = id))
with check ((auth.uid() = id));

create policy "System can manage user profiles"
on "public"."users"
as permissive
for all
to service_role, postgres
using (true)
with check (true);

create policy "Users can insert their own profile on register"
on "public"."users" 
as permissive 
for insert 
to anon, authenticated
with check ((auth.uid() = id OR (auth.uid() IS NULL))); 

create policy "Admins can insert new profiles"
on "public"."users" 
as permissive 
for insert 
to authenticated
with check (((auth.jwt() ->> 'role'::text) = 'admin'::text));