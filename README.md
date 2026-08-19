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
| Live Reload     | Air                   | Live reloading during local dev      |
| Env Management  | direnv                | Auto-load environment variables      |


## Development & Environment Setup

### Tools

- **Air**: Used for live reloading during Go application development (`.air.toml`).
- **direnv**: Used to automatically load environment variables from `.envrc` upon entering the directory.

### Environment Setup

1. Configure your local environment variables in `.envrc`.
2. Allow `direnv` to automatically load environment variables when navigating into the project directory


```bash
direnv allow
```

### Running Locally

To run the application with live reload enabled via Air:

```bash
air
```

Or run directly using Go:

```bash
go run ./cmd/api
```

## License

MIT
