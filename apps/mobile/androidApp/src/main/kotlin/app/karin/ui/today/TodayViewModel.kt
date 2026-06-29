package app.karin.ui.today

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.TodayResponse
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

// TodayViewModel は今日の候を取得して画面状態に変換する薄いグルー。取得は KarinRepository に委ねる。
class TodayViewModel(private val repo: KarinRepository) : ViewModel() {

    sealed interface State {
        data object Loading : State
        data class Loaded(val today: TodayResponse) : State
        data class Error(val message: String) : State
    }

    private val _state = MutableStateFlow<State>(State.Loading)
    val state: StateFlow<State> = _state.asStateFlow()

    init {
        load()
    }

    fun load() {
        _state.value = State.Loading
        viewModelScope.launch {
            _state.value = runCatching { repo.todayKo() }
                .fold(
                    onSuccess = { State.Loaded(it) },
                    onFailure = { State.Error(it.message ?: "今日の候を読み込めませんでした") },
                )
        }
    }
}
