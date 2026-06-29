package app.karin.ui.component

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.unit.dp

// 読む面の縦書き表示（読み取り専用）。Compose に縦書きモードが無いため、文字を縦に積んだ列を
// 右→左に並べて近似する。短い短冊向け。編集は横書き（方針: 編集は横・読む面は縦）。
@Composable
fun VerticalText(
    text: String,
    modifier: Modifier = Modifier,
    charsPerColumn: Int = 14,
    style: TextStyle = MaterialTheme.typography.bodyLarge,
    color: Color = MaterialTheme.colorScheme.onSurface,
) {
    val clean = text.replace("\n", "")
    val columns = if (clean.isEmpty()) listOf("") else clean.chunked(charsPerColumn)
    // 縦書きは右の列から読むため、LTR の Row では列を逆順に並べて先頭列を右端に置く。
    Row(modifier = modifier, horizontalArrangement = Arrangement.spacedBy(6.dp)) {
        columns.asReversed().forEach { column ->
            Column(horizontalAlignment = Alignment.CenterHorizontally) {
                column.forEach { ch ->
                    Text(text = ch.toString(), style = style, color = color)
                }
            }
        }
    }
}
