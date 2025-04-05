# RSS Aggregator CLI (gator)

A command-line tool to aggregate and manage RSS feeds.

## Prerequisites

- Go (version 1.24 or later)
- PostgreSQL database

## Installation

```bash
# Clone the repository
git clone https://github.com/kylektaylor1/rss-aggregator.git

# Navigate to the repository
cd rss-aggregator

# Install the gator CLI
go install
```

## Configuration

Create a `.gatorconfig.json` file in your home directory with the following content:

```json
{
  "db_url": "postgres://username:password@localhost:5432/database_name",
  "current_user_name": ""
}
```

Replace the `db_url` with your PostgreSQL connection string.

## Database Setup

Before using the application, you need to set up the database schema:

1. Create a PostgreSQL database
2. Execute SQL schema files in the `sql/schema/` directory in numerical order

## Commands

- `gator register <username>` - Register a new user
- `gator login <username>` - Login as an existing user
- `gator addfeed <feed-url> <feed-name>` - Add a new RSS feed
- `gator feeds` - List all feeds
- `gator follow <feed-id>` - Follow a feed
- `gator following` - List feeds you're following
- `gator unfollow <feed-id>` - Unfollow a feed
- `gator browse <optional_limt>` - Browse posts from feeds you follow (default limit = 2)
- `gator agg <time_duration>` - Aggregate new posts from all feeds (time duration = "2s", "10m" or similar)
  -- You should run this as a long standing service to continually scrape feeds

