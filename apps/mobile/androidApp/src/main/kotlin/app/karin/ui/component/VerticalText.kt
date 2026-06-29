package app.karin.ui.component

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.unit.dp

// verticalColumns は本文を縦書きの「列」に割る純関数（描画と分離してテスト可能にする）。
// 縦書きでは改行が次の列の始まりなので \n をまず列の区切りにし、各行が1列に収まらない分は
// maxCharsPerColumn ごとにさらに列へ送る。返り値は読む順（先頭が最初の列＝最も右に置く列）。
internal fun verticalColumns(text: String, maxCharsPerColumn: Int): List<String> {
    if (maxCharsPerColumn < 1) return listOf(text)
    return text.split("\n").flatMap { line ->
        if (line.isEmpty()) listOf("") else line.chunked(maxCharsPerColumn)
    }
}

// 読む面の縦書き表示（読み取り専用）。Compose に縦書きモードが無いため、文字を縦に積んだ列を
// 右→左に並べて近似する。1列に収める文字数は利用可能な高さから算出し、札からはみ出して
// クリップしないようにする（高さは呼び出し側が制約として与える前提）。
@Composable
fun VerticalText(
    text: String,
    modifier: Modifier = Modifier,
    style: TextStyle = MaterialTheme.typography.bodyLarge,
    color: Color = MaterialTheme.colorScheme.onSurface,
) {
    BoxWithConstraints(modifier) {
        val lineHeightDp = with(LocalDensity.current) { style.lineHeight.toDp() }
        // 高さが有界なら収まる文字数を求める。無界（測れない）時は控えめな既定にする。
        val perColumn = if (maxHeight.value.isFinite() && lineHeightDp.value > 0f) {
            (maxHeight.value / lineHeightDp.value).toInt().coerceAtLeast(1)
        } else {
            8
        }
        val columns = verticalColumns(text, perColumn)
        // 縦書きは右の列から読む。LTR の Row では列を逆順に置いて先頭列を右端にする。
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            columns.asReversed().forEach { column ->
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    column.forEach { ch ->
                        Text(text = ch.toString(), style = style, color = color)
                    }
                }
            }
        }
    }
}
