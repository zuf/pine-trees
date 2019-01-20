#!/usr/bin/env sh

set -e

#ffmpeg -i ${1} -c:v copy -f flv -
#ffmpeg -i "${1}" -c:v copy -c:a aac -f mpegts -
ffmpeg -i "${1}" -c:v copy -c:a aac -movflags frag_keyframe+empty_moov -f mp4 -
