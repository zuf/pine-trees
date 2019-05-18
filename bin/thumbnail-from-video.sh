#!/usr/bin/env sh

set -e

ffmpeg -i "${1}" -vf "select=gte(n\,100),scale=-1:400" -vframes 1 -f image2pipe pipe:1 | cat
