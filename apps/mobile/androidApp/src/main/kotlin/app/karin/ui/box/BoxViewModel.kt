package app.karin.ui.box

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import app.karin.shared.api.BoxResponse
import app.karin.shared.api.KarinRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

// BoxViewModel は文箱（節気ごとの記録）を取得し画面状態に変換する。収集枚数はクライアントで合計する。
class BoxViewModel(private val repo: KarinRepository) : ViewModel() {

    sealed interface State {
        data object Loading : State
        data class Loaded(val box: BoxResponse, val total: Int) : State
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
            _state.value = runCatching { repo.listBox() }
                .fold(
                    onSuccess = { box -> State.Loaded(box, box.groups.sumOf { it.records.size }) },
                    onFailure = { State.Error(it.message ?: "文箱を読み込めませんでした") },
                )
        }
    }
}
