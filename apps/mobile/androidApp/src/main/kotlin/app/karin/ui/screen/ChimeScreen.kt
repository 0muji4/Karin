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
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import app.karin.shared.api.ReceivedCard
import app.karin.shared.ko.KoCatalog
import app.karin.ui.component.VerticalText
import app.karin.ui.deliveries.ChimeViewModel

// 通報理由。バックエンドの CHECK と一致するコードに、日本語ラベルを添える。
private val reportReasons = listOf(
    "harassment" to "嫌がらせ・攻撃",
    "sexual" to "性的な内容",
    "self_harm" to "自傷・自殺のおそれ",
    "child_safety" to "児童に関わる",
    "spam" to "スパム",
    "other" to "その他",
)

// 09 風鈴が鳴る。届いた他者の一枚を開いて読む。返信は構造的に存在しない。詳細は縦書きで見せる。
// 匿名の受信物なので、控えめな通報導線を置く（受け手は著者を辿れない）。
@Composable
fun ChimeScreen(
    card: ReceivedCard,
    state: ChimeViewModel.State,
    reportState: ChimeViewModel.ReportState,
    onOpened: () -> Unit,
    onKeep: (String) -> Unit,
    onReport: (String) -> Unit,
    onDone: () -> Unit,
    onBack: () -> Unit,
) {
    LaunchedEffect(card.tanzakuId) { onOpened() } // 開封＝端末ローカルに既読を記録
    LaunchedEffect(state) { if (state is ChimeViewModel.State.Kept) onDone() }

    val busy = state is ChimeViewModel.State.Keeping || state is ChimeViewModel.State.Kept
    val reported = reportState is ChimeViewModel.ReportState.Reported
    var showReport by remember { mutableStateOf(false) }

    if (showReport) {
        ReportDialog(
            onSelect = { reason -> onReport(reason); showReport = false },
            onDismiss = { showReport = false },
        )
    }

    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween) {
            TextButton(onClick = onBack, enabled = !busy) { Text("‹ 風だより") }
            if (reported) {
                Text("通報を受け付けました", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant, modifier = Modifier.padding(12.dp))
            } else {
                TextButton(onClick = { showReport = true }) {
                    Text("通報", color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
            }
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

@Composable
private fun ReportDialog(onSelect: (String) -> Unit, onDismiss: () -> Unit) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("この一枚を通報する") },
        text = {
            Column {
                reportReasons.forEach { (code, label) ->
                    TextButton(onClick = { onSelect(code) }, modifier = Modifier.fillMaxWidth()) {
                        Text(label, modifier = Modifier.fillMaxWidth())
                    }
                }
            }
        },
        confirmButton = {},
        dismissButton = { TextButton(onClick = onDismiss) { Text("やめる") } },
    )
}
