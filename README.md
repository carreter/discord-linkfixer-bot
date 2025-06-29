# linkfixer: A simple bot to fix URLs in Discord messages
[![.github/workflows/release.yaml](https://github.com/carreter/discord-linkfixer-bot/actions/workflows/release.yaml/badge.svg)](https://github.com/carreter/discord-linkfixer-bot/actions/workflows/release.yaml)

A Discord bot that automatically fixes URLs posted in your server by applying customizable transformations to links from specific domains.

## Features
- **Domain-Specific Fixes**: Configure different fixes for different domains
- **Multiple Fixer Types**:
  - **Replace**: Simple string replacement in URLs
  - **Regex Replace**: Advanced pattern matching with capture groups
  - **Prepend**: Add prefixes to URLs
- **Per-Server Configuration**: Each Discord server maintains its own set of URL fixers

## Use Cases

Perfect for fixing common URL issues like:
- Converting mobile links to desktop versions
- Adding bypass parameters for paywalled content
- Redirecting through privacy-friendly frontends
- Fixing broken social media embeds

## Installation

### Prerequisites
- Go 1.24.4 or later
- Discord bot token

### Building from Source

```bash
git clone https://github.com/carreter/discord-linkfixer-bot
cd discord-linkfixer-bot
go build -o linkfixer-bot
```

### Running the Bot

```bash
./linkfixer-bot -token YOUR_DISCORD_BOT_TOKEN
```

Optional flags:
- `-db PATH`: Path to database file (default: `./fixers.db`)

## Discord Commands

### `/replace-fixer`
Register a simple string replacement fixer.
- `domain`: The domain to apply this fixer to
- `old`: Substring to replace
- `new`: Replacement substring

**Example**: Fix Twitter mobile links
```
/replace-fixer domain:twitter.com old:mobile.twitter.com new:twitter.com
```

### `/regexp-replace-fixer`
Register a regex-based fixer with capture groups.
- `domain`: The domain to apply this fixer to
- `pattern`: Regular expression pattern
- `replacement`: Replacement string (use `$1`, `$2`, etc. for capture groups)

**Example**: Convert Reddit mobile links
```
/regexp-replace-fixer domain:reddit.com pattern:m\.reddit\.com replacement:old.reddit.com
```

### `/prepend-fixer`
Add a prefix to URLs from a domain.
- `domain`: The domain to apply this fixer to
- `prefix`: String to prepend to the URL

**Example**: Add privacy redirect
```
/prepend-fixer domain:youtube.com prefix:https://invidio.us/
```

### `/list-fixers`
List all registered fixers for the current server.

### `/delete-fixer`
Remove a fixer for a specific domain.
- `domain`: Domain of the fixer to delete

## Setup Instructions

1. Create a Discord application at https://discord.com/developers/applications
2. Create a bot user and copy the token
3. Invite the bot to your server with these permissions:
   - Read Messages
   - Send Messages
   - Use Slash Commands
4. Run the bot with your token

## Development

### Project Structure
```
├── main.go                          # Entry point
├── pkg/
│   ├── fixer/                      # URL fixer implementations
│   │   ├── fixer.go               # Fixer interfaces and types
│   │   └── store.go               # BoltDB storage layer
│   └── linkfixerbot/              # Discord bot implementation
│       ├── bot.go                 # Main bot logic
│       └── commands/              # Slash command handlers
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

Contributions welcome! Please open an issue or submit a pull request.
