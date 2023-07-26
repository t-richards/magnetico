# Builder image
FROM golang:1.20-alpine AS builder

# Config
ENV GOFLAGS="-trimpath -mod=readonly -modcacherw"
ENV CGO_ENABLED=0
WORKDIR /go/src/app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download -x

# Main application source
COPY . .

# Build binary
RUN go build -ldflags="-s -w"

# Intermediate image step
FROM scratch AS root
COPY --chown=0:0 etc /etc
COPY --from=builder /go/src/app/data /data
COPY --from=builder /go/src/app/magnetico /bin/magnetico

# Runner image
FROM scratch AS runner
COPY --from=root / /
EXPOSE 8080/tcp
WORKDIR /data
USER app
ENTRYPOINT ["/bin/magnetico"]
CMD [""]
