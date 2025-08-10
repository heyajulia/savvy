FROM --platform=$BUILDPLATFORM golang:1.24.0-alpine AS builder

ARG TAGETOS

ARG TARGETARCH

ARG VERSION

ARG COMMIT

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /build/bot -trimpath -ldflags="-s -w -X 'main.version=${VERSION}' -X 'main.commit=${COMMIT}'" ./cmd/bot

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder --chown=nonroot:nonroot /build/bot /home/nonroot

ENTRYPOINT ["/home/nonroot/bot"]
