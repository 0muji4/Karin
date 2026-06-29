package app.karin.shared.session

import app.karin.shared.api.AccountResponse
import app.karin.shared.api.createAccount
import app.karin.shared.auth.TokenStore
import app.karin.shared.net.KarinClient

// SessionRepository は匿名アカウントの発行と在席判定を担う。トークンは1度だけ返るため、
// 発行に成功したら即座に保存する。この「発行→保存」を不可分の一手として扱うのが要点で、
// プラットフォーム非依存なので shared に置き、Android/iOS で共有する。
class SessionRepository(
    private val client: KarinClient,
    private val tokenStore: TokenStore,
) {
    // issueAccount は新規アカウントを発行し、返ったトークンを保存して返す。
    // API が失敗した場合はトークンを保存しない（中途半端な在席状態を作らない）。
    suspend fun issueAccount(): AccountResponse {
        val res = client.createAccount()
        tokenStore.save(res.token)
        return res
    }

    // isSignedIn はトークンを保持しているか（＝起動時に本編へ直行してよいか）。
    fun isSignedIn(): Boolean = tokenStore.load() != null
}
