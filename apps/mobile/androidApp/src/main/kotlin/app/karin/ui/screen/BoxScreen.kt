package app.karin.ui.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
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

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun SekkiGroup(group: BoxGroupDto) {
    Column {
        Text(
            "${group.wafuMonth.name}・${group.sekki.name}",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
        Spacer(Modifier.height(12.dp))
        // 細い短冊を左から並べ、入りきらなければ折り返す（短冊なので幅は固定し、横には広げない）。
        FlowRow(
            horizontalArrangement = Arrangement.spacedBy(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
        ) {
            group.records.forEach { TanzakuEntry(it) }
        }
    }
}

// 短冊一枚＋日付・候名。短冊は細長い札なので幅を固定（縦長比）にして、行いっぱいには広げない。
@Composable
private fun TanzakuEntry(record: RecordDto) {
    Column(horizontalAlignment = Alignment.CenterHorizontally, modifier = Modifier.width(96.dp)) {
        TanzakuCard(
            text = record.body,
            modifier = Modifier.fillMaxWidth().height(236.dp),
            textStyle = MaterialTheme.typography.bodyMedium,
            topBandHeight = 4.dp,
            showHole = false,
            elevation = 1.dp,
        )
        Spacer(Modifier.height(6.dp))
        Text(
            entryLabel(record),
            style = MaterialTheme.typography.labelMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
    }
}

// 「6.29　菖蒲華」のように 日付＋候名。日付が解釈できなければ候名だけ。
private fun entryLabel(record: RecordDto): String {
    val ko = KoCatalog.name(record.koWritten)
    val date = runCatching {
        val (_, m, d) = record.createdAt.take(10).split("-")
        "${m.toInt()}.${d.toInt()}"
    }.getOrNull()
    return if (date != null) "$date　$ko" else ko
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
