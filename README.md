# energieprijzen-bot [![Go Report Card](https://goreportcard.com/badge/github.com/heyajulia/energieprijzen)](https://goreportcard.com/report/github.com/heyajulia/energieprijzen)

Run with:

```bash
go run cmd/bot/main.go
```

Run tests with:

```bash
go test ./...
```

Pushing a new version:

```bash
TAG=v1.0.0
git tag $TAG
git push --atomic origin main $TAG
```

## Running the bot on a schedule

Tested on a clean install of Ubuntu Server 22.04.3 LTS with a user named `julia`.

```
sudo apt update
sudo apt install git vim
git clone https://github.com/heyajulia/energieprijzen.git
cd energieprijzen/
```

Then, either download a pre-built binary, or build it yourself:

```bash
VERSION=v1.0.18
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
go build -o energieprijzen cmd/bot/main.go
```

Then, copy the systemd files, edit the token file, and enable the timer:

```
sudo cp init/energieprijzen.* /etc/systemd/system
vim token.txt
sudo systemctl daemon-reload
sudo systemctl enable --now energieprijzen.timer
```

Check that it worked using:

```bash
sudo systemctl status energieprijzen.{service,timer}
```

Note: it's fine if `sudo systemctl status energieprijzen.service` returns a non-zero exit code.

Note: you can run the bot at any time with `sudo systemctl start energieprijzen.service`.
