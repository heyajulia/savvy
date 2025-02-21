FROM golang:1.24.0-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /build/bot -trimpath -ldflags="-s -w" ./cmd/bot

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder --chown=nonroot:nonroot /build/bot /home/nonroot

CMD ["/home/nonroot/bot"]
