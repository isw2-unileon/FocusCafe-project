-- =====================================================================
-- FOCUSCAFE - DATABASE SCHEMA & RLS POLICIES (UNIVERSAL LOCAL & CI/CD)
-- =====================================================================

-- =====================================================================
-- 1. ENABLING ROW LEVEL SECURITY
-- =====================================================================
alter table "public"."cafe_orders" enable row level security;
alter table "public"."questions" enable row level security;
alter table "public"."quizzes" enable row level security;
alter table "public"."study_materials" enable row level security;
alter table "public"."study_sessions" enable row level security;
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

-- =====================================================================
-- 3. RLS POLICIES: CAFE ORDERS & CATALOGS
-- =====================================================================
create policy "All users see the catalog"
on "public"."cafe_orders"
as permissive
for select
to authenticated
using (true);

-- =====================================================================
-- 4. RLS POLICIES: GROUPS
-- =====================================================================
create policy "Permiso de borrado para grupos"
on "public"."groups"
as permissive
for delete
to authenticated
using (((leader_id = auth.uid()) OR (EXISTS ( SELECT 1
   FROM public.users
  WHERE ((users.id = auth.uid()) AND (users.role = 'admin'::text))))));

create policy "Permiso de lectura para grupos"
on "public"."groups"
as permissive
for select
to authenticated
using (((id IN ( SELECT users.group_id
   FROM public.users
  WHERE (users.id = auth.uid()))) OR (EXISTS ( SELECT 1
   FROM public.users
  WHERE ((users.id = auth.uid()) AND (users.role = 'admin'::text))))));

-- =====================================================================
-- 5. RLS POLICIES: PUBLIC READ TABLES
-- =====================================================================
create policy "Permitir lectura publica"
on "public"."questions"
as permissive
for select
to public
using (true);

create policy "Permitir lectura de cuestionarios"
on "public"."quizzes"
as permissive
for select
to public
using (true);

create policy "Permitir lectura de materiales"
on "public"."study_materials"
as permissive
for select
to public
using (true);

-- =====================================================================
-- 6. RLS POLICIES: STUDY SESSIONS
-- =====================================================================
create policy "Users can manage their own study sessions"
on "public"."study_sessions"
as permissive
for all
to authenticated
using ((auth.uid() = user_id))
with check ((auth.uid() = user_id));

-- =====================================================================
-- 7. RLS POLICIES: USER PROGRESS
-- =====================================================================
create policy "Usuarios pueden leer su propio progreso"
on "public"."user_progress"
as permissive
for select
to authenticated
using ((auth.uid() = user_id));

create policy "System can insert user progress"
on "public"."user_progress"
as permissive
for insert
to service_role, postgres
with check (true);

-- =====================================================================
-- 8. RLS POLICIES: USERS (PROFILES)
-- =====================================================================
create policy "Admins can view all profiles"
on "public"."users"
as permissive
for select
to authenticated
using ((EXISTS ( SELECT 1
   FROM public.users users_1
  WHERE ((users_1.id = auth.uid()) AND (users_1.role = 'admin'::text)))));

create policy "Users can view own profile"
on "public"."users"
as permissive
for select
to authenticated
using ((auth.uid() = id));

create policy "System can insert user profiles"
on "public"."users"
as permissive
for insert
to service_role, postgres
with check (true);