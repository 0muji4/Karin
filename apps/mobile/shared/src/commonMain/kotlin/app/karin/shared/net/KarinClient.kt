package app.karin.shared.net

import app.karin.shared.auth.TokenStore
import io.ktor.client.HttpClient
import io.ktor.client.HttpClientConfig
import io.ktor.client.engine.HttpClientEngine
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.plugins.defaultRequest
import io.ktor.client.request.bearerAuth
import io.ktor.client.request.url
import io.ktor.client.statement.HttpResponse
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.http.isSuccess
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

// KarinClient は HTTP の土台。base URL・JSON・Bearer 付与・失敗の KarinError への正規化を担い、
// 個々のエンドポイントは別ファイル（KarinApi）が本クライアントを使って実装する。
// engine を注入できるのはテスト（MockEngine）のため。null なら各プラットフォーム既定エンジン。
class KarinClient(
    baseUrl: String,
    private val tokenStore: TokenStore,
    engine: HttpClientEngine? = null,
) {
    val json: Json = Json {
        ignoreUnknownKeys = true // サーバが増やしたフィールドで壊れない
        encodeDefaults = true
    }

    val http: HttpClient = run {
        val normalizedBase = if (baseUrl.endsWith("/")) baseUrl else "$baseUrl/"
        val config: HttpClientConfig<*>.() -> Unit = {
            install(ContentNegotiation) { json(json) }
            defaultRequest {
                // defaultRequest のブロックはリクエスト毎に評価される＝常に最新トークンを載せる。
                url(normalizedBase)
                contentType(ContentType.Application.Json)
                tokenStore.load()?.let { bearerAuth(it) }
            }
            expectSuccess = false // 失敗時の分岐は throwIfError で明示的に行う
        }
        if (engine != null) HttpClient(engine, config) else HttpClient(config)
    }

    // throwIfError は失敗応答を KarinError に正規化して投げる。成功なら何もしない。
    suspend fun throwIfError(response: HttpResponse) {
        if (response.status.isSuccess()) return
        throw when (response.status.value) {
            401 -> KarinError.Unauthorized
            403 -> KarinError.Forbidden
            404 -> KarinError.NotFound
            409 -> KarinError.Conflict
            in 500..599 -> KarinError.Server(response.status.value)
            else -> runCatching {
                val env = json.decodeFromString(ErrorEnvelope.serializer(), response.bodyAsText())
                KarinError.Api(env.error.code, env.error.message)
            }.getOrElse { KarinError.Server(response.status.value) }
        }
    }
}
