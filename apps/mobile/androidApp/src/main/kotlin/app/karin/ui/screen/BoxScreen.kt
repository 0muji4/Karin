package app.karin.ui.screen

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import app.karin.shared.api.BoxGroupDto
import app.karin.shared.api.RecordDto
import app.karin.shared.ko.KoCatalog
import app.karin.ui.box.BoxViewModel
import app.karin.ui.component.VerticalText

// 11 文箱。自分が詠んだ短冊を暦（和風月名・節気）でまとめて振り返るアーカイブ。
// 短冊は読む面なので縦書きで見せる。候名は記録の ko_written からカタログで解決する。
@Composable
fun BoxScreen(state: BoxViewModel.State, onBack: () -> Unit) {
    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Start) {
            TextButton(onClick = onBack) { Text("もどる") }
        }
        Text("文箱", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
        Spacer(Modifier.height(12.dp))

        when (state) {
            is BoxViewModel.State.Loading -> Centered("…")
            is BoxViewModel.State.Error -> Centered(state.message)
            is BoxViewModel.State.Loaded -> {
                if (state.total == 0) {
                    Centered("まだ一枚もありません。\n今日の一枚を、短冊に。")
                } else {
                    Text("${state.total}まい", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Spacer(Modifier.height(8.dp))
                    LazyColumn(verticalArrangement = Arrangement.spacedBy(20.dp)) {
                        items(state.box.groups.size) { i ->
                            SekkiGroup(state.box.groups[i])
                        }
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
        Spacer(Modifier.height(8.dp))
        LazyRow(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            items(group.records.size) { i ->
                TanzakuCard(group.records[i])
            }
        }
    }
}

@Composable
private fun TanzakuCard(record: RecordDto) {
    // 短冊らしい縦長の細い札。幅は列数に応じて伸び、高さは固定して列がはみ出さないようにする。
    Surface(
        color = MaterialTheme.colorScheme.surface,
        shape = RoundedCornerShape(3.dp),
        border = BorderStroke(1.dp, MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.3f)),
        modifier = Modifier.widthIn(min = 56.dp).height(300.dp),
    ) {
        Column(modifier = Modifier.padding(horizontal = 12.dp, vertical = 16.dp)) {
            VerticalText(text = record.body, modifier = Modifier.weight(1f))
            Spacer(Modifier.height(8.dp))
            Text(
                KoCatalog.name(record.koWritten),
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

@Composable
private fun Centered(text: String) {
    Text(
        text,
        modifier = Modifier.fillMaxSize().wrapContentSize(Alignment.Center),
        color = MaterialTheme.colorScheme.onSurfaceVariant,
    )
}
