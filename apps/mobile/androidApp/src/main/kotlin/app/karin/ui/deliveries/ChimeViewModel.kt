package app.karin.ui.deliveries

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.karin.shared.api.KarinRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

// ChimeViewModel は受信した一枚を文箱にしまう／通報する操作を司る（開封状態は画面側でローカルに記録）。
class ChimeViewModel(private val repo: KarinRepository) : ViewModel() {

    sealed interface State {
        data object Idle : State
        data object Keeping : State
        data object Kept : State
        data class Error(val message: String) : State
    }

    sealed interface ReportState {
        data object Idle : ReportState
        data object Reporting : ReportState
        data object Reported : ReportState
        data class Error(val message: String) : ReportState
    }

    private val _state = MutableStateFlow<State>(State.Idle)
    val state: StateFlow<State> = _state.asStateFlow()

    private val _reportState = MutableStateFlow<ReportState>(ReportState.Idle)
    val reportState: StateFlow<ReportState> = _reportState.asStateFlow()

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

    fun report(tanzakuId: String, reason: String) {
        if (_reportState.value is ReportState.Reporting || _reportState.value is ReportState.Reported) return
        _reportState.value = ReportState.Reporting
        viewModelScope.launch {
            _reportState.value = runCatching { repo.report(tanzakuId, reason, "") }
                .fold(
                    onSuccess = { ReportState.Reported },
                    onFailure = { ReportState.Error(it.message ?: "通報を受け付けられませんでした") },
                )
        }
    }
}
