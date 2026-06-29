package app.karin.shared.api

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// バックエンドの公開 JSON 契約に対応する DTO。キーは実装（apps/backend/internal/api）に厳密一致させる。
// 時刻・日付はサーバが RFC3339/日付文字列で返すため、表示整形は UI 側に委ね、ここでは String で持つ。

@Serializable
data class AccountResponse(
    @SerialName("user_id") val userId: String,
    val token: String,
)

@Serializable
data class WafuMonthDto(val name: String, val kana: String)

@Serializable
data class SekkiDto(val number: Int, val name: String, val kana: String)

@Serializable
data class KoDto(
    val number: Int,
    val name: String,
    val kana: String,
    val meaning: String,
    val sekki: Int,
    val season: String,
)

@Serializable
data class TodayResponse(
    val date: String,
    @SerialName("wafu_month") val wafuMonth: WafuMonthDto,
    val sekki: SekkiDto,
    val ko: KoDto,
)

@Serializable
data class CreateRecordRequest(
    val body: String,
    @SerialName("ko_written") val koWritten: Int? = null,
)

@Serializable
data class RecordDto(
    val id: String,
    @SerialName("ko_written") val koWritten: Int,
    val body: String,
    @SerialName("created_at") val createdAt: String,
)

@Serializable
data class BoxGroupDto(
    @SerialName("wafu_month") val wafuMonth: WafuMonthDto,
    val sekki: SekkiDto,
    val records: List<RecordDto>,
)

@Serializable
data class BoxResponse(val groups: List<BoxGroupDto>)

// 風に乗せる応答。判定結果は著者に見せないため一律 status="cast"。
// 例外として危機（自傷）と判定したときだけ support（支援先の案内）が付く。
@Serializable
data class CastResponse(val status: String, val support: SupportInfo? = null)

@Serializable
data class SupportInfo(val message: String, val url: String)
