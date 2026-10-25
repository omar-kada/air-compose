#!/usr/bin/env bash
# Joins cli.mp4 + ui.webm into demo.mp4 and makes a README-sized demo.gif.
set -euo pipefail
cd "$(dirname "$0")/out"

ffmpeg -y -loglevel error -i cli.mp4 -i ui.webm -filter_complex \
"[0:v]fps=30,scale=1280:720,setsar=1[a];\
 [1:v]fps=30,scale=1280:720,setsar=1[b];\
 [a][b]concat=n=2:v=1:a=0[v]" \
 -map "[v]" -c:v libx264 -crf 22 -pix_fmt yuv420p -movflags +faststart demo.mp4

ffmpeg -y -loglevel error -i demo.mp4 \
 -vf "fps=12,scale=800:-1:flags=lanczos,split[s0][s1];[s0]palettegen[p];[s1][p]paletteuse" \
 demo.gif

ls -lh demo.mp4 demo.gif
