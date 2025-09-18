golangci-lint v2.4.0の設定エラーを修正しました。

  修正内容：
  1. linters.disable-all: true → linters.default: none に変更
  2. linters-settings → linters.settings (lintersセクション内)
  に移動
  3. 無効な設定プロパティを削除・修正：
    - nolintlint.allow-leading-space を削除
    - revive.ignore-generated-header を削除
    - staticcheck.go → staticcheck.checks: [all] に変更

  確認結果：
  - golangci-lint config verify でスキーマ検証が成功
  - golangci-lint run で正常に実行完了
  - 詳細出力で指定した9つのリンターが正しく有効化されていることを
  確認

  これで、GitHub CIでgolangci-lintが実行される際にスキーマバリデー
  ションエラーが解消され、期待通りのカスタム設定（gocycloの複雑度1
  5など）が適用されるようになります。

> 通った。今まで一体何を四苦八苦してたんだ？教えて欲しい。最大の原因
  はどこだった？

⏺ 今回の問題の経緯を振り返ると、実は二つの異なる問題が混在していて
  、それが混乱の原因でした。

  問題の経緯と最大の原因

  1. 最初の問題（P1）: モジュールダウンロードの欠如

  - 原因: lintジョブにgo mod downloadステップがなかった
  - 症状: modules-download-mode: readonly設定下でGitHub
  Runnerでの依存関係エラー
  - 解決: .github/workflows/ci.ymlにlintジョブ用のgo mod
  downloadステップを追加

  2. 二番目の問題（P1）: 設定スキーマの不一致

  - 最大の原因: golangci-lint
  v1.x系の設定形式をv2.4.0で使用していた
  - 具体的な間違い:
    - linters.disable-all: true ← v1形式
    - 最上位のlinters-settings ← v1形式
  - 正解:
    - linters.default: none ← v2形式
    - linters.settings ← v2形式（lintersセクション内）

  混乱の要因

  1. フィードバックの解釈ミス: 最初のフィードバックで「linters.def
  aultとsettingsブロックが無効」と言われた時、v2の正しい形式だと思
  い込んで逆方向（v1形式）に修正してしまった
  2. バージョン間の設定形式変更: golangci-lint
  v1→v2で設定スキーマが大幅変更されたが、その認識が不十分だった
  3. ローカルとCI環境の差異: ローカルでは古いバージョンで動作して
  いたが、CI環境のv2.4.0では厳格なスキーマ検証でエラーになった

  結論

  **最大の原因は「golangci-lint v2.4.0の設定スキーマ変更を理解せず
  に、v1形式の設定を使い続けていたこと」**でした。特にdisable-all
  と最上位linters-settingsがv2では無効になったことが決定的でした。

> エラーメッセージの紛らわしさがあったのか？それかv2が新しすぎてデー
  タが少なかったか？

⏺ はい、両方の要因がありました。分析してみます：

  エラーメッセージの紛らわしさ

  1. 最初のフィードバック（P1）の誤解を招く表現

  The lint configuration was rewritten to use linters.default and
  a nested settings block,
  but golangci-lint's schema expects disable-all under linters and
   a top-level linters-settings.

  この文章は非常に紛らわしい構造でした：
  - 「新しく書き換えた設定（linters.default）は間違い」
  - 「正しくはdisable-allと最上位linters-settings」

  実際には逆で、v2ではlinters.defaultが正解だったのに、このフィー
  ドバックがv1形式を推奨しているように読めました。

  2. エラーメッセージの解釈困難

  jsonschema: "linters" does not validate with
  "/properties/linters/additionalProperties":
  additional properties 'disable-all' not allowed

  このJSONスキーマエラーは：
  - どのバージョンのスキーマなのか不明
  - 正しい代替案が示されていない
  - 技術的すぎて「なぜ無効なのか」の文脈が不足

  v2.4.0の情報不足

  1. 検索結果の問題

⏺ Web Search("golangci-lint v2.4.0 release date documentation
            availability")
  ⎿  Did 1 search in 13s

⏺ v2.4.0の新しさによる情報不足の確認

  1. リリース時期の問題

  - v2.4.0は2025年8月14日リリース -
  つまり約1ヶ月前の非常に新しいバージョン
  - 私のナレッジカットオフ（2025年1月）より後のリリース

  2. 検索時の問題点

  - Web検索でも具体的な設定例が少なかった
  - 公式ドキュメントでも個別リンター設定の詳細が不完全
  - v1からv2への移行情報はあるが、v2.4.0特有の変更点の情報が断片的

  3. ドキュメントの構造的問題

  WebFetchで確認した公式ドキュメントでも：
  nolintlint: No configuration details were provided
  revive: No configuration details were provided
  staticcheck: No configuration details were provided

  つまり、個別リンターの設定スキーマが十分に文書化されていない状況
  でした。

  結論

  両方が原因でした：

  1. 紛らわしいフィードバック:
  「v2の正しい設定が間違い、v1形式が正解」という逆の指示
  2. v2.4.0の新しさ:
  リリース1ヶ月という新しさで、設定例や解決事例の蓄積が不足
  3. ドキュメント不備: 公式でも個別リンター設定の詳細が不完全

  特に「フィードバックの文章構造が逆方向を示唆していた」ことが最大
  の混乱要因でした。
