# Savvy

[![Go Report Card](https://goreportcard.com/badge/github.com/heyajulia/savvy)](https://goreportcard.com/report/github.com/heyajulia/savvy)

Savvy posts Dutch energy prices to Telegram and Bluesky. If you have a dynamic
energy contract ("dynamisch energiecontract"), it helps you see when electricity
is cheapest.

Find the bot on [Bluesky](https://bsky.app/profile/bot.julia.cool) and
[Telegram](https://t.me/energieprijzenbot) (or subscribe to
[the channel](https://t.me/energieprijzen)).

## Installation

Download the latest binary from
[GitHub Releases](https://github.com/heyajulia/savvy/releases). Savvy can also
update itself with `savvy upgrade`.

### Systemd setup

To run Savvy as a systemd service:

```sh
# Create user and directories
sudo useradd -r -s /usr/sbin/nologin savvy
sudo mkdir -p /etc/savvy /var/lib/savvy/stamps
sudo chown savvy:savvy /var/lib/savvy/stamps

# Install binary
sudo cp savvy /usr/local/bin/

# Configure environment (edit with your credentials)
sudo cp init/savvy.env.example /etc/savvy/savvy.env
sudo chmod 600 /etc/savvy/savvy.env
sudo chown savvy:savvy /etc/savvy/savvy.env

# Install and enable services
sudo cp init/savvy.service init/savvy-report.service init/savvy-report.timer /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now savvy savvy-report.timer
```

## Contributing

If you have suggestions or improvements, feel free to open an issue or pull
request.
