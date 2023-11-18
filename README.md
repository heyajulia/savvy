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
> The property list (plist) file used to configure the job will contain your bot token (a password) in plain text. This
> is not ideal. I'm working on remedying this situation (see heyajulia/energieprijzen#23).

1. Download the binary from the [releases page](https://github.com/heyajulia/energieprijzen/releases)
2. Move it somewhere, e.g. `~/bin`
3. Make it executable: `chmod +x ~/bin/energieprijzen`
4. Copy the plist file from this repo to `/Library/LaunchDaemons/`. Adjust as necessary.

   ```bash
   sudo -e /Library/LaunchDaemons/cool.julia.bot.energieprijzen.plist
   ```

5. Adjust the permissions (I have no idea if this is necessary, but everyone else seems to do it):

   ```bash
   chown root:wheel /Library/LaunchDaemons/cool.julia.bot.energieprijzen.plist
   chmod 644 /Library/LaunchDaemons/cool.julia.bot.energieprijzen.plist
   ```

6. Load the plist file with `launchctl bootstrap system /Library/LaunchDaemons/cool.julia.bot.energieprijzen.plist`
7. Check that it worked with `launchctl print system/cool.julia.bot.energieprijzen`

> [!TIP]
>
> If you run into the cryptic error message `Bootstrap failed: 5: Input/output error`, verify your plist file with
> `plutil`: `plutil /Library/LaunchDaemons/cool.julia.bot.energieprijzen.plist`. It'll say something like (and yes, this
> is a real mistake that I made):
>
> ```
> /Library/LaunchDaemons/cool.julia.bot.energieprijzen.plist: Encountered unknown tag UserName on line 35
> ```

> [!IMPORTANT]
>
> If no message was sent for any reason, you can "kickstart" the job:
>
> ```bash
> launchctl kickstart system/cool.julia.bot.energieprijzen
> ```

### Uninstall

> [!CAUTION]
>
> These instructions are out of date. I'll update them when I get around to it.

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
