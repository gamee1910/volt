# Tóm tắt dự án Volt

## Mục đích

Volt là backend Go dùng cho mục đích cá nhân:
1. Đăng nhập cổng khách hàng **EVNHCMC** và lấy sản lượng điện từng ngày.
2. Lưu vào **PostgreSQL**, mỗi ngày 1 dòng (upsert theo `measurement_date`).
3. Tính **tiền điện bậc thang** cộng VAT 8% cho tháng hiện tại.
4. Báo cáo qua **Telegram bot** và một **HTTP API** nhỏ.

Dự án không liên kết với EVNHCMC.

## Thành phần đang chạy (`cmd/api`)

Một process chạy song song 3 thứ, cùng dùng chung `di.Container`:

| Thành phần | Chi tiết |
|---|---|
| HTTP API (chi) | `GET /health`, `POST /api/v1/sync`, `GET /api/v1/`, `GET /api/v1/yesterday` |
| Telegram bot (long polling) | `/yesterday`, `/sync [from] [to]`, `/get`, `/help`; trả lời bằng tiếng Việt |
| Scheduler (robfig/cron) | `0 */12 * * *` giờ Việt Nam: sync từ đầu tháng đến hôm nay. Chạy thêm 1 lần khi khởi động. Sau mỗi lần chạy theo lịch, gửi báo cáo (hoặc thông báo lỗi) tới `TELEGRAM_CHAT_ID` |

CLI chạy một lần: `cmd/cli/sync` (sync tháng hiện tại) và `cmd/cli/report` (gửi báo cáo hôm qua). Hai lệnh này được gọi từ GitHub Actions trên self-hosted runner, hiện chỉ chạy tay (`workflow_dispatch`).

## Kiến trúc

Clean/hexagonal + CQRS:

```
cmd/            entrypoint: config → logger → DB → di.NewContainer
internal/
  domain/         entity + interface repository (không phụ thuộc gì bên ngoài)
  application/    Commands (SyncElectricity) / Queries (AllElectricity, YesterdayUsage)
                  mỗi handler được bọc decorator logging → metrics
  infrastructure/ postgres repository, di/container.go (nơi duy nhất nối mọi thứ)
  interfaces/     api (HTTP), bot (Telegram), scheduler (cron), chỉ gọi vào Application
  common/         decorator generic, metrics (chỉ log), utils (múi giờ, ngày EVN, tính tiền)
pkg/            evnhcmc client, telegram client, zap logger (không import internal)
migrations/     golang-migrate, 3 file tuần tự
```

## Nghiệp vụ chính

- Mọi thao tác ngày tháng dùng múi giờ `Asia/Ho_Chi_Minh`. Dữ liệu múi giờ được nhúng qua `time/tzdata`.
- Nếu EVN trả `TotalOutput = 0` thì dùng `TotalIndex` thay thế.
- Tiền tháng = `CalculateElectricityBill(tổng kWh từ ngày 1 đến hôm qua)`. Các bậc: 1893 / 1956 / 2271 / 2860 / 3197 / 3302 đ/kWh (50/50/100/100/100/trên 400 kWh), cộng VAT 8%.

## Cấu hình

Chỉ đọc từ biến môi trường (`config/config.go`). Khi dev local, direnv nạp từ `.envrc` (mẫu: `.envrc.example`).
Nhóm biến: `APP_*`, `EVN_*`, `TELEGRAM_API_KEY`, `TELEGRAM_CHAT_ID`, `DB_*`, `SERVER_*`.

## Hạ tầng

- **Local / VPS có Docker:** `docker-compose.yml` gồm `db` (postgres:16-alpine, giới hạn WAL 128 MB), `migrate` (tự chạy migration), `api`. Log giới hạn 10 MB × 3 file. Tổng image khoảng 540 MB.
- **Botkeep (không có Docker):** xem `.claude/deploy-checklist.md`.
- `Dockerfile`: build bằng golang:1.26-alpine, chạy trên alpine; image khoảng 35 MB.

## Trạng thái và việc còn tồn

- Chưa có test (`go test ./...` không chạy test nào).
- Migration chưa tự chạy khi app khởi động. Trên môi trường không có Docker phải chạy tay.
- README còn ghi sai `EVN_CUSTOMER_CODE` (đúng là `EVN_CUSTOMER`) và `.env.example` (đúng là `.envrc.example`).
- `scripts/db_init.sql` (`CREATE DATABASE volt`) không còn được docker-compose dùng.
- `persistences/redis` mới chỉ là thư mục trống, chưa có code.
