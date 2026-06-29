package app.karin.shared.session

import app.karin.shared.auth.TokenStore
import app.karin.shared.net.KarinClient
import app.karin.shared.net.KarinError
import io.ktor.client.engine.mock.MockEngine
import io.ktor.client.engine.mock.respond
import io.ktor.http.HttpHeaders
import io.ktor.http.HttpStatusCode
import io.ktor.http.headersOf
import kotlinx.coroutines.test.runTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNull

private class FakeTokenStore(private var token: String? = null) : TokenStore {
    override fun load(): String? = token
    override fun save(token: String) { this.token = token }
    override fun clear() { token = null }
}

private fun client(status: HttpStatusCode, body: String, store: TokenStore): KarinClient {
    val engine = MockEngine { respond(body, status, headersOf(HttpHeaders.ContentType, "application/json")) }
    return KarinClient("https://api.example", store, engine)
}

class SessionRepositoryTest {
    @Test
    fun 発行に成功するとトークンを保存する() = runTest {
        val store = FakeTokenStore()
        val repo = SessionRepository(client(HttpStatusCode.Created, """{"user_id":"u","token":"tok-xyz"}""", store), store)
        val res = repo.issueAccount()
        assertEquals("tok-xyz", res.token)
        assertEquals("tok-xyz", store.load())
        assertEquals(true, repo.isSignedIn())
    }

    @Test
    fun 発行に失敗するとトークンを保存しない() = runTest {
        val store = FakeTokenStore()
        val repo = SessionRepository(client(HttpStatusCode.InternalServerError, "", store), store)
        assertFailsWith<KarinError.Server> { repo.issueAccount() }
        assertNull(store.load())
        assertEquals(false, repo.isSignedIn())
    }
}
