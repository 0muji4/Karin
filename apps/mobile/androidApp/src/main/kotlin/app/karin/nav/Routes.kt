package app.karin.nav

// 画面遷移のルート。記録 MVP の範囲を列挙し、交換系は後続で足す。
object Routes {
    const val SPLASH = "splash" // 起動
    const val ONBOARDING = "onboarding"
    const val HOME = "home" // 今日の候
    const val WRITE = "write" // 短冊を書く
    const val CAST = "cast" // 風に乗せる（確認）
    const val SENT = "sent" // 風に乗りました（完了）
    const val DELIVERIES = "deliveries" // 風だより（受信一覧）
    const val CHIME = "chime" // 風鈴が鳴る（受信を開く）
    const val BOX = "box" // 文箱
}
