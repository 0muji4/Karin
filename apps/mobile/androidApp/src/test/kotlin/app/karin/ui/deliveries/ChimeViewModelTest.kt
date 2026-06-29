package app.karin.ui.deliveries

import app.karin.shared.api.BoxResponse
import app.karin.shared.api.CastResponse
import app.karin.shared.api.DeliveriesResponse
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.RecordDto
import app.karin.shared.api.StatusResponse
import app.karin.shared.api.TodayResponse
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
import kotlin.test.assertTrue

private fun repo(
    keepThrows: Boolean = false,
    onKeep: (String) -> Unit = {},
    reportThrows: Boolean = false,
    onReport: (String, String) -> Unit = { _, _ -> },
) = object : KarinRepository {
    override suspend fun todayKo(): TodayResponse = error("未使用")
    override suspend fun createRecord(body: String, koWritten: Int?): RecordDto = error("未使用")
    override suspend fun listBox(): BoxResponse = error("未使用")
    override suspend fun cast(recordId: String): CastResponse = error("未使用")
    override suspend fun listDeliveries(): DeliveriesResponse = error("未使用")
    override suspend fun keep(tanzakuId: String): StatusResponse {
        if (keepThrows) throw RuntimeException("boom")
        onKeep(tanzakuId)
        return StatusResponse("kept")
    }
    override suspend fun report(tanzakuId: String, reason: String, note: String): StatusResponse {
        if (reportThrows) throw RuntimeException("boom")
        onReport(tanzakuId, reason)
        return StatusResponse("reported")
    }
}

@OptIn(ExperimentalCoroutinesApi::class)
class ChimeViewModelTest {
    @BeforeTest
    fun setUp() = Dispatchers.setMain(StandardTestDispatcher())

    @AfterTest
    fun tearDown() = Dispatchers.resetMain()

    @Test
    fun しまうと当該tanzakuで_Kept() = runTest {
        var kept: String? = null
        val vm = ChimeViewModel(repo(onKeep = { kept = it }))
        vm.keep("t-1")
        advanceUntilIdle()
        assertTrue(vm.state.value is ChimeViewModel.State.Kept)
        assertTrue(kept == "t-1")
    }

    @Test
    fun 失敗で_Error() = runTest {
        val vm = ChimeViewModel(repo(keepThrows = true))
        vm.keep("t-1")
        advanceUntilIdle()
        assertTrue(vm.state.value is ChimeViewModel.State.Error)
    }

    @Test
    fun 通報すると理由つきで_Reported() = runTest {
        var seen: Pair<String, String>? = null
        val vm = ChimeViewModel(repo(onReport = { id, reason -> seen = id to reason }))
        vm.report("t-1", "harassment")
        advanceUntilIdle()
        assertTrue(vm.reportState.value is ChimeViewModel.ReportState.Reported)
        assertTrue(seen == ("t-1" to "harassment"))
    }

    @Test
    fun 通報失敗で_ReportState_Error() = runTest {
        val vm = ChimeViewModel(repo(reportThrows = true))
        vm.report("t-1", "spam")
        advanceUntilIdle()
        assertTrue(vm.reportState.value is ChimeViewModel.ReportState.Error)
    }
}
