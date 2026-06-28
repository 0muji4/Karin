package app.karin.ui.theme

import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable

// 夏鈴のテーマ。静けさを保つため淡い単色基調・低彩度の差し色のみ。
// ダーク/ダイナミックカラーは意図的に使わない（季節の紙の質感を一定に保つ）。
private val KarinColors = lightColorScheme(
    primary = KarinIndigo,
    onPrimary = KarinSurface,
    secondary = KarinIndigoSoft,
    background = KarinPaper,
    onBackground = KarinInk,
    surface = KarinSurface,
    onSurface = KarinInk,
    surfaceVariant = KarinPaper,
    onSurfaceVariant = KarinMuted,
)

@Composable
fun KarinTheme(content: @Composable () -> Unit) {
    MaterialTheme(
        colorScheme = KarinColors,
        typography = KarinTypography,
        content = content,
    )
}
