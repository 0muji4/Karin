package app.karin.shared.ko

// 七十二候の静的メタ（番号→名・読み）。バックエンドの ko_reference seed から生成し、内容を一致させる。
// 文箱で記録の ko_written（番号）から候名を解決するために使う（全候を返す API が無いため同梱）。
// 生成元: apps/backend/db/migrations/20260628000002_ko_reference_seed.up.sql

data class KoEntry(val number: Int, val name: String, val kana: String)

object KoCatalog {
    val entries: List<KoEntry> = listOf(
        KoEntry(1, "東風解凍", "はるかぜこおりをとく"),
        KoEntry(2, "黄鶯睍睆", "うぐいすなく"),
        KoEntry(3, "魚上氷", "うおこおりをいずる"),
        KoEntry(4, "土脉潤起", "つちのしょううるおいおこる"),
        KoEntry(5, "霞始靆", "かすみはじめてたなびく"),
        KoEntry(6, "草木萌動", "そうもくめばえいずる"),
        KoEntry(7, "蟄虫啓戸", "すごもりむしとをひらく"),
        KoEntry(8, "桃始笑", "ももはじめてさく"),
        KoEntry(9, "菜虫化蝶", "なむしちょうとなる"),
        KoEntry(10, "雀始巣", "すずめはじめてすくう"),
        KoEntry(11, "桜始開", "さくらはじめてひらく"),
        KoEntry(12, "雷乃発声", "かみなりすなわちこえをはっす"),
        KoEntry(13, "玄鳥至", "つばめきたる"),
        KoEntry(14, "鴻雁北", "こうがんかえる"),
        KoEntry(15, "虹始見", "にじはじめてあらわる"),
        KoEntry(16, "葭始生", "あしはじめてしょうず"),
        KoEntry(17, "霜止出苗", "しもやみてなえいずる"),
        KoEntry(18, "牡丹華", "ぼたんはなさく"),
        KoEntry(19, "蛙始鳴", "かわずはじめてなく"),
        KoEntry(20, "蚯蚓出", "みみずいずる"),
        KoEntry(21, "竹笋生", "たけのこしょうず"),
        KoEntry(22, "蚕起食桑", "かいこおきてくわをはむ"),
        KoEntry(23, "紅花栄", "べにばなさかう"),
        KoEntry(24, "麦秋至", "むぎのときいたる"),
        KoEntry(25, "螳螂生", "かまきりしょうず"),
        KoEntry(26, "腐草為螢", "くされたるくさほたるとなる"),
        KoEntry(27, "梅子黄", "うめのみきばむ"),
        KoEntry(28, "乃東枯", "なつかれくさかるる"),
        KoEntry(29, "菖蒲華", "あやめはなさく"),
        KoEntry(30, "半夏生", "はんげしょうず"),
        KoEntry(31, "温風至", "あつかぜいたる"),
        KoEntry(32, "蓮始開", "はすはじめてひらく"),
        KoEntry(33, "鷹乃学習", "たかすなわちわざをならう"),
        KoEntry(34, "桐始結花", "きりはじめてはなをむすぶ"),
        KoEntry(35, "土潤溽暑", "つちうるおうてむしあつし"),
        KoEntry(36, "大雨時行", "たいうときどきにふる"),
        KoEntry(37, "涼風至", "すずかぜいたる"),
        KoEntry(38, "寒蝉鳴", "ひぐらしなく"),
        KoEntry(39, "蒙霧升降", "ふかききりまとう"),
        KoEntry(40, "綿柎開", "わたのはなしべひらく"),
        KoEntry(41, "天地始粛", "てんちはじめてさむし"),
        KoEntry(42, "禾乃登", "こくものすなわちみのる"),
        KoEntry(43, "草露白", "くさのつゆしろし"),
        KoEntry(44, "鶺鴒鳴", "せきれいなく"),
        KoEntry(45, "玄鳥去", "つばめさる"),
        KoEntry(46, "雷乃収声", "かみなりすなわちこえをおさむ"),
        KoEntry(47, "蟄虫坏戸", "むしかくれてとをふさぐ"),
        KoEntry(48, "水始涸", "みずはじめてかるる"),
        KoEntry(49, "鴻雁来", "こうがんきたる"),
        KoEntry(50, "菊花開", "きくのはなひらく"),
        KoEntry(51, "蟋蟀在戸", "きりぎりすとにあり"),
        KoEntry(52, "霜始降", "しもはじめてふる"),
        KoEntry(53, "霎時施", "こさめときどきふる"),
        KoEntry(54, "楓蔦黄", "もみじつたきばむ"),
        KoEntry(55, "山茶始開", "つばきはじめてひらく"),
        KoEntry(56, "地始凍", "ちはじめてこおる"),
        KoEntry(57, "金盞香", "きんせんかさく"),
        KoEntry(58, "虹蔵不見", "にじかくれてみえず"),
        KoEntry(59, "朔風払葉", "きたかぜこのはをはらう"),
        KoEntry(60, "橘始黄", "たちばなはじめてきばむ"),
        KoEntry(61, "閉塞成冬", "そらさむくふゆとなる"),
        KoEntry(62, "熊蟄穴", "くまあなにこもる"),
        KoEntry(63, "鱖魚群", "さけのうおむらがる"),
        KoEntry(64, "乃東生", "なつかれくさしょうず"),
        KoEntry(65, "麋角解", "さわしかのつのおつる"),
        KoEntry(66, "雪下出麦", "ゆきわたりてむぎのびる"),
        KoEntry(67, "芹乃栄", "せりすなわちさかう"),
        KoEntry(68, "水泉動", "しみずあたたかをふくむ"),
        KoEntry(69, "雉始雊", "きじはじめてなく"),
        KoEntry(70, "款冬華", "ふきのはなさく"),
        KoEntry(71, "水沢腹堅", "さわみずこおりつめる"),
        KoEntry(72, "鶏始乳", "にわとりはじめてとやにつく"),
    )

    private val byNumber: Map<Int, KoEntry> = entries.associateBy { it.number }

    operator fun get(number: Int): KoEntry? = byNumber[number]
    fun name(number: Int): String = byNumber[number]?.name ?: ""
    fun kana(number: Int): String = byNumber[number]?.kana ?: ""
}
