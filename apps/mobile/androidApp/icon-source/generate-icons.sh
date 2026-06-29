#!/bin/sh
# 夏鈴アプリのランチャーアイコンを元画像（karin_app_icon_1024.png）から再生成する。
# 要 ImageMagick（magick コマンド）。使い方:
#   cd apps/mobile/androidApp && sh icon-source/generate-icons.sh
#
# 設計（なぜこの加工か）:
#   adaptive icon はランチャーが円・角丸など様々な形でマスクし、可視は中央の約 66.6% のみ。
#   この絵は縦長の短冊＋上のハートなので、全面に置くと丸マスクで先端が切れる。一方、
#   背景をぼかすと元デザインの「くっきりした紫陽花」から遠ざかってしまう。そこで:
#   - 構図はくっきりのまま 84% に収め、ハートと短冊が安全圏に入るよう上へ少し寄せる。
#   - 生まれる余白は元画像の地色（ハートより上＝背景の平均色）で埋める。ぼかさず・反射させない。
#     地色なので継ぎ目は目立たず、余白の大半は丸マスクで隠れる。縁は 4px だけ馴染ませる。
set -eu

here=$(CDPATH= cd "$(dirname "$0")" && pwd)
src="$here/karin_app_icon_1024.png"
res="$here/../src/main/res"
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

# 余白のフィル色 = 上端 110px（背景）の平均色
fill=$(magick "$src" -crop 1024x110+0+0 +repage -resize 1x1! -format '%[pixel:p{0,0}]' info:)
# くっきりの構図を 84% に収め、上へ 24px 寄せ、地色で埋める（縁 4px だけ馴染ませる）
magick -size 1024x1024 xc:"$fill" \
  \( "$src" -resize 860x860 -background none -gravity center -extent 880x880 -channel A -blur 0x4 +channel \) \
  -gravity center -geometry +0-24 -compose over -composite -strip -depth 8 "$work/flat.png"

# 密度ごとに adaptive レイヤー（108dp）と従来型（48dp）を書き出す。
# 前景=背景=同一の合成画像（不透明）にすることで継ぎ目もゴーストも生まれない。マスクが外周を切る。
for triple in mdpi:108:48 hdpi:162:72 xhdpi:216:96 xxhdpi:324:144 xxxhdpi:432:192; do
  d="${triple%%:*}"; rest="${triple#*:}"; a="${rest%%:*}"; l="${rest##*:}"
  dir="$res/mipmap-$d"; half=$((l / 2)); mkdir -p "$dir"
  magick "$work/flat.png" -resize "${a}x${a}" -strip -depth 8 "$dir/ic_launcher_foreground.png"
  magick "$work/flat.png" -resize "${a}x${a}" -strip -depth 8 "$dir/ic_launcher_background.png"
  magick "$work/flat.png" -resize "${l}x${l}" -strip -depth 8 "$dir/ic_launcher.png"
  magick "$work/flat.png" -resize "${l}x${l}" \
    \( -size "${l}x${l}" xc:none -fill white -draw "circle $half,$half $half,0" \) \
    -compose DstIn -composite -strip -depth 8 "$dir/ic_launcher_round.png"
done

echo "アイコンを再生成しました: $res/mipmap-*"
