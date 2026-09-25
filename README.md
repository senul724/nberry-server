# Notify Berry Server 🍓

Backend server for the **Notify Berry** notification platform.

---

## 📖 Concept Overview

**Notify Berry** solves a fundamental challenge for businesses and creators: sending push notifications to customers usually requires building, deploying, and maintaining dedicated iOS and Android mobile apps in the App Store and Google Play Store.

Notify Berry eliminates this hurdle by providing a **shared notification ecosystem**:

```
┌─────────────────────────────────┐           ┌─────────────────────────────────┐
│        Vendor Web Portal        │           │     Vendor Backend / Service    │
│  (Create Apps, View Analytics)  │           │   (Automated triggers via API)  │
└────────────────┬────────────────┘           └────────────────┬────────────────┘
                 │                                             │
                 │ Web Auth (JWT / Session)                    │ Secret Key (SHA-256)
                 ▼                                             ▼
┌───────────────────────────────────────────────────────────────────────────────┐
│                           Notify Berry Server (Go)                            │
│  • Unified User Accounts (Vendor & Customer share the same identity model)    │
│  • App Management & API Secret Generation (SHA-256 hashed)                    │
│  • Broadcast Subscriptions (Composite Key: app_id + user_id)                  │
│  • Unicast Subscriptions (Secure Hash + Vendor Webhook Callback)              │
│  • Broadcast & Direct Memos Dispatch                                          │
└───────────────────────────────────────┬───────────────────────────────────────┘
                                        │ Push Notifications (FCM / APNs)
                                        ▼
                        ┌───────────────────────────────┐
                        │      Customer Mobile App      │
                        │ • Discover & subscribe to apps│
                        │ • Receive push notifications  │
                        │ • In-app message inbox        │
                        └───────────────────────────────┘
```

1. **Customers** download the single Notify Berry mobile app, browse or scan to subscribe to vendor apps/channels, and receive instant push notifications.
2. **Vendors** use the web portal to register an app, configure an optional `unicast_callback_url`, receive an API secret, and broadcast notifications directly to subscribers.
3. **Unified User Architecture**: In Notify Berry, both customers and vendors are stored as **`User`** records. There is no rigid role divergence in the backend schema—only the context and client platform (mobile app vs. web portal) changes.

---

## 🗄️ Database Schema & Models

### 1. `User` (`users`)
Both vendors and customers share this unified model.
* `id` (`UUID`, PK)
* `name` (`string`)
* `email` (`string`, unique)
* `password` (`string`, hashed, nullable)
* `photo_url` (`string`, nullable)
* `notify_token` (`string`, nullable) — Device push token (e.g., FCM / APNs) for delivering customer notifications.
* `created_at`, `updated_at`
* **Associations**:
  * `Apps`: Owned apps created by this user.
  * `BroadcastSubs`: Broadcast channels subscribed to by this user.
  * `UnicastSubs`: Unicast subscriptions for targeted vendor notifications.
  * `DirectMemos`: Direct memos targeted to this user.
  * `Sessions`: Active refresh sessions across devices.

### 2. `App` (`apps`)
Represents a vendor's application/channel that customers can subscribe to.
* `id` (`UUID`, PK) — Unique Application ID (`appID`).
* `user_id` (`UUID`, FK `users.id`, `ON DELETE CASCADE`) — The owner of the app.
* `name` (`string`) — Display name of the application.
* `logo` (`string`, nullable) — URL or path to the app icon/logo.
* `description` (`string`) — Overview of what notifications this app sends.
* `secret_hash` (`string`) — SHA-256 hash of the vendor API secret key. The raw secret key is only revealed once upon creation.
* `unicast_callback_url` (`string`, nullable) — Webhook endpoint called when a customer subscribes to unicast notifications.
* `created_at`, `updated_at`

### 3. `BroadcastSub` (`broadcast_subs`)
Links customers to an app's public broadcast feed.
* `app_id` (`UUID`, PK / FK `apps.id`, `ON DELETE CASCADE`)
* `user_id` (`UUID`, PK / FK `users.id`, `ON DELETE CASCADE`)
* `created_at`, `updated_at`
* **Composite Primary Key**: `(app_id, user_id)`.
* **Cascades**: Deleting either the user or the app automatically cascades and removes all associated broadcast subscriptions.

### 4. `BroadcastMemo` (`broadcast_memos`)
Broadcast notifications sent to all broadcast subscribers of an app.
* `id` (`UUID`, PK)
* `app_id` (`UUID`, FK `apps.id`, `ON DELETE CASCADE`)
* `title` (`string`)
* `content` (`string`)
* `image` (`string`, nullable) — Optional image attachment.
* `created_at`, `updated_at`

### 5. `UnicastSub` (`unicast_subs`)
Links a customer to a vendor app for direct 1-to-1 notifications.
* `id` (`UUID`, PK)
* `app_id` (`UUID`, FK `apps.id`, `ON DELETE CASCADE`)
* `user_id` (`UUID`, FK `users.id`, `ON DELETE CASCADE`)
* `unicast_key_hash` (`string`) — SHA-256 hash of the secret key generated for this unicast channel.
* `created_at`, `updated_at`
* **Webhook Lifecycle**: When a user subscribes to unicast, the server creates a unique `unicast_key`, hashes and stores it, then sends an HTTP POST request to the app's `unicast_callback_url` with `{ "unicast_key": "...", "vendor_user_id": "..." }`.

### 6. `DirectMemo` (`direct_memos`)
Direct, targeted notifications sent to an individual customer.
* `id` (`UUID`, PK)
* `app_id` (`UUID`, FK `apps.id`, `ON DELETE CASCADE`)
* `user_id` (`UUID`, FK `users.id`, `ON DELETE CASCADE`)
* `title` (`string`)
* `content` (`string`)
* `image` (`string`, nullable) — Optional image attachment.
* `created_at`, `updated_at`

---

## 🔑 App API Secret Security

When an app is created via `POST /api/apps`:
1. The server generates a cryptographically secure random secret key with prefix: `nb_sec_<64 hex characters>`.
2. The server computes the **SHA-256** checksum of the secret key.
3. Only the **SHA-256 hash** (`secret_hash`) is persisted in the PostgreSQL database.
4. The plain secret key is returned in the API response **once**. Subsequent requests cannot retrieve the plain secret key.

---

## 📡 Unicast Subscription Flow

```
Mobile App (Customer)              Notify Berry Server              Vendor Backend
        │                                  │                              │
        │── POST /subscriptions/unicast ──►│                              │
        │   { app_id, vendor_user_id }     │                              │
        │                                  │── Generate unicast_key       │
        │                                  │── Store SHA-256 key hash     │
        │                                  │                              │
        │                                  │── POST unicast_callback_url ─►
        │                                  │   { unicast_key,             │
        │                                  │     vendor_user_id }         │
        │                                  │◄───────── 200 OK ────────────│
        │◄── 201 Created (with key) ───────│                              │
```

---

## 🚀 API Endpoints Overview

### Authentication (`/api/auth`)
* `POST /api/auth/register-otp` — Register with Email OTP
* `POST /api/auth/login-otp` — Login with Email OTP
* `POST /api/auth/login-password` — Password Login
* `POST /api/auth/refresh` — Refresh access token
* `POST /api/auth/logout` — Revoke session
* `GET  /api/auth/google/login` — Google OAuth initiation
* `POST /api/auth/google/callback` — Google OAuth verification

### Apps (`/api/apps`) *(Requires Bearer Token)*
* `POST   /api/apps` — Create a new app (includes `unicast_callback_url`). Generates and returns the one-time `secret_key` and stores `secret_hash`.
* `GET    /api/apps` — List all apps owned by the authenticated user.
* `GET    /api/apps/:id` — Get details of a specific app owned by the user.
* `DELETE /api/apps/:id` — Delete an app (cascades to all subscriptions and memos).

### Subscriptions (`/api/subscriptions`) *(Requires Bearer Token)*
* `POST   /api/subscriptions/broadcast` — Subscribe authenticated user to an app's broadcast (`{ "app_id": "..." }`).
* `DELETE /api/subscriptions/broadcast/:app_id` — Unsubscribe authenticated user from broadcast.
* `POST   /api/subscriptions/unicast` — Subscribe authenticated user to unicast (`{ "app_id": "...", "vendor_user_id": "..." }`). Generates secret key, stores SHA-256 hash, and calls the app's `unicast_callback_url`.
* `DELETE /api/subscriptions/unicast/:app_id` — Unsubscribe authenticated user from unicast.

### Vendor Push (`/api/vendor` or `/api/push`) *(Requires `X-API-Key` or `Authorization: Bearer <app_secret>`)*
* `POST   /api/vendor/broadcast` — Push a broadcast memo to all app subscribers (`{ "title": "...", "content": "...", "image": "..." }`).
* `POST   /api/vendor/unicast` — Push a direct memo to an individual subscriber via `unicast_key` or `user_id` (`{ "unicast_key": "...", "title": "...", "content": "...", "image": "..." }`).

---

## 🛠️ Tech Stack

* **Language**: Go 1.26+
* **HTTP Framework**: [Gin Web Framework](https://github.com/gin-gonic/gin)
* **ORM**: [GORM](https://gorm.io/) with PostgreSQL Driver
* **Caching & Revocation**: [go-redis](https://github.com/redis/go-redis)
* **Authentication**: JWT (Access Token), Redis (JTIs), Database Sessions (Refresh Token)

---

## 🏃 Running Locally

### 1. Environment Variables
Create a `.env` file in the root directory:
```env
PORT=3030
DATABASE_URL="postgres://user:password@localhost:5432/nberry?sslmode=disable"
REDIS_URL="localhost:6379"
REDIS_PASSWORD=""
JWT_SECRET="your-jwt-secret"
REFRESH_SECRET="your-refresh-secret"
SESSION_SECRET="your-session-secret"
RESEND_API_KEY="your-resend-key"
```

### 2. Run Database Migrations & Server
```bash
go run ./cmd/server
```
Server runs on `http://localhost:3030`. Swagger docs are available at `http://localhost:3030/swagger/index.html`.
