package app.karin.shared.api

import app.karin.shared.net.KarinClient

// KarinRepository は UI(ViewModel) とネットワーク層の境界。ViewModel はこの interface に依存し、
// テストでは fake に差し替える。実装は KarinClient の薄いラッパで、将来のキャッシュ等の置き場にもなる。
// 認証/トークン保存を伴うアカウント系は SessionRepository が担うため、ここには含めない。
interface KarinRepository {
    suspend fun todayKo(): TodayResponse
    suspend fun createRecord(body: String, koWritten: Int? = null): RecordDto
    suspend fun listBox(): BoxResponse
    suspend fun cast(recordId: String): CastResponse
    suspend fun listDeliveries(): DeliveriesResponse
    suspend fun keep(tanzakuId: String): StatusResponse
    suspend fun report(tanzakuId: String, reason: String, note: String): StatusResponse
}

class DefaultKarinRepository(private val client: KarinClient) : KarinRepository {
    override suspend fun todayKo(): TodayResponse = client.todayKo()

    override suspend fun createRecord(body: String, koWritten: Int?): RecordDto =
        client.createRecord(CreateRecordRequest(body = body, koWritten = koWritten))

    override suspend fun listBox(): BoxResponse = client.listBox()

    override suspend fun cast(recordId: String): CastResponse = client.castToWind(recordId)

    override suspend fun listDeliveries(): DeliveriesResponse = client.listDeliveries()

    override suspend fun keep(tanzakuId: String): StatusResponse = client.keep(tanzakuId)

    override suspend fun report(tanzakuId: String, reason: String, note: String): StatusResponse =
        client.report(tanzakuId, reason, note)
}
