package app.karin.shared.api

import app.karin.shared.net.KarinClient
import app.karin.shared.net.KarinError
import io.ktor.client.call.body
import io.ktor.client.request.get
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse

// 記録 MVP が使う4エンドポイント。KarinClient の上に拡張関数として薄く載せる。
// 成功・失敗の正規化は KarinClient.throwIfError、通信/解釈の失敗は KarinError に寄せる。

// 匿名アカウントを発行する（トークンは1度だけ返る。呼び手が TokenStore へ保存する）。
suspend fun KarinClient.createAccount(): AccountResponse =
    call { http.post("accounts") }

// 今日の候を取る（認証不要）。
suspend fun KarinClient.todayKo(): TodayResponse =
    call { http.get("ko/today") }

// 短冊を文箱に保存する（ko_written 省略時はサーバが今日の候を使う）。
suspend fun KarinClient.createRecord(req: CreateRecordRequest): RecordDto =
    call { http.post("records") { setBody(req) } }

// 文箱（節気ごとにまとめた記録）を取る。
suspend fun KarinClient.listBox(): BoxResponse =
    call { http.get("box") }

// 記録を風に乗せる（公開プールへ流す）。応答は一律 status="cast"。危機判定時のみ support が付く。
suspend fun KarinClient.castToWind(recordId: String): CastResponse =
    call { http.post("records/$recordId/cast") }

// call は送受信→エラー正規化→本文デコードを一括で行う。通信失敗は Network、解釈失敗は Decode に寄せる。
private suspend inline fun <reified T> KarinClient.call(block: () -> HttpResponse): T {
    val response = try {
        block()
    } catch (e: KarinError) {
        throw e
    } catch (e: Throwable) {
        throw KarinError.Network(e)
    }
    throwIfError(response)
    return try {
        response.body()
    } catch (e: Throwable) {
        throw KarinError.Decode
    }
}
