# Proxy checker
## Overwiew
Checks list of SOCKS5 proxies. 
Has simple UI through Telegram bot.

## How to run
### Start responder
```
./responder
```
Responder opens simple HTTP server serving thee paths:
* `/ip`: return ip address of client
* `/stats`: return some stats
* `/reset`: reset stats

### Prepare proxies list.
By default `proxychecker` reads `proxies.txt` file.
This file must contain list of addresses to check in format `<IP>:<PORT>`, one per line:  
```
1.2.3.4:1080
5.6.7.8:1014
```

### Start bot
```
proxychecker --check-url <RESPONDER_URL> \
  --max-proxies 20000\
  --tg-api-token <TELEGRAM_BOT_API_TOKEN>\
  --tg-proxy-address <SOCKS5_PROXY_TO_TELEGRAM_API>
```
Check `proxychecker --help` for full list of available arguments.

### Use bot
Issue `/help` or `/start` command to the Bot.
