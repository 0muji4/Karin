package app.karin.ui.box

import app.karin.shared.api.BoxGroupDto
import app.karin.shared.api.BoxResponse
import app.karin.shared.api.CastResponse
import app.karin.shared.api.KarinRepository
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

private fun repo(box: suspend () -> BoxResponse) = object : KarinRepository {
    override suspend fun todayKo(): TodayResponse = error("未使用")
    override suspend fun createRecord(body: String, koWritten: Int?): RecordDto = error("未使用")
    override suspend fun listBox(): BoxResponse = box()
    override suspend fun cast(recordId: String): CastResponse = error("未使用")
}

private fun record(id: String, ko: Int) = RecordDto(id, ko, "本文", "2026-06-29T00:00:00Z")

@OptIn(ExperimentalCoroutinesApi::class)
class BoxViewModelTest {
    @BeforeTest
    fun setUp() = Dispatchers.setMain(StandardTestDispatcher())

    @AfterTest
    fun tearDown() = Dispatchers.resetMain()

    @Test
    fun 収集枚数はグループ横断で合計される() = runTest {
        val box = BoxResponse(
            groups = listOf(
                BoxGroupDto(WafuMonthDto("水無月", "みなづき"), SekkiDto(10, "夏至", "げし"), listOf(record("a", 28), record("b", 29))),
                BoxGroupDto(WafuMonthDto("水無月", "みなづき"), SekkiDto(11, "小暑", "しょうしょ"), listOf(record("c", 31))),
            ),
        )
        val vm = BoxViewModel(repo { box })
        advanceUntilIdle()
        val state = vm.state.value
        assertTrue(state is BoxViewModel.State.Loaded)
        assertEquals(3, (state as BoxViewModel.State.Loaded).total)
    }

    @Test
    fun 失敗で_Error_になる() = runTest {
        val vm = BoxViewModel(repo { throw RuntimeException("boom") })
        advanceUntilIdle()
        assertTrue(vm.state.value is BoxViewModel.State.Error)
    }
}
