# Medicine Inventory — Backend

Go API สำหรับระบบบริหารคลังยา — REST/JSON, JWT auth, RBAC (Admin/Staff),
รองรับการติดตามล็อต/วันหมดอายุ, FEFO, แจ้งเตือน LINE และรายงาน Excel

Frontend repo: https://github.com/xenonkup/medicine-inventory-frontend

## Stack
Go · Gin · GORM · JWT · Swagger · PostgreSQL (Supabase) · Deploy: Render

## Architecture (layered)
```
cmd/api            # entrypoint + dependency wiring
internal/
  config           # env loading
  domain           # entities, enums, errors
  dto              # request/response shapes
  handler          # HTTP layer (Gin)
  middleware       # JWT auth, RBAC, CORS
  service          # business logic (FEFO, alerts, reports)
  repository       # data access (GORM)
  database         # connection + migrations
pkg/               # jwt, hash, response helpers
```
ทิศทาง dependency: `handler → service → repository → domain`

## Run locally
```bash
cp .env.example .env     # ตั้ง DATABASE_URL (Supabase) + JWT_SECRET
go run ./cmd/api         # http://localhost:8080/api/v1/health
```
รันครั้งแรกจะสร้าง Admin เริ่มต้นจากค่า `BOOTSTRAP_ADMIN_*` (ค่าเริ่มต้น `admin` / `admin1234`)

## API (ปัจจุบัน)
- `POST /api/v1/auth/login` · `/auth/refresh` · `/auth/logout` · `GET /auth/me`
- `/api/v1/users` — CRUD (Admin)
- `/api/v1/categories` — read: ผู้ใช้ที่ login, write: Admin
- `/api/v1/medicines` — read: ผู้ใช้ที่ login, write: Admin; `GET /medicines/:id/lots`
- `/api/v1/stock/in` · `/stock/out` (FEFO) · `/stock/return` · `GET /stock/transactions` (Admin + Staff)

## Roadmap
- [x] Phase 0 — Setup
- [x] Phase 1 — Auth & Users
- [x] Phase 2 — Master data (Category, Medicine)
- [x] Phase 3 — Inventory core (Lot, Stock In/Out **FEFO**, Return; ledger; FEFO unit tests)
- [ ] Phase 4 — Dashboard & LINE · Phase 5 — Reports/Excel · Phase 6 — Deploy
