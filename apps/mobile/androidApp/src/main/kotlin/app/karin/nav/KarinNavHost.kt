package app.karin.nav

import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import app.karin.di.AppContainer
import app.karin.ui.account.AccountViewModel
import app.karin.ui.screen.OnboardingScreen
import app.karin.ui.screen.SplashScreen

// アプリのナビゲーション。起動→（在席なら本編／初回ならオンボ）。オンボ完了でアカウントを発行し本編へ。
@Composable
fun KarinNavHost(container: AppContainer) {
    val nav = rememberNavController()
    NavHost(navController = nav, startDestination = Routes.SPLASH) {
        composable(Routes.SPLASH) {
            SplashScreen(onTap = {
                val dest = if (container.sessionRepository.isSignedIn()) Routes.HOME else Routes.ONBOARDING
                nav.navigate(dest) { popUpTo(Routes.SPLASH) { inclusive = true } }
            })
        }
        composable(Routes.ONBOARDING) {
            val vm: AccountViewModel = viewModel(
                factory = viewModelFactory { initializer { AccountViewModel(container.sessionRepository) } },
            )
            val state by vm.state.collectAsStateWithLifecycle()
            LaunchedEffect(state) {
                if (state is AccountViewModel.State.Done) {
                    nav.navigate(Routes.HOME) { popUpTo(Routes.ONBOARDING) { inclusive = true } }
                }
            }
            OnboardingScreen(
                onStart = vm::issue,
                busy = state is AccountViewModel.State.Loading,
                error = (state as? AccountViewModel.State.Error)?.message,
            )
        }
        composable(Routes.HOME) { Placeholder("今日の候（準備中）") }
    }
}

@Composable
private fun Placeholder(label: String) {
    Text(text = label, modifier = Modifier.fillMaxSize().wrapContentSize(Alignment.Center))
}
