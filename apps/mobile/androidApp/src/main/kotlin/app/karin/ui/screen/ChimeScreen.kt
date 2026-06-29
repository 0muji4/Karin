package app.karin.ui.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import app.karin.shared.api.ReceivedCard
import app.karin.shared.ko.KoCatalog
import app.karin.ui.component.VerticalText
import app.karin.ui.deliveries.ChimeViewModel

// 09 風鈴が鳴る。届いた他者の一枚を開いて読む。返信は構造的に存在しない。詳細は縦書きで見せる。
@Composable
fun ChimeScreen(
    card: ReceivedCard,
    state: ChimeViewModel.State,
    onOpened: () -> Unit,
    onKeep: (String) -> Unit,
    onDone: () -> Unit,
    onBack: () -> Unit,
) {
    LaunchedEffect(card.tanzakuId) { onOpened() } // 開封＝端末ローカルに既読を記録
    LaunchedEffect(state) { if (state is ChimeViewModel.State.Kept) onDone() }

    val busy = state is ChimeViewModel.State.Keeping || state is ChimeViewModel.State.Kept

    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        Row(modifier = Modifier.fillMaxWidth()) {
            TextButton(onClick = onBack, enabled = !busy) { Text("‹ 風だより") }
        }
        Text("風鈴が鳴る", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
        Text("ちりん…", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
        if (card.isOfficial) {
            Spacer(Modifier.height(4.dp))
            Text("夏鈴からの一枚", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.primary)
        }

        Box(modifier = Modifier.weight(1f).fillMaxWidth(), contentAlignment = Alignment.Center) {
            VerticalText(text = card.body)
        }

        Text(
            "${KoCatalog.name(card.ko)}　${card.deliveredOn}",
            style = MaterialTheme.typography.labelMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.fillMaxWidth(),
            textAlign = TextAlign.Center,
        )
        Spacer(Modifier.height(8.dp))
        Text(
            "返事はできません。ただ、受け取るだけ。",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.fillMaxWidth(),
            textAlign = TextAlign.Center,
        )

        if (state is ChimeViewModel.State.Error) {
            Spacer(Modifier.height(8.dp))
            Text(state.message, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.primary)
        }
        Spacer(Modifier.height(16.dp))
        Button(onClick = { onKeep(card.tanzakuId) }, enabled = !busy, modifier = Modifier.fillMaxWidth()) {
            Text("文箱にしまう")
        }
    }
}
