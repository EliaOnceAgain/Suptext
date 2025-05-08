FROM golang:latest

ENV TESSDATA_PREFIX=/usr/share/tesseract-ocr/5/tessdata/
COPY ./runner.sh /usr/bin/supcli
RUN apt-get update && \
    apt-get install -y file libtesseract-dev libleptonica-dev tesseract-ocr-eng ffmpeg && \
    chmod +x /usr/bin/supcli && \
    go install github.com/eliaonceagain/suptext@latest
CMD ["tail", "-f", "/dev/null"]
