package app.karin.ui.component

import kotlin.test.Test
import kotlin.test.assertEquals

class VerticalTextTest {
    @Test
    fun 改行は列の区切りになる() {
        assertEquals(listOf("ああ", "いいい"), verticalColumns("ああ\nいいい", maxCharsPerColumn = 10))
    }

    @Test
    fun 列に収まらない分は次の列へ送る() {
        assertEquals(listOf("ああ", "あ"), verticalColumns("あああ", maxCharsPerColumn = 2))
    }

    @Test
    fun 改行と列送りが併存する() {
        // 1行目は2文字で収まり1列、2行目は4文字で2列に割れる。
        assertEquals(listOf("あい", "うえ", "おか"), verticalColumns("あい\nうえおか", maxCharsPerColumn = 2))
    }

    @Test
    fun 空行は空の列になる() {
        assertEquals(listOf("あ", "", "い"), verticalColumns("あ\n\nい", maxCharsPerColumn = 10))
    }

    @Test
    fun 不正な列長は分割しない() {
        assertEquals(listOf("あいう"), verticalColumns("あいう", maxCharsPerColumn = 0))
    }
}
