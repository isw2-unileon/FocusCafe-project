# FocusCafe - Local Development Seed Script
# Creates test users in auth.users + public.users + user_progress + user_orders
# Prerequisites: supabase start must be running

param(
    [string]$ApiUrl = "http://127.0.0.1:54321",
    [string]$ServiceKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6InNlcnZpY2Vfcm9sZSIsImV4cCI6MTk4MzgxMjk5Nn0.EGIM96RAZx35lJzdJsyH-qQwv8Hdp7fsn3W0YpN81IU",
    [string]$AnonKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJzdXBhYmFzZS1kZW1vIiwicm9sZSI6ImFub24iLCJleHAiOjE5ODM4MTI5OTZ9.CRXP1A7WOeoJeXxjNni43kdQwgnWNReilDMblYTn_I0"
)

$ErrorActionPreference = "Stop"

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  FocusCafe - Seeding Local Database" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# --- Helper: Create auth user in GoTrue ---
function Create-AuthUser {
    param(
        [string]$Email,
        [string]$Password
    )

    $body = @{
        email = $Email
        password = $Password
        email_confirm = $true
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$ApiUrl/auth/v1/admin/users" `
            -Method POST `
            -Headers @{
                "Authorization" = "Bearer $ServiceKey"
                "apikey" = $ServiceKey
                "Content-Type" = "application/json"
            } `
            -Body $body

        Write-Host "  [OK] Auth user created: $Email" -ForegroundColor Green
        return $response.id
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 409 -or $statusCode -eq 422) {
            Write-Host "  [SKIP] Auth user already exists: $Email" -ForegroundColor Yellow
            # Fetch existing user ID
            try {
                $existing = Invoke-RestMethod -Uri "$ApiUrl/auth/v1/admin/users?email=$Email" `
                    -Method GET `
                    -Headers @{
                        "Authorization" = "Bearer $ServiceKey"
                        "apikey" = $ServiceKey
                    }
                return $existing.users[0].id
            }
            catch {
                Write-Host "  [ERROR] Could not fetch existing user: $Email" -ForegroundColor Red
                return $null
            }
        }
        else {
            Write-Host "  [ERROR] Failed to create auth user $Email : $_" -ForegroundColor Red
            return $null
        }
    }
}

# --- Helper: Insert into public.users via PostgREST ---
function Insert-UserProfile {
    param(
        [string]$UserId,
        [string]$FirstName,
        [string]$LastName,
        [string]$Username,
        [string]$Email,
        [string]$Role
    )

    $body = @{
        id = $UserId
        first_name = $FirstName
        last_name = $LastName
        username = $Username
        email = $Email
        role = $Role
    } | ConvertTo-Json

    try {
        Invoke-RestMethod -Uri "$ApiUrl/rest/v1/users" `
            -Method POST `
            -Headers @{
                "Authorization" = "Bearer $ServiceKey"
                "apikey" = $ServiceKey
                "Content-Type" = "application/json"
                "Prefer" = "return=minimal"
            } `
            -Body $body | Out-Null

        Write-Host "  [OK] Profile created: $FirstName $LastName ($Role)" -ForegroundColor Green
    }
    catch {
        $errorResponse = $_.Exception.Message
        if ($errorResponse -match "duplicate key" -or $errorResponse -match "already exists" -or $errorResponse -match "23505") {
            Write-Host "  [SKIP] Profile already exists: $FirstName $LastName" -ForegroundColor Yellow
        }
        else {
            Write-Host "  [ERROR] Failed to create profile $FirstName $LastName : $_" -ForegroundColor Red
        }
    }
}

# --- Helper: Insert user_progress via PostgREST ---
function Insert-UserProgress {
    param(
        [string]$UserId,
        [int]$Energy,
        [int]$Level,
        [int]$Xp
    )

    $body = @{
        user_id = $UserId
        energy = $Energy
        level = $Level
        xp = $Xp
    } | ConvertTo-Json

    try {
        Invoke-RestMethod -Uri "$ApiUrl/rest/v1/user_progress" `
            -Method POST `
            -Headers @{
                "Authorization" = "Bearer $ServiceKey"
                "apikey" = $ServiceKey
                "Content-Type" = "application/json"
                "Prefer" = "return=minimal"
            } `
            -Body $body | Out-Null

        Write-Host "  [OK] Progress created: Level $Level, XP $Xp, Energy $Energy" -ForegroundColor Green
    }
    catch {
        $errorResponse = $_.Exception.Message
        if ($errorResponse -match "duplicate key" -or $errorResponse -match "already exists" -or $errorResponse -match "23505") {
            Write-Host "  [SKIP] Progress already exists for user" -ForegroundColor Yellow
        }
        else {
            Write-Host "  [ERROR] Failed to create progress: $_" -ForegroundColor Red
        }
    }
}

# --- Helper: Insert user_order via PostgREST ---
function Insert-UserOrder {
    param(
        [string]$UserId,
        [int]$CafeOrderId,
        [string]$Status = "completed"
    )

    $body = @{
        user_id = $UserId
        cafe_order_id = $CafeOrderId
        status = $Status
    } | ConvertTo-Json

    try {
        Invoke-RestMethod -Uri "$ApiUrl/rest/v1/user_orders" `
            -Method POST `
            -Headers @{
                "Authorization" = "Bearer $ServiceKey"
                "apikey" = $ServiceKey
                "Content-Type" = "application/json"
                "Prefer" = "return=minimal"
            } `
            -Body $body | Out-Null
    }
    catch {
        # Silent fail for orders - they're not critical
    }
}

# ========================================
# TEST USERS DEFINITION
# ========================================

$testUsers = @(
    @{
        Email     = "admin@focuscafe.local"
        Password  = "admin123"
        FirstName = "Admin"
        LastName  = "User"
        Username  = "admin"
        Role      = "admin"
        Level     = 10
        Xp        = 5000
        Energy    = 500
        Orders    = @(50, 46, 36)  # Focus Cafe Special, Eggs Benedict, Club Sandwich
    },
    @{
        Email     = "user@focuscafe.local"
        Password  = "user123"
        FirstName = "Test"
        LastName  = "User"
        Username  = "testuser"
        Role      = "user"
        Level     = 5
        Xp        = 2500
        Energy    = 300
        Orders    = @(10, 15, 22)  # Cappuccino, Caffe Latte, Cappuccino & Toastie
    },
    @{
        Email     = "marina@focuscafe.local"
        Password  = "test123"
        FirstName = "Marina"
        LastName  = "Garcia"
        Username  = "marina"
        Role      = "user"
        Level     = 3
        Xp        = 1200
        Energy    = 200
        Orders    = @(7, 19)      # Americano, Blueberry Scone
    },
    @{
        Email     = "carlos@focuscafe.local"
        Password  = "test123"
        FirstName = "Carlos"
        LastName  = "Lopez"
        Username  = "carlos"
        Role      = "user"
        Level     = 7
        Xp        = 4000
        Energy    = 450
        Orders    = @(23, 37, 40)  # Matcha Latte, Cold Brew & Avocado Toast, Turkish Coffee
    },
    @{
        Email     = "lucia@focuscafe.local"
        Password  = "test123"
        FirstName = "Lucia"
        LastName  = "Martinez"
        Username  = "lucia"
        Role      = "user"
        Level     = 1
        Xp        = 100
        Energy    = 50
        Orders    = @(5)           # Espresso
    },
    @{
        Email     = "pablo@focuscafe.local"
        Password  = "test123"
        FirstName = "Pablo"
        LastName  = "Ruiz"
        Username  = "pablo"
        Role      = "user"
        Level     = 2
        Xp        = 500
        Energy    = 150
        Orders    = @(9, 14)       # Macchiato, Cookie Duo
    }
)

# ========================================
# EXECUTE SEED
# ========================================

Write-Host "Creating test users..." -ForegroundColor Cyan
Write-Host ""

foreach ($user in $testUsers) {
    Write-Host ">>> $($user.Email)" -ForegroundColor White

    # 1. Create auth user
    $userId = Create-AuthUser -Email $user.Email -Password $user.Password

    if ([string]::IsNullOrEmpty($userId)) {
        Write-Host "  [SKIP] Skipping profile creation - no user ID" -ForegroundColor Yellow
        Write-Host ""
        continue
    }

    # 2. Create profile in public.users
    Insert-UserProfile -UserId $userId `
        -FirstName $user.FirstName `
        -LastName $user.LastName `
        -Username $user.Username `
        -Email $user.Email `
        -Role $user.Role

    # 3. Create user progress
    Insert-UserProgress -UserId $userId `
        -Energy $user.Energy `
        -Level $user.Level `
        -Xp $user.Xp

    # 4. Create user orders
    foreach ($orderId in $user.Orders) {
        Insert-UserOrder -UserId $userId -CafeOrderId $orderId -Status "completed"
    }
    Write-Host "  [OK] $($user.Orders.Count) orders created" -ForegroundColor Green

    Write-Host ""
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Seed completed!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test accounts:" -ForegroundColor White
Write-Host "  Admin:  admin@focuscafe.local / admin123" -ForegroundColor Yellow
Write-Host "  User:   user@focuscafe.local  / user123" -ForegroundColor Yellow
Write-Host "  Others: *_@focuscafe.local    / test123" -ForegroundColor Yellow
Write-Host ""
Write-Host "Open Studio at: http://127.0.0.1:54323" -ForegroundColor Cyan