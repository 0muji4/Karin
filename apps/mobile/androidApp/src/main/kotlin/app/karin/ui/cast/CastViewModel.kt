package app.karin.ui.cast

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.SupportInfo
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

// CastViewModel は「風に乗せる」を司る。cast は記録 ID を要するため、まず保存してから風に乗せる
// （記録は文箱にも残る）。判定結果は著者に見せないので、成功は一律 Sent。危機時のみ support を伴う。
class CastViewModel(private val repo: KarinRepository) : ViewModel() {

    sealed interface State {
        data object Idle : State
        data object Casting : State
        data class Sent(val support: SupportInfo?) : State
        data class Error(val message: String) : State
    }

    private val _state = MutableStateFlow<State>(State.Idle)
    val state: StateFlow<State> = _state.asStateFlow()

    fun cast(body: String) {
        val text = body.trim()
        if (text.isEmpty() || _state.value is State.Casting) return
        _state.value = State.Casting
        viewModelScope.launch {
            _state.value = runCatching {
                val record = repo.createRecord(text)
                repo.cast(record.id)
            }.fold(
                onSuccess = { State.Sent(it.support) },
                onFailure = { State.Error(it.message ?: "風に乗せられませんでした。少し待って、もう一度。") },
            )
        }
    }
}
