package app.karin.shared.ko

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNull
import kotlin.test.assertTrue

class KoCatalogTest {
    @Test
    fun 七十二候を網羅し番号が連続する() {
        assertEquals(72, KoCatalog.entries.size)
        assertEquals((1..72).toList(), KoCatalog.entries.map { it.number })
    }

    @Test
    fun 番号から候名と読みを引ける() {
        assertEquals("東風解凍", KoCatalog.name(1))
        assertEquals("はるかぜこおりをとく", KoCatalog.kana(1))
        assertTrue(KoCatalog[72] != null)
    }

    @Test
    fun 範囲外は_null_と空文字() {
        assertNull(KoCatalog[0])
        assertNull(KoCatalog[73])
        assertEquals("", KoCatalog.name(73))
    }
}
