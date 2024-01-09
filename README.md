# energieprijzen-bot [![Go Report Card](https://goreportcard.com/badge/github.com/heyajulia/energieprijzen)](https://goreportcard.com/report/github.com/heyajulia/energieprijzen)

Run with:

```bash
go run ./cmd/bot
```

Run tests with:

```bash
go test ./...
```

To release a new version, run `./script/release`.

To upgrade Templ, run `./script/upgrade-templ`.

## Running the bot on a schedule

Tested on a clean install of Ubuntu Server 22.04.3 LTS with a user named `julia` with the repo cloned to
`/home/julia/energieprijzen`.

```
sudo apt update
sudo apt install git vim
git clone https://github.com/heyajulia/energieprijzen.git
cd energieprijzen/
```

Then, either download a pre-built binary, or build it yourself:

```bash
VERSION=v1.0.30
wget https://github.com/heyajulia/energieprijzen/releases/download/$VERSION/energieprijzen
chmod +x energieprijzen
```

or:

```bash
cd $(mktemp -d)
wget https://go.dev/dl/go1.21.5.linux-arm64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-arm64.tar.gz
export PATH=$PATH:/usr/local/go/bin
cd -
go build -o energieprijzen ./cmd/bot
```

Then, copy the systemd files and edit the token file:

```
sudo cp init/energieprijzen.* /etc/systemd/system
vim token.txt
```

The `token.txt` file should have contents similar to:

```json
{
  "telegram": "foo",
  "cronitor_url": "bar"
}
```

Where `telegram` is the Telegram bot token, and `cronitor_url` is your Cronitor Telemetry URL.

Then, enable and start the timer:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now energieprijzen.timer
```

Check that it worked using:

```bash
systemctl status energieprijzen.{service,timer}
```

> [!TIP]
>
> You can run the bot at any time with `sudo systemctl start energieprijzen.service`.
