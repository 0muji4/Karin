# 同梱フォント（Shippori Mincho）

夏鈴は全文を明朝で組む。システムの Serif は端末によって日本語がゴシック化するため、
OFL の明朝を同梱する。

- 書体: **Shippori Mincho**（ふぉんとうれしい / The Shippori Mincho Project, SIL OFL 1.1）
- 配置: `src/main/res/font/shippori_mincho_regular.ttf`
- ライセンス本文: `src/main/assets/licenses/ShipporiMincho-OFL.txt`（OFL は配布物への license 同梱を要求）

## 出所と再生成

元ファイルは google/fonts（OFL）から取得し、日本語の常用範囲に絞って同梱サイズを下げている
（B1 版は 15MB、素の Regular でも 8.7MB。常用漢字を間引くと季語の常用外漢字が脱落するため、
CJK 統合漢字は全域残す）。要 `hb-subset`（HarfBuzz）。

```sh
curl -sL -o ShipporiMincho-Regular.ttf \
  https://github.com/google/fonts/raw/main/ofl/shipporimincho/ShipporiMincho-Regular.ttf

hb-subset ShipporiMincho-Regular.ttf \
  --unicodes="U+0020-007E,U+00A0-00FF,U+2010-2027,U+2030-205E,U+2190-21FF,U+2460-24FF,U+25A0-26FF,U+3000-303F,U+3040-309F,U+30A0-30FF,U+3190-31FF,U+3220-32FF,U+4E00-9FFF,U+F900-FAFF,U+FF00-FFEF" \
  --output-file=../src/main/res/font/shippori_mincho_regular.ttf
```

字形は B1 版と同一で、標準日本語を網羅する（B1 限定の稀少字のみ非対応）。太字が要るロールは
今は同梱せず、見出しは Regular を大きく組んで対応している（重さが要れば別ウェイトを追加する）。
