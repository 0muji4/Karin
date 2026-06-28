package app.karin.shared.auth

// TokenStore は匿名アカウントの Bearer トークンを安全に保持するポート。
// トークンは発行時に1度だけ返り再取得できないため、保存の失敗＝アカウント喪失。
// 実装はプラットフォーム依存（Android=EncryptedSharedPreferences, iOS=Keychain）で、
// commonMain は interface だけを知る（expect/actual ではなく注入で差し替える）。
interface TokenStore {
    fun load(): String?
    fun save(token: String)
    fun clear()
}
