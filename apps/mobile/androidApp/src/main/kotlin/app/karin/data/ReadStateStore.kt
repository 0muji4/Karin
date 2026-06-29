package app.karin.data

import android.content.Context

// ReadStateStore は「どの風だよりを開封したか」を端末ローカルに持つ。
// バックエンドに既読の概念が無いため、未読の合図はこの開封済み集合から導出する。
// 機微情報ではないので暗号化はせず通常の SharedPreferences を使う。
interface ReadStateStore {
    fun openedIds(): Set<String>
    fun markOpened(id: String)
}

class PrefsReadStateStore(context: Context) : ReadStateStore {
    private val prefs = context.applicationContext.getSharedPreferences("karin_read_state", Context.MODE_PRIVATE)

    override fun openedIds(): Set<String> = prefs.getStringSet(KEY, emptySet()).orEmpty()

    override fun markOpened(id: String) {
        prefs.edit().putStringSet(KEY, openedIds() + id).apply()
    }

    private companion object {
        const val KEY = "opened_tanzaku_ids"
    }
}
