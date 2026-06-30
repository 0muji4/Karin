package app.karin.ui.screen

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp
import app.karin.shared.api.ReceivedCard
import app.karin.shared.ko.KoCatalog
import app.karin.ui.deliveries.DeliveriesViewModel

// 10 風だより。未読は上の「届いています」カードから開き、下に受信履歴を並べる（screens/10 に準拠）。
// 未読は端末ローカルの開封済みから導出する。風鈴イラストは後続のため、未読のしるしは小さなドットで代替。
@Composable
fun DeliveriesScreen(
    state: DeliveriesViewModel.State,
    onOpen: (ReceivedCard) -> Unit,
) {
    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        Text("風だより", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
        Text("見知らぬ誰かから、届いた言葉", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
        Spacer(Modifier.height(16.dp))

        when (state) {
            is DeliveriesViewModel.State.Loading -> Centered("…")
            is DeliveriesViewModel.State.Error -> Centered(state.message)
            is DeliveriesViewModel.State.Loaded -> {
                if (state.items.isEmpty()) {
                    Centered("まだ風だよりは届いていません。\n出さなければ、受け取れない。")
                } else {
                    val firstUnread = state.items.firstOrNull { it.unread }
                    if (firstUnread != null) {
                        UnreadCard(count = state.unreadCount, onClick = { onOpen(firstUnread.card) })
                        Spacer(Modifier.height(16.dp))
                    }
                    InfoBox()
                    Spacer(Modifier.height(24.dp))
                    Text("これまでに受け取った言葉", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                    Spacer(Modifier.height(8.dp))
                    LazyColumn(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                        items(state.items.size) { i -> DeliveryRow(state.items[i], onOpen) }
                    }
                }
            }
        }
    }
}

// 未読を開く入口。風鈴イラストの代わりに紺のドットで合図する。
@Composable
private fun UnreadCard(count: Int, onClick: () -> Unit) {
    Surface(
        onClick = onClick,
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(16.dp),
        color = MaterialTheme.colorScheme.surface,
        shadowElevation = 2.dp,
    ) {
        Row(modifier = Modifier.fillMaxWidth().padding(18.dp), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Box(Modifier.size(10.dp).clip(CircleShape).background(MaterialTheme.colorScheme.primary))
                Spacer(Modifier.size(12.dp))
                Column {
                    Text(if (count == 1) "一枚、届いています" else "${count}枚、届いています", style = MaterialTheme.typography.bodyLarge, color = MaterialTheme.colorScheme.primary)
                    Spacer(Modifier.height(2.dp))
                    Text("ひらいて、風鈴を鳴らす", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
                }
            }
            Text("›", style = MaterialTheme.typography.headlineMedium, color = MaterialTheme.colorScheme.primary)
        }
    }
}

// 交換の約束ごとを静かに添える説明。
@Composable
private fun InfoBox() {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(12.dp))
            .background(MaterialTheme.colorScheme.onSurfaceVariant.copy(alpha = 0.08f))
            .padding(16.dp),
    ) {
        Text(
            "出さなければ、受け取れない。\n風だよりは、あなたが一枚を風に乗せたあとに、別の誰かから届きます。",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )
    }
}

@Composable
private fun DeliveryRow(item: DeliveriesViewModel.Item, onOpen: (ReceivedCard) -> Unit) {
    Column(modifier = Modifier.fillMaxWidth().clickable { onOpen(item.card) }.padding(vertical = 10.dp)) {
        Text(item.card.body.replace("\n", " "), style = MaterialTheme.typography.bodyLarge, color = MaterialTheme.colorScheme.onBackground, maxLines = 1)
        Spacer(Modifier.height(2.dp))
        Text("${KoCatalog.name(item.card.ko)} に届いた", style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
    }
}

@Composable
private fun Centered(text: String) {
    Text(text, modifier = Modifier.fillMaxSize().wrapContentSize(Alignment.Center), color = MaterialTheme.colorScheme.onSurfaceVariant)
}
