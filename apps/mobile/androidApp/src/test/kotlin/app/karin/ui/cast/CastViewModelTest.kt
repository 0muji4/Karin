package app.karin.ui.cast

import app.karin.shared.api.BoxResponse
import app.karin.shared.api.CastResponse
import app.karin.shared.api.DeliveriesResponse
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.RecordDto
import app.karin.shared.api.StatusResponse
import app.karin.shared.api.SupportInfo
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
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

private val savedRecord = RecordDto(id = "r-1", koWritten = 28, body = "蝉の声", createdAt = "2026-06-29T00:00:00Z")

// createRecord→cast を続けて行う fake。create が投げる場合と、cast の応答を差し替えられる。
private fun repo(
    createThrows: Boolean = false,
    castResult: CastResponse = CastResponse(status = "cast"),
) = object : KarinRepository {
    override suspend fun todayKo(): TodayResponse = error("未使用")
    override suspend fun createRecord(body: String, koWritten: Int?): RecordDto {
        if (createThrows) throw RuntimeException("boom")
        return savedRecord
    }
    override suspend fun listBox(): BoxResponse = error("未使用")
    override suspend fun cast(recordId: String): CastResponse = castResult
    override suspend fun listDeliveries(): DeliveriesResponse = error("未使用")
    override suspend fun keep(tanzakuId: String): StatusResponse = error("未使用")
    override suspend fun report(tanzakuId: String, reason: String, note: String): StatusResponse = error("未使用")
}

@OptIn(ExperimentalCoroutinesApi::class)
class CastViewModelTest {
    @BeforeTest
    fun setUp() = Dispatchers.setMain(StandardTestDispatcher())

    @AfterTest
    fun tearDown() = Dispatchers.resetMain()

    @Test
    fun 成功すると支援先なしで_Sent() = runTest {
        val vm = CastViewModel(repo(castResult = CastResponse(status = "cast")))
        vm.cast("蝉の声")
        advanceUntilIdle()
        val s = vm.state.value
        assertTrue(s is CastViewModel.State.Sent)
        assertNull((s as CastViewModel.State.Sent).support)
    }

    @Test
    fun 危機判定では支援先付きで_Sent() = runTest {
        val support = SupportInfo("つらいときは", "https://example/support")
        val vm = CastViewModel(repo(castResult = CastResponse(status = "cast", support = support)))
        vm.cast("もう消えたい")
        advanceUntilIdle()
        val s = vm.state.value
        assertTrue(s is CastViewModel.State.Sent)
        assertEquals("https://example/support", (s as CastViewModel.State.Sent).support?.url)
    }

    @Test
    fun 保存に失敗すると_Error() = runTest {
        val vm = CastViewModel(repo(createThrows = true))
        vm.cast("蝉の声")
        advanceUntilIdle()
        assertTrue(vm.state.value is CastViewModel.State.Error)
    }
}
