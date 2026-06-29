package app.karin.ui.today

import app.karin.shared.api.BoxResponse
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.KoDto
import app.karin.shared.api.RecordDto
import app.karin.shared.api.SekkiDto
import app.karin.shared.api.TodayResponse
import app.karin.shared.api.WafuMonthDto
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import kotlin.test.AfterTest
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

private val sampleToday = TodayResponse(
    date = "2026-06-29",
    wafuMonth = WafuMonthDto("水無月", "みなづき"),
    sekki = SekkiDto(10, "夏至", "げし"),
    ko = KoDto(28, "乃東枯", "なつかれくさかるる", "夏枯草が枯れる", 10, "夏"),
)

// 未使用のメソッドは呼ばれたら失敗させる fake リポジトリ。
private fun repo(today: suspend () -> TodayResponse) = object : KarinRepository {
    override suspend fun todayKo(): TodayResponse = today()
    override suspend fun createRecord(body: String, koWritten: Int?): RecordDto = error("未使用")
    override suspend fun listBox(): BoxResponse = error("未使用")
}

@OptIn(ExperimentalCoroutinesApi::class)
class TodayViewModelTest {
    @BeforeTest
    fun setUp() = Dispatchers.setMain(StandardTestDispatcher())

    @AfterTest
    fun tearDown() = Dispatchers.resetMain()

    @Test
    fun 読み込み成功で候を表示できる() = runTest {
        val vm = TodayViewModel(repo { sampleToday })
        advanceUntilIdle()
        val state = vm.state.value
        assertTrue(state is TodayViewModel.State.Loaded)
        assertEquals("乃東枯", (state as TodayViewModel.State.Loaded).today.ko.name)
    }

    @Test
    fun 失敗で_Error_になる() = runTest {
        val vm = TodayViewModel(repo { throw RuntimeException("boom") })
        advanceUntilIdle()
        assertTrue(vm.state.value is TodayViewModel.State.Error)
    }
}
