# ☀️ energieprijzen [![Go Report Card](https://goreportcard.com/badge/github.com/heyajulia/energieprijzen)](https://goreportcard.com/report/github.com/heyajulia/energieprijzen)

**energieprijzen** ("energy prices" in Dutch) is a Telegram bot that posts tomorrow's energy prices to
[a Telegram channel](https://t.me/energieprijzen) and [on Bluesky](https://bsky.app/profile/bot.julia.cool). If, like
me, you have a **dynamic energy contract** ("dynamisch energiecontract") at ANWB Energie, this bot could help you save
both time and money.

You can interact with the bot directly at [@energieprijzenbot](https://t.me/energieprijzenbot). It doesn't do much, but
it can tell you about your privacy rights when using the bot. [The privacy policy](./cmd/bot/templates/privacy.tmpl) is
only available in Dutch for now, but the long and short of it is that I'm actively disinterested in your data.

## 🤖 Installation and usage

You likely won't need to run the bot yourself, as you can just join the channel or start a chat with the instance I set
up (see above). However, if you want to run your own instance, follow these steps:

1. Download the latest binary from the [Releases page](https://github.com/heyajulia/energieprijzen/releases)
   (`linux/amd64` only for now, but it should build and run on other platforms as well).
2. Create a `config.json` file with your Telegram bot token and chat ID, as well as your Bluesky credentials:

   ```json
   {
     "telegram": {
       "token": "your-telegram-bot-token",
       "chat_id": 123456789
     },
     "bluesky": {
       "identifier": "username.bsky.social",
       "password": "your-app-specific-password"
     },
     "cronitor": {
       "telemetry_url": "https://cronitor.link/your/telemetry/endpoint"
     }
   }
   ```

   - The `cronitor` section is optional.
   - The chat ID can also be a string username if you prefer.

3. Run the bot:

   ```sh
   ./energieprijzen
   ```

   It'll start an infinite loop, responding to incoming messages and posting the energy prices at the right time.

4. To check the version number and build timestamp, run:

   ```sh
   ./energieprijzen -v
   ```

## 🔨 Contributing

If you have any suggestions or improvements, feel free to open an issue or a pull request. I'd be happy to hear from
you!
