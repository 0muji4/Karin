package app.karin.nav

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import app.karin.di.AppContainer
import app.karin.shared.api.ReceivedCard
import app.karin.ui.account.AccountViewModel
import app.karin.ui.box.BoxViewModel
import app.karin.ui.cast.CastViewModel
import app.karin.ui.deliveries.ChimeViewModel
import app.karin.ui.deliveries.DeliveriesViewModel
import app.karin.ui.screen.BoxScreen
import app.karin.ui.screen.CastScreen
import app.karin.ui.screen.ChimeScreen
import app.karin.ui.screen.DeliveriesScreen
import app.karin.ui.screen.OnboardingScreen
import app.karin.ui.screen.SentScreen
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
                onDeliveries = { nav.navigate(Routes.DELIVERIES) },
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
                onCast = { body ->
                    // 下書き本文を次画面へ渡す（URL エンコードを避け savedStateHandle 経由）。
                    nav.currentBackStackEntry?.savedStateHandle?.set("castBody", body)
                    nav.navigate(Routes.CAST)
                },
                onBack = { nav.popBackStack() },
            )
        }
        composable(Routes.CAST) {
            val vm: CastViewModel = viewModel(
                factory = viewModelFactory { initializer { CastViewModel(container.repository) } },
            )
            val state by vm.state.collectAsStateWithLifecycle()
            val body = remember { nav.previousBackStackEntry?.savedStateHandle?.get<String>("castBody").orEmpty() }
            CastScreen(
                body = body,
                state = state,
                onConfirm = { vm.cast(body) },
                onSent = { nav.navigate(Routes.SENT) { popUpTo(Routes.HOME) { inclusive = false } } },
                onHome = { nav.popBackStack(Routes.HOME, inclusive = false) },
                onBack = { nav.popBackStack() },
            )
        }
        composable(Routes.SENT) {
            SentScreen(onHome = { nav.popBackStack(Routes.HOME, inclusive = false) })
        }
        composable(Routes.DELIVERIES) {
            val vm: DeliveriesViewModel = viewModel(
                factory = viewModelFactory { initializer { DeliveriesViewModel(container.repository, container.readState) } },
            )
            val state by vm.state.collectAsStateWithLifecycle()
            DeliveriesScreen(
                state = state,
                onOpen = { card ->
                    nav.currentBackStackEntry?.savedStateHandle?.set("openedCard", card)
                    nav.navigate(Routes.CHIME)
                },
                onBack = { nav.popBackStack() },
            )
        }
        composable(Routes.CHIME) {
            val vm: ChimeViewModel = viewModel(
                factory = viewModelFactory { initializer { ChimeViewModel(container.repository) } },
            )
            val state by vm.state.collectAsStateWithLifecycle()
            val reportState by vm.reportState.collectAsStateWithLifecycle()
            val card = remember { nav.previousBackStackEntry?.savedStateHandle?.get<ReceivedCard>("openedCard") }
            if (card == null) {
                LaunchedEffect(Unit) { nav.popBackStack() }
            } else {
                ChimeScreen(
                    card = card,
                    state = state,
                    reportState = reportState,
                    onOpened = { container.readState.markOpened(card.tanzakuId) },
                    onKeep = vm::keep,
                    onReport = { reason -> vm.report(card.tanzakuId, reason) },
                    onDone = { nav.popBackStack(Routes.DELIVERIES, inclusive = false) },
                    onBack = { nav.popBackStack() },
                )
            }
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
