package app.karin.nav

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import app.karin.di.AppContainer
import app.karin.ui.account.AccountViewModel
import app.karin.ui.box.BoxViewModel
import app.karin.ui.screen.BoxScreen
import app.karin.ui.screen.OnboardingScreen
import app.karin.ui.screen.SplashScreen
import app.karin.ui.screen.TodayScreen
import app.karin.ui.screen.WriteScreen
import app.karin.ui.today.TodayViewModel
import app.karin.ui.write.WriteViewModel

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
        composable(Routes.HOME) {
            val vm: TodayViewModel = viewModel(
                factory = viewModelFactory { initializer { TodayViewModel(container.repository) } },
            )
            val state by vm.state.collectAsStateWithLifecycle()
            TodayScreen(
                state = state,
                onReload = vm::load,
                onWrite = { nav.navigate(Routes.WRITE) },
                onBox = { nav.navigate(Routes.BOX) },
            )
        }
        composable(Routes.WRITE) {
            val vm: WriteViewModel = viewModel(
                factory = viewModelFactory { initializer { WriteViewModel(container.repository) } },
            )
            val state by vm.state.collectAsStateWithLifecycle()
            WriteScreen(
                state = state,
                onSave = vm::save,
                onSaved = { nav.popBackStack(Routes.HOME, inclusive = false) },
                onBack = { nav.popBackStack() },
            )
        }
        composable(Routes.BOX) {
            val vm: BoxViewModel = viewModel(
                factory = viewModelFactory { initializer { BoxViewModel(container.repository) } },
            )
            val state by vm.state.collectAsStateWithLifecycle()
            BoxScreen(state = state, onBack = { nav.popBackStack() })
        }
    }
}
