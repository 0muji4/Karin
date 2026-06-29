package app.karin.shared.api

import kotlin.test.Test
import kotlin.test.assertEquals

class CardTransportTest {
    @Test
    fun encode_してdecodeすると元のカードに戻る() {
        val card = ReceivedCard(
            tanzakuId = "t-1",
            body = "夕立のあと、虹。",
            ko = 29,
            isOfficial = false,
            deliveredOn = "2026-06-29",
            kept = false,
        )
        assertEquals(card, decodeReceivedCard(card.encode()))
    }

    @Test
    fun 改行や記号を含む本文も往復で保たれる() {
        val card = ReceivedCard(
            tanzakuId = "t-2",
            body = "ひぐらし、\n夕暮れ。\"引用\"も。",
            ko = 1,
            isOfficial = true,
            deliveredOn = "2026-08-12",
            kept = true,
        )
        assertEquals(card, decodeReceivedCard(card.encode()))
    }
}
