package app.karin.ui.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import app.karin.ui.today.TodayViewModel

// 05 今日の候。その日の七十二候を便箋のように見せ、「短冊を書く」へ導く起点。
@Composable
fun TodayScreen(
    state: TodayViewModel.State,
    onReload: () -> Unit,
    onWrite: () -> Unit,
    onBox: () -> Unit,
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
            Column(
                modifier = Modifier.fillMaxSize().verticalScroll(rememberScrollState()).padding(32.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                Spacer(Modifier.height(24.dp))
                Text("${t.date}　${t.wafuMonth.name}（${t.wafuMonth.kana}）", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                Spacer(Modifier.height(20.dp))
                Text("${t.sekki.name}（${t.sekki.kana}）", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.primary)
                Spacer(Modifier.height(8.dp))
                Text(t.ko.name, style = MaterialTheme.typography.displaySmall, color = MaterialTheme.colorScheme.onBackground)
                Text(t.ko.kana, style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                Text("第${t.ko.number}候", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                Spacer(Modifier.height(24.dp))
                Text(t.ko.meaning, style = MaterialTheme.typography.bodyLarge, color = MaterialTheme.colorScheme.onBackground, textAlign = TextAlign.Center)
                Spacer(Modifier.height(40.dp))
                Button(onClick = onWrite, modifier = Modifier.fillMaxWidth()) { Text("今日の一枚を、短冊に") }
                Spacer(Modifier.height(8.dp))
                TextButton(onClick = onBox) { Text("文箱をひらく") }
            }
        }
    }
}

@Composable
private fun CenteredText(text: String) {
    Text(text, modifier = Modifier.fillMaxSize().wrapContentSize(Alignment.Center), color = MaterialTheme.colorScheme.onSurfaceVariant)
}
