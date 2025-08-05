# ☀️ Savvy [![Go Report Card](https://goreportcard.com/badge/github.com/heyajulia/savvy)](https://goreportcard.com/report/github.com/heyajulia/savvy)

**Savvy** is a Telegram bot that posts tomorrow's energy prices to Telegram and Bluesky. If, like me, you have a
**dynamic energy contract** ("dynamisch energiecontract") at ANWB Energie, Savvy could help you save both time and
money.

You can find the bot on [Bluesky](https://bsky.app/profile/bot.julia.cool) and
[Telegram](https://t.me/energieprijzenbot) (though [the channel](https://t.me/energieprijzen) might be more
interesting).

## 🤖 Installation and usage

You most likely won't need to run the bot yourself. However, if you want to run your own instance for some reason,
here's how:

1. Create an `.env` file:

   ```env
   # https://t.me/userinfobot
   TG_CHAT_ID=channelusername
   # https://t.me/botfather
   TG_TOKEN=123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11
   # https://bsky.app
   BS_IDENTIFIER=botusername.bsky.social
   # https://bsky.app/settings/app-passwords
   BS_PASSWORD=1234-abcd-5678-efgh
   # https://cronitor.io/app, optional
   CR_URL=https://cronitor.link/p/your-monitor-id-here
   ```

   The chat ID can also be a user ID.

2. Run the bot:

   ```sh
   docker run --detach --restart always --name savvy -it --env-file .env ghcr.io/heyajulia/savvy
   ```

   It'll start an infinite loop, responding to incoming messages and posting the energy prices at the right time. 


## 🔨 Contributing

If you have any suggestions or improvements, feel free to open an issue or a pull request. I'd be happy to hear from
you!
