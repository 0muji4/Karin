package app.karin.shared.net

import app.karin.shared.auth.TokenStore
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.client.request.get
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNull

private class FakeTokenStore(private var token: String?) : TokenStore {
    override fun load(): String? = token
    override fun save(token: String) { this.token = token }
    override fun clear() { token = null }
}

class KarinClientTest {
    private fun clientReturning(status: HttpStatusCode, body: String, capture: MutableList<String?>? = null): KarinClient {
        val engine = MockEngine { request ->
            capture?.add(request.headers[HttpHeaders.Authorization])
            respond(
                content = body,
                status = status,
                headers = headersOf(HttpHeaders.ContentType, "application/json"),
            )
        }
        return KarinClient("https://api.example", FakeTokenStore("tok123"), engine)
    }

    @Test
    fun トークンがあれば_Bearer_を載せる() = runTest {
        val seen = mutableListOf<String?>()
        val c = clientReturning(HttpStatusCode.OK, "{}", seen)
        c.http.get("ko/today")
        assertEquals("Bearer tok123", seen.single())
    }

    @Test
    fun 成功なら_throwIfError_は投げない() = runTest {
        val c = clientReturning(HttpStatusCode.OK, "{}")
        val resp = c.http.get("ko/today")
        assertNull(runCatching { c.throwIfError(resp) }.exceptionOrNull())
    }

    @Test
    fun _401_は_Unauthorized() = runTest {
        val c = clientReturning(HttpStatusCode.Unauthorized, "")
        val resp = c.http.get("box")
        assertFailsWith<KarinError.Unauthorized>("Unauthorized を期待") { c.throwIfError(resp) }
    }

    @Test
    fun エラー包みは_Api_に正規化される() = runTest {
        val c = clientReturning(HttpStatusCode.BadRequest, """{"error":{"code":"invalid_record","message":"本文が空"}}""")
        val resp = c.http.get("records")
        val err = runCatching { c.throwIfError(resp) }.exceptionOrNull()
        val api = err as? KarinError.Api
        assertEquals("invalid_record", api?.code)
    }
}
