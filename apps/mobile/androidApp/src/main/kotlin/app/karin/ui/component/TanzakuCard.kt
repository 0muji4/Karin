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
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

// 短冊。白い紙に色帯と（任意で）紐穴、本文は縦書き。夏鈴の中心モチーフを記録・受信・文箱で共有する。
// 受信(09)の大きな札は太い帯＋紐穴、文箱(11)の小さな札は細い線・紐穴なし、と頭の太さ・紐穴・字の
// 大きさで使い分ける（高さは呼び手が与える：縦書きは有界の高さから列数を決めるため）。
@Composable
fun TanzakuCard(
    text: String,
    modifier: Modifier = Modifier,
    textColor: Color = MaterialTheme.colorScheme.onSurface,
    textStyle: TextStyle = MaterialTheme.typography.titleLarge,
    topBandHeight: Dp = 24.dp,
    showHole: Boolean = true,
    elevation: Dp = 6.dp,
) {
    Surface(
        modifier = modifier,
        shape = RoundedCornerShape(8.dp),
        color = MaterialTheme.colorScheme.surface,
        shadowElevation = elevation,
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally) {
            // 頭の色帯。太いと札の頭、細いと上端の線になる。
            Box(Modifier.fillMaxWidth().height(topBandHeight).background(MaterialTheme.colorScheme.primary))
            if (showHole) {
                // 紐を通す穴
                Box(
                    Modifier
                        .padding(top = 12.dp)
                        .size(10.dp)
                        .clip(CircleShape)
                        .background(MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.35f)),
                )
            }
            // 本文（縦書き）。残りの高さいっぱいに、中央へ。
            Box(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .padding(horizontal = 14.dp, vertical = 14.dp),
                contentAlignment = Alignment.Center,
            ) {
                VerticalText(text = text, style = textStyle, color = textColor)
            }
        }
    }
}
