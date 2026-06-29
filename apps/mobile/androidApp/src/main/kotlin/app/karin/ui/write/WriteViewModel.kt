package app.karin.ui.write

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.karin.shared.api.KarinRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

// WriteViewModel は短冊の保存（文箱にしまう）を司る。候は省略しサーバが今日の候を補う。
class WriteViewModel(private val repo: KarinRepository) : ViewModel() {

    sealed interface State {
        data object Editing : State
        data object Saving : State
        data object Saved : State
        data class Error(val message: String) : State
    }

    private val _state = MutableStateFlow<State>(State.Editing)
    val state: StateFlow<State> = _state.asStateFlow()

    // 文箱にしまう。空文字は保存しない（UI でもボタンを無効化する）。
    fun save(body: String) {
        val text = body.trim()
        if (text.isEmpty() || _state.value is State.Saving) return
        _state.value = State.Saving
        viewModelScope.launch {
            _state.value = runCatching { repo.createRecord(text) }
                .fold(
                    onSuccess = { State.Saved },
                    onFailure = { State.Error(it.message ?: "しまえませんでした。少し待って、もう一度。") },
                )
        }
    }
}
