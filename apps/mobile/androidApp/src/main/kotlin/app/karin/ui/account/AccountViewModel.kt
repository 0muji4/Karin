package app.karin.ui.account

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.karin.shared.session.SessionRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

// AccountViewModel はオンボーディング完了時のアカウント発行を司る薄いグルー。
// 発行ロジックは SessionRepository（共有層・テスト済）にあり、ここは UI 状態だけを持つ。
class AccountViewModel(private val session: SessionRepository) : ViewModel() {

    sealed interface State {
        data object Idle : State
        data object Loading : State
        data object Done : State
        data class Error(val message: String) : State
    }

    private val _state = MutableStateFlow<State>(State.Idle)
    val state: StateFlow<State> = _state.asStateFlow()

    fun issue() {
        if (_state.value is State.Loading) return
        _state.value = State.Loading
        viewModelScope.launch {
            _state.value = runCatching { session.issueAccount() }
                .fold(
                    onSuccess = { State.Done },
                    onFailure = { State.Error(it.message ?: "はじめられませんでした。少し待って、もう一度。") },
                )
        }
    }
}
