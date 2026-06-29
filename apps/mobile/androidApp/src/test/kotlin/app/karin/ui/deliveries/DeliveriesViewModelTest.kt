package app.karin.ui.deliveries

import app.karin.data.ReadStateStore
import app.karin.shared.api.BoxResponse
import app.karin.shared.api.CastResponse
import app.karin.shared.api.DeliveriesResponse
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.ReceivedCard
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
import kotlin.test.assertEquals
import kotlin.test.assertTrue

private fun card(id: String) = ReceivedCard(tanzakuId = id, body = "本文", ko = 29, isOfficial = false, deliveredOn = "2026-06-29", kept = false)

private fun repo(received: List<ReceivedCard>) = object : KarinRepository {
    override suspend fun todayKo(): TodayResponse = error("未使用")
    override suspend fun createRecord(body: String, koWritten: Int?): RecordDto = error("未使用")
    override suspend fun listBox(): BoxResponse = error("未使用")
    override suspend fun cast(recordId: String): CastResponse = error("未使用")
    override suspend fun listDeliveries(): DeliveriesResponse = DeliveriesResponse(received)
    override suspend fun keep(tanzakuId: String): StatusResponse = error("未使用")
    override suspend fun report(tanzakuId: String, reason: String, note: String): StatusResponse = error("未使用")
}

private fun readState(opened: Set<String>) = object : ReadStateStore {
    override fun openedIds(): Set<String> = opened
    override fun markOpened(id: String) {}
}

@OptIn(ExperimentalCoroutinesApi::class)
class DeliveriesViewModelTest {
    @BeforeTest
    fun setUp() = Dispatchers.setMain(StandardTestDispatcher())

    @AfterTest
    fun tearDown() = Dispatchers.resetMain()

    @Test
    fun 開封済みでないものを未読として数える() = runTest {
        val vm = DeliveriesViewModel(
            repo(listOf(card("a"), card("b"), card("c"))),
            readState(setOf("b")), // b は開封済み
        )
        vm.load() // 読み込みは画面表示時に呼ぶ設計（init では走らない）
        advanceUntilIdle()
        val s = vm.state.value
        assertTrue(s is DeliveriesViewModel.State.Loaded)
        s as DeliveriesViewModel.State.Loaded
        assertEquals(2, s.unreadCount)
        assertEquals(false, s.items.first { it.card.tanzakuId == "b" }.unread)
        assertEquals(true, s.items.first { it.card.tanzakuId == "a" }.unread)
    }

    @Test
    fun 受信ゼロは未読ゼロ() = runTest {
        val vm = DeliveriesViewModel(repo(emptyList()), readState(emptySet()))
        vm.load() // 読み込みは画面表示時に呼ぶ設計（init では走らない）
        advanceUntilIdle()
        val s = vm.state.value
        assertTrue(s is DeliveriesViewModel.State.Loaded)
        assertEquals(0, (s as DeliveriesViewModel.State.Loaded).unreadCount)
    }
}
