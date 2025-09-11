lists all the endpoints

# HTTP Endpoints

## GET
- `/` - Serves static frontend (mainpage.html)
- `/dashboard` - Dashboard page (dashboard.html)
- `/user/:name` - Get Twitch user info (JSON)
- `/stream/:name` - Get Twitch stream info (JSON)
- `/games/top` - Get top Twitch games (JSON)
- `/ws` - WebSocket endpoint for live chat
- `/irc/subscribe/:channel` - Subscribe to IRC chat for a channel (JSON or HTML)
- `/irc/unsubscribe/:channel` - Unsubscribe from IRC chat for a channel (JSON or HTML)
- `/authorize` - Start OAuth authorization with Twitch
- `/auth/status` - Check OAuth status (JSON)
- `/irc/stream/:channel` - SSE stream of IRC chat messages for a channel

## POST
- `/irc/subscribe` - Subscribe to IRC chat for a channel (JSON, body: `{channel}`)
- `/irc/subscribe/:channel` - Subscribe to IRC chat for a channel (JSON)
- `/irc/unsubscribe` - Unsubscribe from IRC chat for a channel (JSON, body: `{channel}`)
- `/irc/send` - Send a chat message to a channel (JSON, body: `{channel, message}`)
