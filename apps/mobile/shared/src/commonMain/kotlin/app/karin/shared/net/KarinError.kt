package app.karin.shared.net

import kotlinx.serialization.Serializable

// KarinError は関門（バックエンド）からの失敗を、UI が分岐できる形に正規化したもの。
// HTTP ステータスの意味（401=トークン無効/欠落, 403=利用停止, 404, 409=二重）と、
// バックエンドの {"error":{code,message}} 包みを区別して持つ。
sealed class KarinError : Exception() {
    data object Unauthorized : KarinError()
    data object Forbidden : KarinError()
    data object NotFound : KarinError()
    data object Conflict : KarinError()
    data class Server(val status: Int) : KarinError()
    data class Api(val code: String, val detail: String) : KarinError()
    data class Network(override val cause: Throwable?) : KarinError()
    data object Decode : KarinError()
}

// ErrorEnvelope はバックエンド共通のエラー応答 {"error":{"code","message"}} に対応する。
@Serializable
data class ErrorEnvelope(val error: ErrorBody)

@Serializable
data class ErrorBody(val code: String, val message: String)
