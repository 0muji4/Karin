package app.karin.nav

import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.NavigationBar
import androidx.compose.material3.NavigationBarItem
import androidx.compose.material3.NavigationBarItemDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LifecycleEventEffect
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.lifecycle.viewmodel.initializer
import androidx.lifecycle.viewmodel.viewModelFactory
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.currentBackStackEntryAsState
import androidx.navigation.compose.rememberNavController
import app.karin.di.AppContainer
import app.karin.shared.api.decodeReceivedCard
import app.karin.shared.api.encode
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

// 常時表示するタブ（今＝今日の候／風＝風だより／文＝文箱）。書く・送信・風鈴が鳴るは上に積む。
private val tabRoutes = setOf(Routes.HOME, Routes.DELIVERIES, Routes.BOX)

// アプリのナビゲーション。起動→（在席なら本編／初回ならオンボ）。本編は3タブを下部に常設する。
@Composable
fun KarinNavHost(container: AppContainer) {
    val nav = rememberNavController()
    val backEntry by nav.currentBackStackEntryAsState()
    val route = backEntry?.destination?.route

    // タブ切替：各タブの状態を保持しつつ単一インスタンスで行き来する。
    val selectTab: (String) -> Unit = { dest ->
        nav.navigate(dest) {
            popUpTo(Routes.HOME) { saveState = true }
            launchSingleTop = true
            restoreState = true
        }
    }

    Scaffold(
        bottomBar = {
            if (route in tabRoutes) KarinBottomBar(currentRoute = route, onSelect = selectTab)
        },
    ) { inner ->
        NavHost(navController = nav, startDestination = Routes.SPLASH, modifier = Modifier.padding(inner)) {
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
                    factory = viewModelFactory { initializer { TodayViewModel(container.repository, container.readState) } },
                )
                val state by vm.state.collectAsStateWithLifecycle()
                // 戻るたびに読み直し、風だよりの未読数を最新化する。
                LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { vm.load() }
                TodayScreen(
                    state = state,
                    onReload = vm::load,
                    onWrite = { nav.navigate(Routes.WRITE) },
                    onBox = { selectTab(Routes.BOX) },
                    onDeliveries = { selectTab(Routes.DELIVERIES) },
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
                // 表示・復帰のたびに読み直す。文箱にしまった結果（kept）や新着配信を反映する。
                LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { vm.load() }
                DeliveriesScreen(
                    state = state,
                    onOpen = { card ->
                        // ReceivedCard は @Serializable だが Parcelable ではないため、SavedStateHandle に
                        // そのまま入れると落ちる。JSON 文字列に畳んで渡す（castBody と同じ「String を渡す」流儀）。
                        nav.currentBackStackEntry?.savedStateHandle?.set("openedCard", card.encode())
                        nav.navigate(Routes.CHIME)
                    },
                )
            }
            composable(Routes.CHIME) {
                val vm: ChimeViewModel = viewModel(
                    factory = viewModelFactory { initializer { ChimeViewModel(container.repository) } },
                )
                val state by vm.state.collectAsStateWithLifecycle()
                val reportState by vm.reportState.collectAsStateWithLifecycle()
                val card = remember {
                    nav.previousBackStackEntry?.savedStateHandle?.get<String>("openedCard")?.let(::decodeReceivedCard)
                }
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
                // タブは状態を保持するため、再表示のたびに読み直して新しい記録を反映する。
                LifecycleEventEffect(Lifecycle.Event.ON_RESUME) { vm.load() }
                BoxScreen(state = state)
            }
        }
    }
}

// 下部タブ。漢字（今／風／文）をしるしに、選択中は紺で示す（余計な下地は置かない）。
@Composable
private fun KarinBottomBar(currentRoute: String?, onSelect: (String) -> Unit) {
    val tabs = listOf(
        Triple(Routes.HOME, "今", "今日"),
        Triple(Routes.DELIVERIES, "風", "風だより"),
        Triple(Routes.BOX, "文", "文箱"),
    )
    NavigationBar(containerColor = MaterialTheme.colorScheme.surface) {
        tabs.forEach { (dest, kanji, label) ->
            NavigationBarItem(
                selected = currentRoute == dest,
                onClick = { onSelect(dest) },
                icon = { Text(kanji, style = MaterialTheme.typography.titleLarge) },
                label = { Text(label, style = MaterialTheme.typography.labelMedium) },
                colors = NavigationBarItemDefaults.colors(
                    selectedIconColor = MaterialTheme.colorScheme.primary,
                    selectedTextColor = MaterialTheme.colorScheme.primary,
                    unselectedIconColor = MaterialTheme.colorScheme.onSurfaceVariant,
                    unselectedTextColor = MaterialTheme.colorScheme.onSurfaceVariant,
                    indicatorColor = Color.Transparent,
                ),
            )
        }
    }
}
