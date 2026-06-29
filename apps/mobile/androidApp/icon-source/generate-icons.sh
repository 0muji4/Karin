#!/bin/sh
# 夏鈴アプリのランチャーアイコンを元画像（karin_app_icon_1024.png）から再生成する。
# 要 ImageMagick（magick コマンド）。使い方:
#   cd apps/mobile/androidApp && sh icon-source/generate-icons.sh
#
# 設計（なぜこの加工か）:
#   adaptive icon はランチャーが円・角丸など様々な形でマスクし、可視は中央の約 66.6% のみ。
#   この絵は縦長の短冊＋上のハートなので、全面に置くと丸マスクで先端が切れる。そこで:
#   - 前景: 元画像を 80% に縮め上へ 28px 寄せ、構図全体（ハート＋短冊）を安全圏に収める。
#           縁のアルファをぼかして背景へ溶かし、四角い継ぎ目を消す。
#   - 背景: 元画像を全面に敷き強くぼかす（紫陽花の余韻。マスクで外周が切れても成立する）。
set -eu

here=$(CDPATH= cd "$(dirname "$0")" && pwd)
src="$here/karin_app_icon_1024.png"
res="$here/../src/main/res"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# 背景: 全面・強ぼかし・わずかに明るく
magick "$src" -resize 1024x1024^ -gravity center -extent 1024x1024 -blur 0x24 -modulate 104 "$work/bg.png"
# 前景: 80% 縮小・上へ 28px・縁フェザリング
magick -size 1024x1024 xc:none \( "$src" -resize 819x819 \) -gravity center -geometry +0-28 -composite \
  -channel A -blur 0x14 +channel "$work/fg.png"
# 従来型アイコン用の合成（前景 over 背景）
magick "$work/bg.png" "$work/fg.png" -flatten "$work/flat.png"

# 密度ごとに adaptive レイヤー（108dp）と従来型（48dp）を書き出す
for triple in mdpi:108:48 hdpi:162:72 xhdpi:216:96 xxhdpi:324:144 xxxhdpi:432:192; do
  d="${triple%%:*}"; rest="${triple#*:}"; a="${rest%%:*}"; l="${rest##*:}"
  dir="$res/mipmap-$d"; half=$((l / 2)); mkdir -p "$dir"
  magick "$work/fg.png"   -resize "${a}x${a}" -strip -depth 8 "$dir/ic_launcher_foreground.png"
  magick "$work/bg.png"   -resize "${a}x${a}" -strip -depth 8 "$dir/ic_launcher_background.png"
  magick "$work/flat.png" -resize "${l}x${l}" -strip -depth 8 "$dir/ic_launcher.png"
  magick "$work/flat.png" -resize "${l}x${l}" \
    \( -size "${l}x${l}" xc:black -fill white -draw "circle $half,$half $half,0" \) \
    -compose CopyOpacity -composite -strip -depth 8 "$dir/ic_launcher_round.png"
done

echo "アイコンを再生成しました: $res/mipmap-*"
