package app.karin.ui.today

import app.karin.data.ReadStateStore
import app.karin.shared.api.BoxResponse
import app.karin.shared.api.CastResponse
import app.karin.shared.api.DeliveriesResponse
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.KoDto
import app.karin.shared.api.ReceivedCard
import app.karin.shared.api.RecordDto
import app.karin.shared.api.SekkiDto
import app.karin.shared.api.StatusResponse
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

private fun card(id: String) =
    ReceivedCard(tanzakuId = id, body = "本文", ko = 29, isOfficial = false, deliveredOn = "2026-06-29", kept = false)

// 未使用のメソッドは呼ばれたら失敗させる fake リポジトリ。受信一覧は未読導出に使う。
private fun repo(
    today: suspend () -> TodayResponse,
    received: List<ReceivedCard> = emptyList(),
) = object : KarinRepository {
    override suspend fun todayKo(): TodayResponse = today()
    override suspend fun createRecord(body: String, koWritten: Int?): RecordDto = error("未使用")
    override suspend fun listBox(): BoxResponse = error("未使用")
    override suspend fun cast(recordId: String): CastResponse = error("未使用")
    override suspend fun listDeliveries(): DeliveriesResponse = DeliveriesResponse(received)
    override suspend fun keep(tanzakuId: String): StatusResponse = error("未使用")
    override suspend fun report(tanzakuId: String, reason: String, note: String): StatusResponse = error("未使用")
}

private fun readState(opened: Set<String> = emptySet()) = object : ReadStateStore {
    override fun openedIds(): Set<String> = opened
    override fun markOpened(id: String) {}
}

@OptIn(ExperimentalCoroutinesApi::class)
class TodayViewModelTest {
    @BeforeTest
    fun setUp() = Dispatchers.setMain(StandardTestDispatcher())

    @AfterTest
    fun tearDown() = Dispatchers.resetMain()

    @Test
    fun 読み込み成功で候を表示できる() = runTest {
        val vm = TodayViewModel(repo({ sampleToday }), readState())
        vm.load()
        advanceUntilIdle()
        val state = vm.state.value
        assertTrue(state is TodayViewModel.State.Loaded)
        assertEquals("乃東枯", (state as TodayViewModel.State.Loaded).today.ko.name)
    }

    @Test
    fun 失敗で_Error_になる() = runTest {
        val vm = TodayViewModel(repo({ throw RuntimeException("boom") }), readState())
        vm.load()
        advanceUntilIdle()
        assertTrue(vm.state.value is TodayViewModel.State.Error)
    }

    @Test
    fun 開封済みでない受信を未読として数える() = runTest {
        val vm = TodayViewModel(
            repo({ sampleToday }, received = listOf(card("a"), card("b"), card("c"))),
            readState(setOf("a")), // a は開封済み
        )
        vm.load()
        advanceUntilIdle()
        val state = vm.state.value
        assertTrue(state is TodayViewModel.State.Loaded)
        assertEquals(2, (state as TodayViewModel.State.Loaded).unreadCount)
    }
}
