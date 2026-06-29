package app.karin.ui.theme

import androidx.compose.material3.Typography
import androidx.compose.ui.text.font.Font
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.sp
import app.karin.R

// 明朝基調。Shippori Mincho（OFL）を同梱し、全文字を明朝で組む。システム Serif は端末によって
// 日本語が明朝にならずゴシック化するため、専用フォントを束ねる。季語など常用外の漢字も崩れない
// よう、日本語の常用範囲（CJK統合漢字＋かな＋記号）を網羅したサブセットを使う。
val KarinSerif = FontFamily(Font(R.font.shippori_mincho_regular))

// Material の全ロールに明朝を当てる（ボタン＝labelLarge 等も含め、取りこぼしなく和の書体に）。
// 主要ロールだけ夏鈴用に行間・字送りを整え、他は既定サイズのまま書体だけ差し替える。
private val base = Typography()

val KarinTypography = Typography(
    displayLarge = base.displayLarge.copy(fontFamily = KarinSerif),
    displayMedium = base.displayMedium.copy(fontFamily = KarinSerif),
    displaySmall = base.displaySmall.copy(fontFamily = KarinSerif, fontSize = 28.sp, lineHeight = 40.sp),
    headlineLarge = base.headlineLarge.copy(fontFamily = KarinSerif),
    headlineMedium = base.headlineMedium.copy(fontFamily = KarinSerif, fontSize = 24.sp, lineHeight = 34.sp),
    headlineSmall = base.headlineSmall.copy(fontFamily = KarinSerif),
    titleLarge = base.titleLarge.copy(fontFamily = KarinSerif, fontSize = 20.sp, lineHeight = 30.sp),
    titleMedium = base.titleMedium.copy(fontFamily = KarinSerif),
    titleSmall = base.titleSmall.copy(fontFamily = KarinSerif),
    bodyLarge = base.bodyLarge.copy(fontFamily = KarinSerif, fontSize = 16.sp, lineHeight = 28.sp),
    bodyMedium = base.bodyMedium.copy(fontFamily = KarinSerif, fontSize = 14.sp, lineHeight = 24.sp),
    bodySmall = base.bodySmall.copy(fontFamily = KarinSerif),
    labelLarge = base.labelLarge.copy(fontFamily = KarinSerif),
    labelMedium = base.labelMedium.copy(fontFamily = KarinSerif, fontSize = 12.sp, lineHeight = 18.sp),
    labelSmall = base.labelSmall.copy(fontFamily = KarinSerif),
)
