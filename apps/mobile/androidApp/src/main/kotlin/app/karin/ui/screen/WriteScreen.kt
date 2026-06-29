package app.karin.ui.screen

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
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
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import app.karin.ui.write.WriteViewModel

// 06 短冊を書く。記録 MVP では「文箱にしまう」のみ（風に乗せるは交換系で後続）。
// 縦書きは読む面（文箱・受信）に用い、編集は実用性のため横書きにする（方針: 編集は横・読む面は縦）。
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

    Column(modifier = Modifier.fillMaxSize().padding(24.dp)) {
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.Start) {
            TextButton(onClick = onBack, enabled = !saving) { Text("もどる") }
        }
        Text("短冊を書く", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.onBackground)
        Spacer(Modifier.height(20.dp))

        OutlinedTextField(
            value = body,
            onValueChange = { body = it },
            modifier = Modifier.fillMaxWidth().weight(1f),
            enabled = !saving,
            placeholder = { Text("今日の季節を、ひとこと。", color = MaterialTheme.colorScheme.onSurfaceVariant) },
            textStyle = MaterialTheme.typography.bodyLarge,
        )

        if (state is WriteViewModel.State.Error) {
            Spacer(Modifier.height(8.dp))
            Text(state.message, style = MaterialTheme.typography.labelMedium, color = MaterialTheme.colorScheme.primary)
        }

        Spacer(Modifier.height(16.dp))
        val canSubmit = body.isNotBlank() && !saving
        Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            OutlinedButton(onClick = { onSave(body) }, enabled = canSubmit, modifier = Modifier.weight(1f)) {
                Text("文箱にしまう")
            }
            Button(onClick = { onCast(body) }, enabled = canSubmit, modifier = Modifier.weight(1f)) {
                Text("風に乗せる")
            }
        }
    }
}
