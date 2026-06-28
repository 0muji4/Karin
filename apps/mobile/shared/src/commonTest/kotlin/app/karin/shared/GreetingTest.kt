package app.karin.shared

import kotlin.test.Test
import kotlin.test.assertTrue

class GreetingTest {
    @Test
    fun greeting_is_not_empty() {
        assertTrue(greeting().isNotEmpty())
    }
}
