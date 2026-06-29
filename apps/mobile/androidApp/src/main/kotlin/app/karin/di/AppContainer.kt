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
        // debug 既定はエミュレータ→ホストのループバック。実機/本番ではビルド設定で差し替える。
        const val BASE_URL = "http://10.0.2.2:8080/"
    }
}
