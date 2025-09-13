# Suptext
Suptext is a PGS/SUP to SRT subtitles converter written in Go and uses
[Tesseract](https://github.com/tesseract-ocr/tesseract) for OCR.

### Intro
Suptext expects an english PGS/SUP encoded subtitles file as input, and outputs an SRT SubRip file.  
If you're unable or couldn't be bothered to extract the SUP file yourself from a video then jump to the 
Run via Docker section for a utility that supports video input and does that for you.

### Run Locally
- Install [Tesseract-OCR](https://github.com/tesseract-ocr/tessdoc/blob/main/Installation.md) based on your
OS and download the english language package 
- Run `go install github.com/eliaonceagain/suptext@latest`

### Run via Docker
The following instructions use [eliaonceagain/suptext](https://hub.docker.com/r/eliaonceagain/suptext/tags) Docker image.
The image contains a utility command that extracts eng-PGS/SUP subtitles from a video file using `ffmpeg` 
then calls `Suptext` Go package to decode and output the result SRT SubRip file.

- Run a container based on `eliaonceagain/suptext` Docker image and mount a media directory.  
If your media is not in the current directory then change `"$(pwd)"` accordingly.
```bash
docker run --rm -itd \
       --name suptext \
       -v "$(pwd)":/mymedia \
       eliaonceagain/suptext:latest
```
- Verify container is up
```bash
docker exec suptext supcli --help
```
```text
Usage: /usr/bin/supcli [OPTION] <path>
Exactly one of the following options must be set
  -v,  --video <path>      : Set the video file
  -s,  --sup <path>        : Set the supplementary file
  -h,  --help              : Display this help message
```
- Run `supcli` command in the container while providing either a video or a PGS/SUP file. 
Output filename will be identical to input filename besides having `.srt` extension; e.g. `video.mkv -> video.srt` 
```bash
# from video file
docker exec suptext supcli -v "/mymedia/video.mkv"

# or from sup file
docker exec suptext supcli -s "/mymedia/subtitles.sup"
```
- Example output
```bash
docker exec suptext supcli -v "/mymedia/SV-S04E10.mkv"
```
```text
2024/07/04 17:20:15 Processing video file: SV-S04E10.mkv
2024/07/04 17:20:16 Found eng-PGS subtitles at stream index: 3
2024/07/04 17:20:23 Successfully extracted subtitles from input video
2024/07/04 17:20:23 Reading SUP file: /tmp/tmp.xY7TPdfhCp.sup
2024/07/04 17:20:23 Writing SRT file: /tmp/tmp.xY7TPdfhCp.srt
2024/07/04 17:26:31 Success
```

### Changelog
v0.1.1 - Added support for fragmented ODS  
v0.1.2 - Ignore truncated PCS extension  
v0.1.3 - Fix for issue #5 (output contains timestamps but no text)

### Issues
Not working as expected? Open an issue, add description, and upload a link to the video or subtitles.

### SRT Cleanup
Want to further cleanup your output SRT file? Check out [SRT-Link](https://github.com/EliaOnceAgain/SRT-Link), a CLI tool for filtering SRT subtitles based on 
specific text, characters, or brackets.
