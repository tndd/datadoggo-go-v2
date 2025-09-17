# datadoggo-go-v2
あらかじめ用意したrssフィードから記事を取得し保存するアプリ。

# テーブル定義
Option指定なき場合、NOT NULL制約とする。

## ArticleLink
記事のリンク

| name     | type       | description                                   |
| -------- | ---------- | --------------------------------------------- |
| url      | text       | URLを主キーとする                             |
| title    | text       | 記事のタイトル                                |
| source   | text       | どこから取得されたリンクかを表す              |
| pub_data | timestampz | 記事の公開日時。ニュースという特性上UTCを使う |

## ArticleContent
記事の内容を保存するテーブル。linkとjoinして使う。

| name        | type       | description                                |
| ----------- | ---------- | ------------------------------------------ |
| url         | str(FK)    | linkのurlを外部キーとする                  |
| created_at  | timestampz | 作成日時。デフォルトは現在時刻             |
| updated_at  | timestampz | 更新日時。デフォルトは現在時刻             |
| status_code | int        | HTTPステータスコード                       |
| content     | text       | 記事の内容。取得に失敗しても空文字は入れる |

# ドメインモデル
## ArticleUrlStatus
URLごとの取得状況を確認するためのモデル。

| name        | type          | description                                                            |
| ----------- | ------------- | ---------------------------------------------------------------------- |
| url         | text          | URLを主キーとする                                                      |
| status      | text          | 取得状況。成功か失敗かを表す                                           |
| pub_data    | timestampz    | 記事の公開日時                                                         |
| status_code | int(Optional) | HTTPステータスコード。未実行のものがjoinされる可能性があるためOptional |


## Article
ArticleLinkにArticleContentをjoinしたもの。
ユーザーから見ればsourceは関心ごとではない。createとupdateも同じく。status_codeについては、ユーザーははじめから成功した記事を要求してるのだから、この項目は不要。contentも同じく失敗記事が要求されることはないためOptional指定を外している。

| name     | type       | description                             |
| -------- | ---------- | --------------------------------------- |
| url      | text       | URLを主キーとする                       |
| title    | text       | 記事のタイトル                          |
| pub_data | timestampz | 公開日時。ニュースという特性上UTCを使う |
| content  | text       | 記事の内容                              |

# 関数
## データ収集&保存
- feeds.yml -> list[str]
  - load_rss_urls(group: Optional[str]) -> list[str]
    - groupはfeeds.ymlのkeyを指定する。指定しない場合は全てのURLを返す。
- url -> Feed
  - fetch_feed(url: str) -> Feed
- Feed -> ArticleLink[list]
  - get_article_links(rss: Feed) -> ArticleLink[list]
    - Feedはxml形式のデータ(ex. parser.ParseURL("https://zenn.dev/spiegel/feed"))
  - store_article_links(article_links: ArticleLink[list]) -> None
- ArtileLink -> ArticleContent
  - fetch_article_content(url: str) -> ArticleContent
  - store_article_content(article_content: ArticleContent) -> None

## データ取得
- ArticleUrlStatus
  - search_article_url_status(query: ArticleUrlQuery) -> ArticleUrlStatus[list]
- Article
  - search_article(query: ArticleQuery) -> Article[list]

## ワークフロー
- feeds.ymlを元にArticleLinkを取得しDBに保存する
- ArticleLinkについて、未取得か失敗した内容のArticleContentを取得しDBに保存する

# 開発における注意
- まだ使用されることのない関数を実装しないこと。実装は必要になってから

## テスト実装
- 機能の実装とテストは常にペアで行うこと
- 必ずテスト通過が確認されてから次の機能実装に進むように

## コミット
- コミットメッセージは日本語で行う
- 機能ごとにdevelopからブランチを切る
  - feat/{}
  - refactor/{}
  - fix/{}
  - ...
- コミットの粒度は機能とテストの1つの固まりごとに行う
  - 複数の関数や機能を横着に一斉実装しないように。後で変更が意味不明になることを防ぐため。
- ブランチにおける目的の実装が完了し、全テストが通過したら、developにmergeする。(no ff)
- そして次の機能実装においては新たなdevelopブランチを切る。この繰り返し。

### 参考: 階層的コミット戦略
上の説明の実例として以下を挙げる。

**要約**
- フィーチャーブランチ内では構造体やサービスクラスなど、論理的に独立したコンポーネント単位でコミットを分ける。
- フィーチャー完了時はdevelopに全体を統合し、develop履歴では大きな機能追加として記録される
  - ブランチ = 大きな意味のある機能
  - コミット = その機能を構成する独立したコンポーネント
  - マージ = developブランチに意味のある変更の塊として記録を行う行為

**レベル1: フィーチャーブランチ（抽象的な意味のある塊）**
- feat/article-models
- feat/database-layer
- feat/rss-processing

**レベル2: 機能単位コミット（そのフィーチャーを構成する意味のある部品）**

feat/article-models ブランチ内で:
- ArticleLink構造体とそのテスト実装
- ArticleContent構造体とそのテスト実装
- ArticleUrlStatus構造体とそのテスト実装
- Article構造体とそのテスト実装

**レベル3: developへのマージ（完成した機能として統合）**

develop履歴:
- feat: ドメインモデル群の実装 (merge feat/article-models)
- feat: データベースアクセス層の実装 (merge feat/database-layer)


## クローリングについて
**記事取得**:
- 記事の取得はfirecrawlを用いるものとする。
- バージョンは不都合が生じない限りは[v2](https://github.com/firecrawl/firecrawl-go)を使うこと。
- 開発を進めるにあたってはモックを使って開発を行うこと。(通信の実行は最後の結合テストでのみ行う)
- モックの内容は`mock/firecrawl/`下のjsonファイルを参照。

**rss取得**:
- rss取得についてもモックを使うように。
- モックの内容は`mock/rss/`下のrssファイルを参照すること。

# Todo
