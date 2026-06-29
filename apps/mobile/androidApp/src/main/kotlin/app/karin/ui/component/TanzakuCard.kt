package app.karin.ui.component

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp

// 短冊。白い紙に紺の帯と紐穴、本文は縦書き。季節の言葉を「一枚」として見せる夏鈴の中心モチーフ
// （記録・受信・文箱で共有する）。高さは呼び手が与える（縦書きは有界の高さから列数を決めるため）。
@Composable
fun TanzakuCard(
    text: String,
    modifier: Modifier = Modifier,
    textColor: Color = MaterialTheme.colorScheme.onSurface,
) {
    Surface(
        modifier = modifier,
        shape = RoundedCornerShape(8.dp),
        color = MaterialTheme.colorScheme.surface,
        shadowElevation = 6.dp,
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            // 紺の帯（短冊の頭）
            Box(Modifier.fillMaxWidth().height(24.dp).background(MaterialTheme.colorScheme.primary))
            // 紐を通す穴
            Box(
                Modifier
                    .padding(top = 12.dp)
                    .size(10.dp)
                    .clip(CircleShape)
                    .background(MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.35f)),
            )
            // 本文（縦書き）。残りの高さいっぱいに、中央へ。
            Box(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .padding(horizontal = 18.dp, vertical = 18.dp),
                contentAlignment = Alignment.Center,
            ) {
                VerticalText(
                    text = text,
                    style = MaterialTheme.typography.titleLarge,
                    color = textColor,
                )
            }
        }
    }
}
