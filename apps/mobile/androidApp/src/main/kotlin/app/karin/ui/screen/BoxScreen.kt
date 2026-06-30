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
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import app.karin.shared.api.BoxGroupDto
import app.karin.shared.api.RecordDto
import app.karin.shared.ko.KoCatalog
import app.karin.ui.box.BoxViewModel
import app.karin.ui.component.TanzakuCard

// 11 文箱。自分が詠んだ短冊を暦（和風月名・節気）でまとめて振り返るアーカイブ。
// 短冊は読む面なので共有の TanzakuCard（紺帯・紐穴）で縦書きに見せ、2列の格子に並べる。
@Composable
fun BoxScreen(state: BoxViewModel.State) {
    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        when (state) {
            is BoxViewModel.State.Loading -> {
                Text("文箱", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
                Centered("…")
            }
            is BoxViewModel.State.Error -> {
                Text("文箱", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
                Centered(state.message)
            }
            is BoxViewModel.State.Loaded -> {
                Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.Bottom) {
                    Column {
                        Text("文箱", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
                        Text("あなたの季節の記録", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                    if (state.total > 0) {
                        Text("${state.total} まい", style = MaterialTheme.typography.headlineMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    }
                }
                Spacer(Modifier.height(16.dp))
                if (state.total == 0) {
                    Centered("まだ一枚もありません。\n今日の一枚を、短冊に。")
                } else {
                    LazyColumn(verticalArrangement = Arrangement.spacedBy(24.dp)) {
                        items(state.box.groups.size) { i -> SekkiGroup(state.box.groups[i]) }
                    }
                }
            }
        }
    }
}

@Composable
private fun SekkiGroup(group: BoxGroupDto) {
    Column {
        Text(
            "${group.wafuMonth.name}・${group.sekki.name}",
            style = MaterialTheme.typography.titleLarge,
            color = MaterialTheme.colorScheme.primary,
        )
        Spacer(Modifier.height(12.dp))
        // 2列の格子。奇数の余りは空きで埋めて左寄せを保つ。
        group.records.chunked(2).forEach { pair ->
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                pair.forEach { record ->
                    Box(Modifier.weight(1f)) { TanzakuEntry(record) }
                }
                if (pair.size == 1) Spacer(Modifier.weight(1f))
            }
            Spacer(Modifier.height(16.dp))
        }
    }
}

// 短冊一枚＋候名。
@Composable
private fun TanzakuEntry(record: RecordDto) {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        TanzakuCard(text = record.body, modifier = Modifier.fillMaxWidth().height(220.dp))
        Spacer(Modifier.height(6.dp))
        Text(
            KoCatalog.name(record.koWritten),
            style = MaterialTheme.typography.labelMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
    }
}

@Composable
private fun Centered(text: String) {
    Text(
        text,
        modifier = Modifier.fillMaxSize().wrapContentSize(Alignment.Center),
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        textAlign = TextAlign.Center,
    )
}
