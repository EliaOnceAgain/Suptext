#!/bin/bash
set -eu

function logerr {
  echo "ERROR: $*" >&2
  exit 1
}

function log {
  timestamp="$(date +"%Y/%m/%d %H:%M:%S")"
  echo "${timestamp} $*"
}

function help {
  echo "Usage: $0 [OPTION] <path>"
  echo "Exactly one of the following options must be set"
  echo "  -v,  --video <path>      : Set the video file"
  echo "  -s,  --sup <path>        : Set the supplementary file"
  echo "  -h,  --help              : Display this help message"
  exit 1
}

function validate_input {
  count=0
  [ -n "${input_video}" ] && count=$((count + 1))
  [ -n "${input_sup}" ] && count=$((count + 1))
  [ "${count}" -eq 1 ] || help
}

function get_tempfile {
  if ! tmpfile=$(mktemp) ; then
    logerr "Failed creating temp file"
  fi
  echo "${tmpfile}"
}

function get_largest_size_eng_pgs_stream_index {
  # Probe for PGS ENG subtitles in input video
  eng_pgs_stream="$(ffprobe -loglevel error \
    -select_streams s \
    -show_entries stream=index:stream_tags=language:stream=codec_name:stream_tags=number_of_bytes-eng \
    -of csv=p=0 \
    "${input_video}" | grep eng | grep hdmv_pgs_subtitle)"
  # Validate PGS ENG subtitles was found
  if [[ -z "${eng_pgs_stream}" ]]; then
    logerr "No eng-PGS subtitles detected in video file."
  fi
  if [[ $(echo "${eng_pgs_stream}" | wc -l) == 1 ]]; then
    echo "${eng_pgs_stream%%,*}"
  else
    # If more than one eng-PGS stream then return the largest size stream
    max_size=-1
    largest_stream_index=-1
    while IFS=, read -r stream_index _ _ stream_size; do
        if (( stream_size > max_size )); then
            max_size=${stream_size}
            largest_stream_index=${stream_index}
        fi
    done <<< "${eng_pgs_stream}"
    echo "${largest_stream_index}"
  fi
}

function video_input {
  if file -i "${input_video}" | grep -q video ; then
    stream_index="$(get_largest_size_eng_pgs_stream_index)"
    log "Found eng-PGS subtitles at stream index: ${stream_index}"
  else
    logerr "Input is not a video file"
  fi

  input_video_path_no_extension="${input_video%.*}"
  tempfile_path="$(get_tempfile)"
  # Extract subtitles
  if ! ffmpeg -loglevel error -i "${input_video}" -map 0:"${stream_index}" -c:s copy "${tempfile_path}.sup"; then
    logerr "Failed to extract subtitles from input video"
  fi
  log "Successfully extracted subtitles from input video"
  suptext "${tempfile_path}.sup"
  # Cleanup
  cp "${tempfile_path}.srt" "${input_video_path_no_extension}.srt"
  rm "${tempfile_path}.sup" "${tempfile_path}.srt"
}

input_video=""
input_sup=""
while [[ $# -gt 0 ]]; do
  case $1 in
    -v|--video)
      input_video="$2"
      shift
      ;;
    -s|--sup)
      input_sup="$2"
      shift
      ;;
    *)
      help
      ;;
  esac
  shift
done

validate_input
if [[ -n "${input_video}" ]]; then
  log "Processing video file: $(basename "${input_video}")"
  video_input
else
  log "Processing video file: $(basename "${input_sup}")"
  suptext "${input_sup}"
fi
