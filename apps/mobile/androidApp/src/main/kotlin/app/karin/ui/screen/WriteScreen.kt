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
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
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
import androidx.compose.ui.unit.dp
import app.karin.ui.component.TanzakuCard
import app.karin.ui.write.WriteViewModel

// ピル型ボタン（モックの塗り／枠線）。
private val pill = RoundedCornerShape(percent = 50)

// 06 短冊を書く。書いた言葉を短冊として縦書きで見せ（読む面＝縦書き）、編集は横書き（実用性）。
// 文箱にしまう（枠線ピル）／風に乗せる（塗りピル）の二択。
@Composable
fun WriteScreen(
    state: WriteViewModel.State,
    onSave: (String) -> Unit,
    onSaved: () -> Unit,
    onCast: (String) -> Unit,
    onBack: () -> Unit,
) {
    var body by remember { mutableStateOf("") }
    LaunchedEffect(state) {
        if (state is WriteViewModel.State.Saved) onSaved()
    }

    val saving = state is WriteViewModel.State.Saving
    val isEmpty = body.isBlank()

    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Start) {
            TextButton(onClick = onBack, enabled = !saving) { Text("もどる") }
        }
        Text("短冊に書く", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
        Spacer(Modifier.height(16.dp))

        // 書いている言葉を短冊として見せる。空のときは淡い手引きを置く。
        Box(modifier = Modifier.weight(1f).fillMaxWidth(), contentAlignment = Alignment.Center) {
            TanzakuCard(
                text = if (isEmpty) "今日の季節を、ひとこと。" else body,
                textColor = if (isEmpty) MaterialTheme.colorScheme.onSurfaceVariant else MaterialTheme.colorScheme.onSurface,
                modifier = Modifier.fillMaxHeight(0.78f).fillMaxWidth(0.52f),
            )
        }

        Spacer(Modifier.height(12.dp))
        OutlinedTextField(
            value = body,
            onValueChange = { body = it },
            modifier = Modifier.fillMaxWidth(),
            enabled = !saving,
            placeholder = { Text("ここに書く", color = MaterialTheme.colorScheme.onSurfaceVariant) },
            textStyle = MaterialTheme.typography.bodyLarge,
        )

        if (state is WriteViewModel.State.Error) {
            Spacer(Modifier.height(8.dp))
            Text(state.message, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.primary)
        }

        Spacer(Modifier.height(16.dp))
        val canSubmit = body.isNotBlank() && !saving
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            OutlinedButton(onClick = { onSave(body) }, enabled = canSubmit, shape = pill, modifier = Modifier.weight(1f)) {
                Text("文箱にしまう")
            }
            Button(onClick = { onCast(body) }, enabled = canSubmit, shape = pill, modifier = Modifier.weight(1f)) {
                Text("風に乗せる")
            }
        }
    }
}
