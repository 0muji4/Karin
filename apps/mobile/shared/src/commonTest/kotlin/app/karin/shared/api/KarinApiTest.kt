package app.karin.shared.api

import app.karin.shared.auth.TokenStore
import app.karin.shared.net.KarinClient
import app.karin.shared.net.KarinError
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.client.request.HttpRequestData
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.content.OutgoingContent
import io.ktor.http.headersOf
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertTrue

private class FakeTokenStore(private var token: String? = null) : TokenStore {
    override fun load(): String? = token
    override fun save(token: String) { this.token = token }
    override fun clear() { token = null }
}

private fun client(
    status: HttpStatusCode = HttpStatusCode.OK,
    body: String = "{}",
    record: ((HttpRequestData) -> Unit)? = null,
): KarinClient {
    val engine = MockEngine { req ->
        record?.invoke(req)
        respond(body, status, headersOf(HttpHeaders.ContentType, "application/json"))
    }
    return KarinClient("https://api.example", FakeTokenStore("tok"), engine)
}

class KarinApiTest {
    @Test
    fun createAccount_は_user_id_と_token_を読む() = runTest {
        val seen = mutableListOf<String>()
        val c = client(HttpStatusCode.Created, """{"user_id":"u-1","token":"t-1"}""") { seen += it.url.encodedPath }
        val res = c.createAccount()
        assertEquals("u-1", res.userId)
        assertEquals("t-1", res.token)
        assertTrue(seen.single().endsWith("/accounts"))
    }

    @Test
    fun todayKo_は入れ子の候まで読む() = runTest {
        val json = """
            {"date":"2026-06-29","wafu_month":{"name":"水無月","kana":"みなづき"},
             "sekki":{"number":10,"name":"夏至","kana":"げし"},
             "ko":{"number":28,"name":"乃東枯","kana":"なつかれくさかるる","meaning":"夏枯草が枯れる","sekki":10,"season":"夏"}}
        """.trimIndent()
        val res = client(body = json).todayKo()
        assertEquals("乃東枯", res.ko.name)
        assertEquals("みなづき", res.wafuMonth.kana)
        assertEquals(10, res.sekki.number)
    }

    @Test
    fun createRecord_は本文と候を送る() = runTest {
        var sentBody = ""
        val c = client(HttpStatusCode.Created, """{"id":"r-1","ko_written":28,"body":"蝉の声","created_at":"2026-06-29T00:00:00Z"}""") {
            sentBody = (it.body as OutgoingContent.ByteArrayContent).bytes().decodeToString()
        }
        val res = c.createRecord(CreateRecordRequest(body = "蝉の声", koWritten = 28))
        assertEquals("r-1", res.id)
        assertTrue(sentBody.contains("蝉の声"))
        assertTrue(sentBody.contains("28"))
    }

    @Test
    fun listBox_は節気グループを読む() = runTest {
        val json = """
            {"groups":[{"wafu_month":{"name":"水無月","kana":"みなづき"},
              "sekki":{"number":10,"name":"夏至","kana":"げし"},
              "records":[{"id":"r-1","ko_written":28,"body":"蝉の声","created_at":"2026-06-29T00:00:00Z"}]}]}
        """.trimIndent()
        val res = client(body = json).listBox()
        assertEquals(1, res.groups.size)
        assertEquals("蝉の声", res.groups.first().records.first().body)
    }

    @Test
    fun _401_は_Unauthorized_に正規化される() = runTest {
        val c = client(HttpStatusCode.Unauthorized, "")
        assertFailsWith<KarinError.Unauthorized> { c.listBox() }
    }

    @Test
    fun cast_は危機時に支援先を読む() = runTest {
        val json = """{"status":"cast","support":{"message":"つらいときは","url":"https://example/support"}}"""
        val res = client(body = json).castToWind("r-1")
        assertEquals("cast", res.status)
        assertEquals("https://example/support", res.support?.url)
    }

    @Test
    fun cast_は通常は支援先なし() = runTest {
        val res = client(body = """{"status":"cast"}""").castToWind("r-9")
        assertEquals("cast", res.status)
        assertEquals(null, res.support)
    }

    @Test
    fun listDeliveries_は受信一覧を読む() = runTest {
        val json = """{"received":[{"tanzaku_id":"t-1","body":"青梅","ko":29,"is_official":false,"delivered_on":"2026-06-29","kept":false}]}"""
        val res = client(body = json).listDeliveries()
        assertEquals(1, res.received.size)
        assertEquals("t-1", res.received.first().tanzakuId)
        assertEquals(false, res.received.first().kept)
    }

    @Test
    fun keep_は当該tanzakuへ_POSTする() = runTest {
        val seen = mutableListOf<String>()
        val c = client(HttpStatusCode.OK, """{"status":"kept"}""") { seen += it.url.encodedPath }
        val res = c.keep("t-1")
        assertEquals("kept", res.status)
        assertTrue(seen.single().endsWith("/deliveries/t-1/keep"))
    }

    @Test
    fun report_は_tanzaku_id_と_reason_を送る() = runTest {
        var sent = ""
        val c = client(HttpStatusCode.OK, """{"status":"reported"}""") {
            sent = (it.body as OutgoingContent.ByteArrayContent).bytes().decodeToString()
        }
        val res = c.report("t-1", "harassment", "ひどい")
        assertEquals("reported", res.status)
        assertTrue(sent.contains("t-1"))
        assertTrue(sent.contains("harassment"))
    }
}
