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

TODO: test these instructions in a clean VM

```
git clone https://github.com/heyajulia/energieprijzen.git
cd energieprijzen/
wget https://github.com/heyajulia/energieprijzen/releases/download/VERSION/energieprijzen
chmod +x energieprijzen
sudo cp init/energieprijzen.* /etc/systemd/system
vim token.txt
sudo systemctl daemon-reload
sudo systemctl enable --now energieprijzen.timer
```
