package app.karin.ui.deliveries

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.karin.data.ReadStateStore
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.ReceivedCard
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

// DeliveriesViewModel は風だより（受信一覧）を取得し、端末ローカルの開封済み集合と突き合わせて
// 未読を導出する（バックエンドに既読の概念が無いため）。
class DeliveriesViewModel(
    private val repo: KarinRepository,
    private val readState: ReadStateStore,
) : ViewModel() {

    data class Item(val card: ReceivedCard, val unread: Boolean)

    sealed interface State {
        data object Loading : State
        data class Loaded(val items: List<Item>, val unreadCount: Int) : State
        data class Error(val message: String) : State
    }

    private val _state = MutableStateFlow<State>(State.Loading)
    val state: StateFlow<State> = _state.asStateFlow()

    // 読み込みは画面の表示（ON_RESUME）から呼ぶ。風だよりは別画面での操作（文箱にしまう）や
    // 新着配信を、戻ってきた時に反映する必要があるため、VM 生成時ではなく表示のたびに読み直す。
    fun load() {
        _state.value = State.Loading
        viewModelScope.launch {
            _state.value = runCatching {
                val opened = readState.openedIds()
                repo.listDeliveries().received.map { Item(it, unread = it.tanzakuId !in opened) }
            }.fold(
                onSuccess = { items -> State.Loaded(items, items.count { it.unread }) },
                onFailure = { State.Error(it.message ?: "風だよりを読み込めませんでした") },
            )
        }
    }
}
