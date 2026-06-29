package app.karin.nav

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import app.karin.di.AppContainer

// アプリのナビゲーション。起動時、トークンの有無で初回（オンボーディング）か本編（今日の候）かを分ける。
// 各画面の中身は後続 PR で差し込む（本 PR は骨格と遷移の土台）。
@Composable
fun KarinNavHost(container: AppContainer) {
    val nav = rememberNavController()
    val start = if (container.hasAccount()) Routes.HOME else Routes.ONBOARDING
    NavHost(navController = nav, startDestination = start) {
        composable(Routes.ONBOARDING) { Placeholder("オンボーディング（準備中）") }
        composable(Routes.HOME) { Placeholder("今日の候（準備中）") }
    }
}

@Composable
private fun Placeholder(label: String) {
    Text(text = label, modifier = Modifier.fillMaxSize().wrapContentSize(Alignment.Center))
}
