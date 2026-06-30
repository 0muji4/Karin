package app.karin.ui.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.launch

private data class OnbPage(val symbol: String, val label: String, val heading: String, val body: String)

// 02-04 オンボーディング。短冊（書く）→風（送る・相互性）→鈴（受けとる・無返信）の順で世界観を伝える。
// 文言・体裁は screens/02-04 のモックに合わせる。
private val onbPages = listOf(
    OnbPage("短", "たんざく ・ 記録", "書く", "今日の季節を、短い言葉で。\nだれにも見せない、\nあなただけの記録です。"),
    OnbPage("風", "かぜ ・ 交換", "送る", "気が向いたら、一枚だけ風に乗せる。\n出さなければ、受け取れません。"),
    OnbPage("鈴", "すず ・ 受信", "受けとる", "数日のうちに、別の誰かの一枚が届く。\n返事はいらない。\nただ、受け取るだけ。"),
)

// つぎへ／夏鈴をはじめる は中央の塗りピル（全幅にはしない）。
private val pill = RoundedCornerShape(percent = 50)

@Composable
fun OnboardingScreen(
    onStart: () -> Unit,
    busy: Boolean,
    error: String?,
) {
    val pager = rememberPagerState { onbPages.size }
    val scope = rememberCoroutineScope()
    val isLast = pager.currentPage == onbPages.lastIndex

    Column(modifier = Modifier.fillMaxSize().padding(32.dp)) {
        HorizontalPager(state = pager, modifier = Modifier.weight(1f)) { index ->
            val page = onbPages[index]
            Column(
                modifier = Modifier.fillMaxSize(),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.Center,
            ) {
                Text(page.symbol, style = MaterialTheme.typography.displayMedium, color = MaterialTheme.colorScheme.primary)
                Spacer(Modifier.height(12.dp))
                Text(page.label, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                Spacer(Modifier.height(24.dp))
                Text(page.heading, style = MaterialTheme.typography.headlineMedium, color = MaterialTheme.colorScheme.onBackground)
                Spacer(Modifier.height(16.dp))
                Text(
                    page.body,
                    style = MaterialTheme.typography.bodyLarge,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    textAlign = TextAlign.Center,
                )
            }
        }

        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Center) {
            onbPages.indices.forEach { i ->
                Text(
                    if (i == pager.currentPage) "●" else "○",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.padding(horizontal = 4.dp),
                )
            }
        }
        Spacer(Modifier.height(20.dp))

        if (error != null) {
            Text(error, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.primary)
            Spacer(Modifier.height(8.dp))
        }

        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Center) {
            Button(
                onClick = {
                    if (isLast) onStart() else scope.launch { pager.animateScrollToPage(pager.currentPage + 1) }
                },
                enabled = !busy,
                shape = pill,
                modifier = Modifier.widthIn(min = 160.dp),
            ) {
                Text(if (isLast) "夏鈴をはじめる" else "つぎへ")
            }
        }
    }
}
