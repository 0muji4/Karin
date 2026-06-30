package app.karin.ui.today

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.karin.data.ReadStateStore
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.TodayResponse
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

// TodayViewModel は今日の候を取得して画面状態に変換する。あわせて風だよりの未読数を導出し、
// home に「届いています」の合図を出す（未読は受信一覧と端末ローカルの開封済みの差）。
// 読み込みは画面表示（ON_RESUME）から呼ぶ。戻った時に未読を最新化するため、生成時ではなく表示で読む。
class TodayViewModel(
    private val repo: KarinRepository,
    private val readState: ReadStateStore,
) : ViewModel() {

    sealed interface State {
        data object Loading : State
        data class Loaded(val today: TodayResponse, val unreadCount: Int) : State
        data class Error(val message: String) : State
    }

    private val _state = MutableStateFlow<State>(State.Loading)
    val state: StateFlow<State> = _state.asStateFlow()

    fun load() {
        _state.value = State.Loading
        viewModelScope.launch {
            _state.value = runCatching { repo.todayKo() }.fold(
                onSuccess = { today -> State.Loaded(today, unreadCount()) },
                onFailure = { State.Error(it.message ?: "今日の候を読み込めませんでした") },
            )
        }
    }

    // 未読は付随情報。取得に失敗しても今日の候は見せたいので、失敗時は 0 とみなす。
    private suspend fun unreadCount(): Int = runCatching {
        val opened = readState.openedIds()
        repo.listDeliveries().received.count { it.tanzakuId !in opened }
    }.getOrDefault(0)
}
