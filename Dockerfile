# ---------- build stage ----------
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Install build dependencies for CGO
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        build-essential \
        libtesseract-dev \
        libleptonica-dev \
        tesseract-ocr-eng && \
    rm -rf /var/lib/apt/lists/*

# Install the Go binary
RUN go install github.com/eliaonceagain/suptext@v0.2.2

# ---------- runtime stage ----------
FROM debian:bookworm-slim

ENV TESSDATA_PREFIX=/usr/share/tesseract-ocr/5/tessdata/

# Install runtime dependencies only
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        file \
        tesseract-ocr \
        tesseract-ocr-eng \
        ffmpeg \
        liblept5 \
        libtesseract5 && \
    rm -rf /var/lib/apt/lists/*

# Copy the compiled binary from builder
COPY --from=builder /go/bin/suptext /usr/bin/suptext

# Copy runner script
COPY ./runner.sh /usr/bin/supcli
RUN chmod +x /usr/bin/supcli

CMD ["tail", "-f", "/dev/null"]
