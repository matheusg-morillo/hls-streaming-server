#!/usr/bin/env bash
# ^ "shebang": tells the OS to use bash to run this file.

# exit on error, error on unset variables, fail whole pipe if any step fails
set -euo pipefail

# ---------------------------------------------------------------------------
# generate_hls.sh — Convert an MP4 into multi-resolution HLS
#
# Produces one folder per resolution, each with .ts segments and a media
# playlist, plus a top-level master.m3u8 for adaptive bitrate switching.
#
# Usage:
#   ./generate_hls.sh <input.mp4> [output_dir] [segment_duration]
#
# Examples:
#   ./generate_hls.sh movie.mp4
#   ./generate_hls.sh movie.mp4 ./output 6
# ---------------------------------------------------------------------------

INPUT="${1:-}"
OUTPUT_DIR="${2:-./hls_output}"
SEGMENT_DURATION="${3:-10}"

# ---------------------------------------------------------------------------
# Resolution ladder — each entry is a colon-separated string:
#   label : width : height : video_bitrate : audio_bitrate
#
# The player will pick the best variant based on network speed.
# Bitrates are intentional: lower resolution = lower bitrate = less data.
# ---------------------------------------------------------------------------
RESOLUTIONS=(
  "360p:640:360:800k:96k"
  "480p:854:480:1400k:128k"
  "720p:1280:720:2800k:128k"
  "1080p:1920:1080:5000k:192k"
)

# --- Validation -------------------------------------------------------------

usage() {
  echo "Usage: $0 <input.mp4> [output_dir] [segment_duration_seconds]"
  exit 1
}

[[ -z "$INPUT" ]] && {
  echo "Error: no input file provided."
  usage
}
[[ ! -f "$INPUT" ]] && {
  echo "Error: file '$INPUT' not found."
  exit 1
}

if ! command -v ffmpeg &>/dev/null; then
  echo "Error: ffmpeg is not installed or not in PATH."
  exit 1
fi

if ! [[ "$SEGMENT_DURATION" =~ ^[1-9][0-9]*$ ]]; then
  echo "Error: segment_duration must be a positive integer (got '$SEGMENT_DURATION')."
  exit 1
fi

# --- Setup ------------------------------------------------------------------

# Strip directory path and .mp4 extension from input to get a clean base name.
# E.g. "./videos/my-movie.mp4" → "my-movie"
BASENAME=$(basename "$INPUT" .mp4)

mkdir -p "$OUTPUT_DIR"

# This file will reference all resolution variants so the player can
# switch quality levels automatically. We write it incrementally below.
MASTER_PLAYLIST="$OUTPUT_DIR/master.m3u8"

# Start the master playlist with required HLS headers.
# The ">" operator creates (or overwrites) the file.
{
  echo "#EXTM3U"          # marks this as an HLS playlist
  echo "#EXT-X-VERSION:3" # HLS protocol version
} >"$MASTER_PLAYLIST"

echo "Input:            $INPUT"
echo "Output directory: $OUTPUT_DIR"
echo "Segment duration: ${SEGMENT_DURATION}s"
echo "Master playlist:  $MASTER_PLAYLIST"
echo ""

# --- Encode each resolution -------------------------------------------------

# Loop over each entry in the RESOLUTIONS array.
# IFS=':', read -r splits the colon-separated string into individual variables.
for RES in "${RESOLUTIONS[@]}"; do
  IFS=':' read -r LABEL WIDTH HEIGHT VIDEO_BIT AUDIO_BIT <<<"$RES"

  echo ">>> Encoding ${LABEL} (${WIDTH}x${HEIGHT}, video=${VIDEO_BIT}, audio=${AUDIO_BIT})..."

  # Each resolution gets its own subdirectory to keep segments organized.
  RES_DIR="$OUTPUT_DIR/$LABEL"
  mkdir -p "$RES_DIR"

  MEDIA_PLAYLIST="$RES_DIR/$BASENAME.m3u8"
  SEGMENT_PATTERN="$RES_DIR/${BASENAME}_%03d.ts"
  # %03d → ffmpeg replaces this with 000, 001, 002… for each segment file.

  ffmpeg -y -i "$INPUT" \
    -vf "scale=${WIDTH}:${HEIGHT}" \
    -c:v libx264 \
    -b:v "$VIDEO_BIT" \
    -c:a aac \
    -b:a "$AUDIO_BIT" \
    -f hls -hls_time "$SEGMENT_DURATION" \
    -hls_playlist_type vod \
    -hls_segment_filename "$SEGMENT_PATTERN" \
    "$MEDIA_PLAYLIST"

  # After encoding, append this variant to the master playlist.
  # BANDWIDTH must be in bits/s; VIDEO_BIT is like "800k" so we strip the "k"
  # and multiply by 1000.
  BANDWIDTH_BPS=$((${VIDEO_BIT%k} * 1000))

  # ">>" appends to the file (vs ">" which would overwrite it).
  {
    echo "#EXT-X-STREAM-INF:BANDWIDTH=${BANDWIDTH_BPS},RESOLUTION=${WIDTH}x${HEIGHT}"
    # Path is relative to the master playlist location.
    echo "${LABEL}/${BASENAME}.m3u8"
  } >>"$MASTER_PLAYLIST"

  echo "    Done. Playlist: $MEDIA_PLAYLIST"
  echo ""
done

# --- Summary ----------------------------------------------------------------

echo "All resolutions encoded."
echo "Master playlist: $MASTER_PLAYLIST"
