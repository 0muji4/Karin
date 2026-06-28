package app.karin

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier

import app.karin.nav.KarinNavHost
import app.karin.ui.theme.KarinTheme

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        val container = (application as KarinApplication).container
        setContent {
            KarinTheme {
                Surface(modifier = Modifier.fillMaxSize()) {
                    KarinNavHost(container)
                }
            }
        }
    }
}
