package app.karin.di

import android.content.Context
import app.karin.shared.api.DefaultKarinRepository
import app.karin.shared.api.KarinRepository
import app.karin.shared.auth.EncryptedTokenStore
import app.karin.shared.auth.TokenStore
import app.karin.shared.net.KarinClient
import app.karin.shared.session.SessionRepository

// AppContainer は手動 DI のコンテナ。小規模なので DI ライブラリは入れず、生成と配線をここに集約する。
// 規模が増えたら Koin 等へ移す（binding time を遅らせる）。
class AppContainer(context: Context) {
    val tokenStore: TokenStore = EncryptedTokenStore(context.applicationContext)
    val client: KarinClient = KarinClient(BASE_URL, tokenStore)
    val sessionRepository: SessionRepository = SessionRepository(client, tokenStore)
    val repository: KarinRepository = DefaultKarinRepository(client)

    companion object {
        // 接続先はビルド設定（BuildConfig.BASE_URL）で決まる。既定はエミュレータ用の 10.0.2.2、
        // 実機は local.properties の karinBaseUrl に LAN IP を書いて差し替える。
        val BASE_URL: String = app.karin.BuildConfig.BASE_URL
    }
}
