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

## Running the bot on a schedule using a launch agent

### Setup

> [!WARNING]
>
> The plist file used to configure the job will contain your bot token (a password) in plain text. This is not ideal.
> I'm working on remedying this situation (see heyajulia/energieprijzen#23).

1. Download the binary from the [releases page](https://github.com/heyajulia/energieprijzen/releases)
2. Move it somewhere, e.g. `~/bin`
3. Copy the property list file from this repo to `~/Library/LaunchAgents/cool.julia.bot.energieprijzen.plist`
4. Adjust the paths and environment variables in the property list file to match your setup
5. Load the property list file with
   `launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/cool.julia.bot.energieprijzen.plist`
6. Check that it's running with `launchctl list | grep energieprijzen`

> [!IMPORTANT]
>
> If you're not logged in, no message will be sent (see heyajulia/energieprijzen#25). However, you can "kickstart" the
> job to send a message immediately:
>
> ```bash
> launchctl kickstart gui/$(id -u)/cool.julia.bot.energieprijzen
> ```

### Uninstall

1. ```bash
   launchctl bootout gui/$(id -u)/cool.julia.bot.energieprijzen
   rm ~/Library/LaunchAgents/cool.julia.bot.energieprijzen.plist
   rm -r ~/Library/Logs/energieprijzen
   ```
2. The command to remove the binary depends on where you put it. For example, if you put it in `~/bin` like I did, run:

   ```bash
   rm ~/bin/energieprijzen
   ```

3. Reboot
