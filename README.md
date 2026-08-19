# Volt

Volt is a personal project that collects electricity consumption data available through the EVNHCMC customer portal, stores historical data in PostgreSQL, and sends notifications through Telegram.

The project is intended for personal and educational use.

## Architecture

```text
                    GitHub Actions
                  +----------------+
                  | Cron / Worker  |
                  +-------+--------+
                          |
                          v
                    Go Application
                          |
              +-----------+-----------+
              v                       v
        EVNHCMC Portal            Supabase
        HTTP Client               PostgreSQL
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

MIT
