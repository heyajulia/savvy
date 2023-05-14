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