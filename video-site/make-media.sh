#!/bin/sh
# Makes the films the example plays, in video-site/media, with ffmpeg
# (libx264): two films of ffmpeg's test patterns and tones, each at 360p
# and at 720p, with a poster and English subtitles, cut for HLS beforehand
# only in that their keyframes come every 2s; fit serves the files as they
# are, and never transcodes. Your own films go in video-site/media the same
# way: films/<name>.mp4, films/<name>.720p.mp4, films/<name>.jpg,
# films/<name>.en.srt.
#
#   video-site/make-media.sh [SECONDS]
#
# SECONDS (60 by default) is how long each film is.
#
# 用 ffmpeg（libx264）生成本示例播放的影片，放在 video-site/media：两部由
# ffmpeg 测试图案和音调构成的影片，各有 360p 和 720p 两个版本，带海报和
# 英文字幕；事先为 HLS 做的唯一准备是每 2 秒一个关键帧。fit 按文件原样提供，
# 从不转码。你自己的影片也照此放入 video-site/media：films/<name>.mp4、
# films/<name>.720p.mp4、films/<name>.jpg、films/<name>.en.srt。
# SECONDS（默认 60）是每部影片的时长。
set -eu

dir=$(cd "$(dirname "$0")" && pwd)/media/films
secs=${1:-60}
mkdir -p "$dir"
q="-hide_banner -loglevel error -y -nostdin"
enc="-c:v libx264 -preset veryfast -pix_fmt yuv420p -g 50 -keyint_min 50 -sc_threshold 0 -c:a aac -b:a 96k -ac 2"

film() { # name, video source, tone, subtitle, title
	src="-f lavfi -i $2=size=1280x720:rate=25 -f lavfi -i sine=frequency=$3:sample_rate=48000:beep_factor=4"
	ffmpeg $q $src -t "$secs" -vf scale=640:360 $enc -b:v 600k -movflags +faststart "$dir/$1.mp4"
	ffmpeg $q $src -t "$secs" $enc -b:v 1500k -movflags +faststart "$dir/$1.720p.mp4"
	ffmpeg $q -f lavfi -i "$2=size=640x360:rate=25" -frames:v 1 "$dir/$1.jpg"
	printf '1\n00:00:01,000 --> 00:00:04,000\n%s\n\n2\n00:00:06,000 --> 00:00:09,000\nPlayed by fit, as stored.\n' "$4" >"$dir/$1.en.srt"
	# A sidecar names what file names cannot, as the title.
	printf '{"title": "%s"}\n' "$5" >"$dir/$1.json"
}

film testcard testsrc2 440 "A test card." "Test card"
film bars smptehdbars 660 "Colour bars." "Colour bars"
echo "made $(ls "$dir" | wc -l | tr -d ' ') files in $dir"
