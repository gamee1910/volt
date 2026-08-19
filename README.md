# Volt

Electricity consumption tracker for EVNHCMC (Ho Chi Minh City Power Corporation). Built with Go.

Volt connects to the EVNHCMC customer portal, authenticates via session cookies, fetches daily electricity consumption data, stores it in PostgreSQL, and sends notifications to Telegram.

## Architecture

```text
                    GitHub Actions
                  +----------------+
                  | Cron every 30m |
                  +-------+--------+
                          |
                          v
                    Go Application
                          |
              +-----------+-----------+
              v                       v
        EVNHCMC API                Supabase
        login + cookie             PostgreSQL
              |                       |
              +----------+------------+
                         |
                         v
                  Business Logic
                         |
                         v
                   Telegram Bot
```

### Stack

| Component       | Technology            | Purpose                              |
|-----------------|-----------------------|--------------------------------------|
| Backend         | Go                    | HTTP client, business logic, worker  |
| Database        | PostgreSQL (Supabase) | Persistent storage                   |
| Scheduler       | GitHub Actions        | Cron-based worker, no VPS required   |
| Notifications   | Telegram Bot API      | Push alerts to phone                 |

## Design Principles

1. Keep it simple. No unnecessary abstractions.
2. Separate the EVNHCMC client from business logic.
3. One `http.Client` per session. Never recreate it between login and data fetch.
4. Use `http.CookieJar` for automatic cookie management.
5. No hardcoded credentials. Everything comes from environment variables.
6. Never commit `.env` or `.envrc`.
7. Do not guess EVNHCMC payload fields. Only use what has been observed in actual requests.
8. Parse all responses into explicit Go structs.
9. Get it working first, then refactor.

## License

Private project.
