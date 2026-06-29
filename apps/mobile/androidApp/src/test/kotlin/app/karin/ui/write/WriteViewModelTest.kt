package app.karin.ui.write

import app.karin.shared.api.BoxResponse
import app.karin.shared.api.CastResponse
import app.karin.shared.api.KarinRepository
import app.karin.shared.api.RecordDto
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

private fun repo(create: suspend (String) -> RecordDto) = object : KarinRepository {
    override suspend fun todayKo(): TodayResponse = error("未使用")
    override suspend fun createRecord(body: String, koWritten: Int?): RecordDto = create(body)
    override suspend fun listBox(): BoxResponse = error("未使用")
    override suspend fun cast(recordId: String): CastResponse = error("未使用")
}

@OptIn(ExperimentalCoroutinesApi::class)
class WriteViewModelTest {
    @BeforeTest
    fun setUp() = Dispatchers.setMain(StandardTestDispatcher())

    @AfterTest
    fun tearDown() = Dispatchers.resetMain()

    private val saved = RecordDto(id = "r-1", koWritten = 28, body = "蝉の声", createdAt = "2026-06-29T00:00:00Z")

    @Test
    fun 保存に成功すると_Saved_になり本文が送られる() = runTest {
        var sent: String? = null
        val vm = WriteViewModel(repo { sent = it; saved })
        vm.save("　蝉の声　") // 前後の空白は trim される
        advanceUntilIdle()
        assertTrue(vm.state.value is WriteViewModel.State.Saved)
        assertEquals("蝉の声", sent)
    }

    @Test
    fun 空文字は保存しない() = runTest {
        var called = false
        val vm = WriteViewModel(repo { called = true; saved })
        vm.save("   ")
        advanceUntilIdle()
        assertTrue(vm.state.value is WriteViewModel.State.Editing)
        assertEquals(false, called)
    }

    @Test
    fun 失敗で_Error_になる() = runTest {
        val vm = WriteViewModel(repo { throw RuntimeException("boom") })
        vm.save("蝉の声")
        advanceUntilIdle()
        assertTrue(vm.state.value is WriteViewModel.State.Error)
    }
}
