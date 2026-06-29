package app.karin.shared.api

import kotlinx.serialization.decodeFromString
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.Json

// 受信カード(ReceivedCard)を画面遷移で受け渡すための文字列変換。
//
// Android の SavedStateHandle は Bundle へ格納できる型（Parcelable/プリミティブ等）しか保持できず、
// kotlinx の @Serializable DTO をそのまま渡すと実行時に IllegalArgumentException で落ちる。
// 受け渡しは文字列に統一し、DTO を JSON 文字列に変換する処理は共有層に集約する
// （iOS の状態復元でも同じ制約に当たるため、両 OS から再利用できる場所に置く）。
private val cardJson = Json { ignoreUnknownKeys = true }

fun ReceivedCard.encode(): String = cardJson.encodeToString(this)

fun decodeReceivedCard(value: String): ReceivedCard = cardJson.decodeFromString(value)
