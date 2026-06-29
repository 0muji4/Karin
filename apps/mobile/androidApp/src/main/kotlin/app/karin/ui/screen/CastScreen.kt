package app.karin.ui.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import app.karin.ui.cast.CastViewModel
import app.karin.ui.component.TanzakuCard

// ピル型ボタン。
private val pill = RoundedCornerShape(percent = 50)

// 07 風に乗せる。送る前の最終確認。経路と相互性を伝え、確定で公開プールへ流す。
// 判定は著者に見せないため、成功（support なし）は黙って 08 へ。危機（support あり）のときだけ
// この画面で支援先を案内する。
@Composable
fun CastScreen(
    body: String,
    state: CastViewModel.State,
    onConfirm: () -> Unit,
    onSent: () -> Unit,
    onHome: () -> Unit,
    onBack: () -> Unit,
) {
    LaunchedEffect(state) {
        if (state is CastViewModel.State.Sent && state.support == null) onSent()
    }

    val support = (state as? CastViewModel.State.Sent)?.support
    val casting = state is CastViewModel.State.Casting

    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        Row(modifier = Modifier.fillMaxWidth()) {
            TextButton(onClick = onBack, enabled = !casting) { Text("もどる") }
        }
        Text("風に乗せる", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
        Spacer(Modifier.height(16.dp))

        if (support != null) {
            // 危機と判定したときの支援先案内（本人にだけ）。
            Column(modifier = Modifier.weight(1f).fillMaxWidth(), verticalArrangement = Arrangement.Center, horizontalAlignment = Alignment.CenterHorizontally) {
                Text(support.message, style = MaterialTheme.typography.bodyLarge, color = MaterialTheme.colorScheme.onBackground, textAlign = TextAlign.Center)
                Spacer(Modifier.height(16.dp))
                val uriHandler = LocalUriHandler.current
                TextButton(onClick = { uriHandler.openUri(support.url) }) { Text("相談できる窓口をひらく") }
            }
            Button(onClick = onHome, shape = pill, modifier = Modifier.fillMaxWidth()) { Text("今日へもどる") }
        } else {
            Box(modifier = Modifier.weight(1f).fillMaxWidth(), contentAlignment = Alignment.Center) {
                TanzakuCard(
                    text = body,
                    modifier = Modifier.fillMaxHeight(0.72f).fillMaxWidth(0.52f),
                )
            }
            Spacer(Modifier.height(16.dp))
            Text(
                "あなたの一枚が、数日のうちに別の誰かへ。\n出さなければ、受け取れない。",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                textAlign = TextAlign.Center,
                modifier = Modifier.fillMaxWidth(),
            )
            if (state is CastViewModel.State.Error) {
                Spacer(Modifier.height(8.dp))
                Text(state.message, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.primary)
            }
            Spacer(Modifier.height(16.dp))
            Button(onClick = onConfirm, enabled = !casting, shape = pill, modifier = Modifier.fillMaxWidth()) {
                Text("この一枚を、風に乗せる")
            }
        }
    }
}
