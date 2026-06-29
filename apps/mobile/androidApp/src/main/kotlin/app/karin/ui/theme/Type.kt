package app.karin.ui.theme

import androidx.compose.material3.Typography
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.unit.sp

// 明朝（Serif）基調で落ち着いた読み心地に。専用フォント同梱は後続（まずはシステム Serif）。
val KarinTypography = Typography(
    displaySmall = TextStyle(fontFamily = FontFamily.Serif, fontSize = 28.sp, lineHeight = 40.sp),
    headlineMedium = TextStyle(fontFamily = FontFamily.Serif, fontSize = 24.sp, lineHeight = 34.sp),
    titleLarge = TextStyle(fontFamily = FontFamily.Serif, fontSize = 20.sp, lineHeight = 30.sp),
    bodyLarge = TextStyle(fontFamily = FontFamily.Serif, fontSize = 16.sp, lineHeight = 28.sp),
    bodyMedium = TextStyle(fontFamily = FontFamily.Serif, fontSize = 14.sp, lineHeight = 24.sp),
    labelMedium = TextStyle(fontFamily = FontFamily.Serif, fontSize = 12.sp, lineHeight = 18.sp),
)
