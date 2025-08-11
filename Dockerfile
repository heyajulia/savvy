FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

ARG TARGETOS

ARG TARGETARCH

ARG VERSION

ARG COMMIT

WORKDIR /build

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /build/serve -trimpath -ldflags="-s -w -X 'github.com/heyajulia/savvy/internal.Version=${VERSION}' -X 'github.com/heyajulia/savvy/internal.Commit=${COMMIT}'" ./cmd/serve && \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /build/report -trimpath -ldflags="-s -w -X 'github.com/heyajulia/savvy/internal.Version=${VERSION}' -X 'github.com/heyajulia/savvy/internal.Commit=${COMMIT}'" ./cmd/report

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder --chown=nonroot:nonroot /build/serve /build/report /home/nonroot/

ENTRYPOINT ["/home/nonroot/serve"]
