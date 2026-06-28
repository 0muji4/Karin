package app.karin

import android.app.Application
import app.karin.di.AppContainer

// KarinApplication はプロセス唯一の合成ルート。AppContainer をここで一度だけ作る。
class KarinApplication : Application() {
    lateinit var container: AppContainer
        private set

    override fun onCreate() {
        super.onCreate()
        container = AppContainer(this)
    }
}
