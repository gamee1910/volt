# Checklist deploy Volt lên Botkeep

Nền tảng: [botkeep.cloud](https://botkeep.cloud). Botkeep là panel chạy runtime có sẵn, **không có Docker và không có quyền root**, nên không dùng `docker-compose.yml` ở đây.
Gói Founder Free: 5 slot, 2 GB RAM, 1.5 vCore, 2 GB dung lượng (dùng chung cho mọi slot). Volt cần 2 slot: PostgreSQL + Go.

## 1. Chuẩn bị ở máy local

- [ ] Code đã build và vet sạch: `go build ./... && go vet ./...`
- [ ] Build sẵn file chạy cho Linux. Không build trên server vì runtime Go của Botkeep có thể cũ hơn 1.26 và build sẽ tốn RAM, dung lượng:
  ```bash
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o volt ./cmd/api
  ```
- [ ] Dữ liệu múi giờ đã nhúng sẵn qua `time/tzdata` trong `internal/common/utils/utils.go`, server không cần cài gì thêm.
- [ ] Có đủ thông tin: tài khoản EVN, các URL/path API EVN, Telegram bot token.
- [ ] Đã lấy chat ID: nhắn bot 1 tin, mở `https://api.telegram.org/bot<TOKEN>/getUpdates` và xem trường `chat.id`.

## 2. Slot PostgreSQL

- [ ] Tạo server PostgreSQL trên Botkeep.
- [ ] Ghi lại host, port, user, password, tên database và cài đặt TLS từ tab **Database**.
- [ ] Chạy migration (thư mục `migrations/`, 3 file). Hiện app **chưa** tự chạy migration, nên chọn một trong hai cách:
  - Nếu DB cho kết nối từ ngoài: chạy từ máy local
    `migrate -path ./migrations -database "postgres://USER:PASS@HOST:PORT/DB?sslmode=require" up`
  - Nếu không: chạy lần lượt nội dung các file `*.up.sql` theo thứ tự `000001` → `000003` bằng công cụ SQL mà Botkeep cung cấp.
- [ ] Kiểm tra bảng `electricity_consumption` đã có, gồm các cột `measurement_date` (unique) và `updated_at`.

## 3. Slot Go (Volt API)

- [ ] Tạo server với runtime **Go**.
- [ ] Upload file chạy `volt` qua ZIP hoặc Files. Kiểm tra file có quyền thực thi.
- [ ] Đặt lệnh khởi động: `./volt`
- [ ] Đặt biến môi trường trong mục **Environment**. Các thông tin bí mật lưu dạng secret:

  | Biến | Ghi chú |
  |---|---|
  | `APP_ENV` | `production` |
  | `APP_NAME` | `volt` |
  | `EVN_USERNAME`, `EVN_PASSWORD` | secret |
  | `EVN_CUSTOMER` | mã khách hàng (không phải `EVN_CUSTOMER_CODE`) |
  | `EVN_BASE_URL`, `EVN_LOGIN_API`, `EVN_ELECTRICITY_CONSUMPTION_API` | bắt buộc, thiếu là app dừng ngay khi khởi động |
  | `TELEGRAM_API_KEY` | secret |
  | `TELEGRAM_CHAT_ID` | chat nhận báo cáo tự động; để trống là tắt |
  | `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME` | lấy từ slot PostgreSQL |
  | `DB_SSL_MODE` | `require` nếu DB bật TLS, nếu không thì `disable` |
  | `SERVER_PORT` | **không đặt**, Botkeep tự cấp |

## 4. Khởi động và kiểm tra

- [ ] Start server và xem **Console**. Log khởi động đúng sẽ có lần lượt:
  - `telegram_bot_started`
  - `electricity_scheduler: started, schedule=0 */12 * * *`
  - `starting application [volt] port [...] env [production]`
- [ ] Không có các lỗi sau:
  - `database error`: sai thông tin DB hoặc `DB_SSL_MODE`
  - `error call getMe`: sai token, hoặc server không ra được internet
  - `load Vietnam timezone`: không còn xảy ra vì tzdata đã nhúng
  - `EVN_BASE_URL is required` (hoặc thông báo tương tự cho path API): thiếu biến EVN
- [ ] Lần sync đầu (chạy ngay khi khởi động) không báo lỗi `electricity_scheduler: initial run failed`. Lỗi này nghĩa là không đăng nhập hoặc không gọi được EVN.
- [ ] Nhắn bot `/help`, rồi `/yesterday` và `/get`: bot trả lời có dữ liệu.
- [ ] Nếu bật domain: `GET https://<domain>/health` trả về `OK`.
- [ ] Đến mốc 0h hoặc 12h giờ Việt Nam: `TELEGRAM_CHAT_ID` nhận báo cáo tự động.

## 5. Sau khi deploy

- [ ] Tạo backup cho slot PostgreSQL (dừng DB trước khi backup).
- [ ] Chỉ chạy **một** instance bot. Hai instance cùng long polling sẽ tranh nhau tin nhắn, và report CLI dùng `getUpdates` sẽ lỗi.
- [ ] Nếu đã chạy trên Botkeep thì không bật lịch chạy (cron) của workflow `sync-electricity.yml`, vì app đã tự sync.
- [ ] Cập nhật bản mới: build lại file `volt`, upload đè, rồi restart.
