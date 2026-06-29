package app.karin.ui.deliveries

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.karin.shared.api.KarinRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

// ChimeViewModel は受信した一枚を文箱にしまう操作を司る（開封状態は画面側でローカルに記録する）。
class ChimeViewModel(private val repo: KarinRepository) : ViewModel() {

    sealed interface State {
        data object Idle : State
        data object Keeping : State
        data object Kept : State
        data class Error(val message: String) : State
    }

    private val _state = MutableStateFlow<State>(State.Idle)
    val state: StateFlow<State> = _state.asStateFlow()

    fun keep(tanzakuId: String) {
        if (_state.value is State.Keeping || _state.value is State.Kept) return
        _state.value = State.Keeping
        viewModelScope.launch {
            _state.value = runCatching { repo.keep(tanzakuId) }
                .fold(
                    onSuccess = { State.Kept },
                    onFailure = { State.Error(it.message ?: "文箱にしまえませんでした") },
                )
        }
    }
}
