package app.karin.ui.screen

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import app.karin.ui.today.TodayViewModel

// 05 今日の候。その日の七十二候を便箋のように左寄せで見せ、「短冊を書く」へ導く起点。
@Composable
fun TodayScreen(
    state: TodayViewModel.State,
    onReload: () -> Unit,
    onWrite: () -> Unit,
    onBox: () -> Unit,
    onDeliveries: () -> Unit,
) {
    when (state) {
        is TodayViewModel.State.Loading ->
            CenteredText("…")

        is TodayViewModel.State.Error ->
            Column(
                modifier = Modifier.fillMaxSize().padding(32.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center,
            ) {
                Text(state.message, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onBackground, textAlign = TextAlign.Center)
                Spacer(Modifier.height(16.dp))
                TextButton(onClick = onReload) { Text("もう一度") }
            }

        is TodayViewModel.State.Loaded -> {
            val t = state.today
            val unread = state.unreadCount
            Column(
                modifier = Modifier.fillMaxSize().verticalScroll(rememberScrollState()).padding(horizontal = 28.dp, vertical = 36.dp),
            ) {
                // 暦の行（和風月名 ／ 日付）
                Spacer(Modifier.height(8.dp))
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
                    Text("${t.wafuMonth.name}・${t.wafuMonth.kana}", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Text(t.date, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
                Spacer(Modifier.height(28.dp))

                // 七十二候・第N候 ／ 節気（大）／ 候（紺）
                Text("七十二候・第${t.ko.number}候", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                Spacer(Modifier.height(6.dp))
                Text("${t.sekki.name}（${t.sekki.kana}）", style = MaterialTheme.typography.displaySmall, color = MaterialTheme.colorScheme.onBackground)
                Spacer(Modifier.height(8.dp))
                Text(t.ko.name, style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.primary)
                Text(t.ko.kana, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)

                Spacer(Modifier.height(20.dp))
                HorizontalDivider(modifier = Modifier.width(40.dp), color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.4f))
                Spacer(Modifier.height(20.dp))
                Text(t.ko.meaning, style = MaterialTheme.typography.bodyLarge, color = MaterialTheme.colorScheme.onBackground)

                Spacer(Modifier.height(36.dp))

                // 短冊を書く導線（カード）
                Surface(
                    onClick = onWrite,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(16.dp),
                    color = MaterialTheme.colorScheme.surface,
                    shadowElevation = 2.dp,
                ) {
                    Row(modifier = Modifier.fillMaxWidth().padding(20.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
                        Column {
                            Text("今日の一枚を、短冊に", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onSurface)
                            Spacer(Modifier.height(2.dp))
                            Text("記録は、あなただけのもの", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                        }
                        Text("›", style = MaterialTheme.typography.headlineMedium, color = MaterialTheme.colorScheme.primary)
                    }
                }

                Spacer(Modifier.height(20.dp))
                NavRow(
                    label = if (unread > 0) "風だよりが、${unreadWord(unread)}届いています" else "風だより",
                    onClick = onDeliveries,
                    highlighted = unread > 0,
                )
                HorizontalDivider(color = MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.2f))
                NavRow(label = "文箱・これまでの一枚", onClick = onBox)
            }
        }
    }
}

// ラベル＋しるし「›」の遷移行。未読があれば紺＋ドットで知らせる。
@Composable
private fun NavRow(label: String, onClick: () -> Unit, highlighted: Boolean = false) {
    val accent = MaterialTheme.colorScheme.primary
    Surface(onClick = onClick, color = Color.Transparent, modifier = Modifier.fillMaxWidth()) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(vertical = 14.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                if (highlighted) {
                    Box(Modifier.size(8.dp).clip(CircleShape).background(accent))
                    Spacer(Modifier.width(8.dp))
                }
                Text(label, style = MaterialTheme.typography.bodyLarge, color = if (highlighted) accent else MaterialTheme.colorScheme.onBackground)
            }
            Text("›", style = MaterialTheme.typography.titleLarge, color = if (highlighted) accent else MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
}

// 未読数の言い回し（1 は「ひとつ」、それ以外は「N つ」）。
private fun unreadWord(n: Int): String = if (n == 1) "ひとつ" else "${n}つ"

@Composable
private fun CenteredText(text: String) {
    Text(text, modifier = Modifier.fillMaxSize().wrapContentSize(Alignment.Center), color = MaterialTheme.colorScheme.onSurfaceVariant)
}
