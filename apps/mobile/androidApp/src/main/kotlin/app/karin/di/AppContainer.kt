package app.karin.di

import android.content.Context
import app.karin.shared.auth.EncryptedTokenStore
import app.karin.shared.auth.TokenStore
import app.karin.shared.net.KarinClient

// AppContainer は手動 DI のコンテナ。小規模なので DI ライブラリは入れず、生成と配線をここに集約する。
// 規模が増えたら Koin 等へ移す（binding time を遅らせる）。
class AppContainer(context: Context) {
    val tokenStore: TokenStore = EncryptedTokenStore(context.applicationContext)
    val client: KarinClient = KarinClient(BASE_URL, tokenStore)

    // アカウント（＝トークン）を保持しているか。起動時の遷移先判定に使う。
    fun hasAccount(): Boolean = tokenStore.load() != null

    companion object {
        // debug 既定はエミュレータ→ホストのループバック。実機/本番ではビルド設定で差し替える。
        const val BASE_URL = "http://10.0.2.2:8080/"
    }
}
