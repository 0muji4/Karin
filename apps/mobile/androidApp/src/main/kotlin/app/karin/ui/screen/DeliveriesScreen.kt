package app.karin.ui.screen

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
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
import androidx.compose.ui.unit.dp
import app.karin.shared.api.ReceivedCard
import app.karin.shared.ko.KoCatalog
import app.karin.ui.deliveries.DeliveriesViewModel

// 10 風だより。届いた一枚への入口と受信履歴。未読は端末ローカルの開封済みから導出する。
// 一覧プレビューは横書き（詳細＝09 で縦書き）。
@Composable
fun DeliveriesScreen(
    state: DeliveriesViewModel.State,
    onOpen: (ReceivedCard) -> Unit,
) {
    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        Text("風だより", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
        Spacer(Modifier.height(12.dp))

        when (state) {
            is DeliveriesViewModel.State.Loading -> Centered("…")
            is DeliveriesViewModel.State.Error -> Centered(state.message)
            is DeliveriesViewModel.State.Loaded -> {
                if (state.items.isEmpty()) {
                    Centered("まだ風だよりは届いていません。\n出さなければ、受け取れない。")
                } else {
                    if (state.unreadCount > 0) {
                        Text(
                            "風だよりが、ひとつ届いています。",
                            style = MaterialTheme.typography.bodyLarge,
                            color = MaterialTheme.colorScheme.primary,
                        )
                        Spacer(Modifier.height(12.dp))
                    }
                    Text("これまでに受け取った言葉", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Spacer(Modifier.height(8.dp))
                    LazyColumn(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        items(state.items.size) { i -> DeliveryRow(state.items[i], onOpen) }
                    }
                }
            }
        }
    }
}

@Composable
private fun DeliveryRow(item: DeliveriesViewModel.Item, onOpen: (ReceivedCard) -> Unit) {
    Row(
        modifier = Modifier.fillMaxWidth().clickable { onOpen(item.card) }.padding(vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            if (item.unread) "●" else "○",
            style = MaterialTheme.typography.labelMedium,
            color = if (item.unread) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurfaceVariant,
            modifier = Modifier.padding(end = 12.dp),
        )
        Column(modifier = Modifier.fillMaxWidth()) {
            Text(item.card.body.replace("\n", " "), style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onBackground, maxLines = 1)
            Text(
                "${KoCatalog.name(item.card.ko)}　${item.card.deliveredOn}",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}

@Composable
private fun Centered(text: String) {
    Text(text, modifier = Modifier.fillMaxSize().wrapContentSize(Alignment.Center), color = MaterialTheme.colorScheme.onSurfaceVariant)
}
