# saveArticleOnSlack
- go 1.12
- GCP Cloud Functions
- GCP DataStore
- slack スラッシュコマンドで技術記事などをタグ付けして保村とリスト表示できる Go app

## GCP projectID
- saveArticleOnSlack

## go moduls
下記のように環境変数の設定と、go modulesの初期化が必要。
CloudFunctionsではgo modulesによるバージョン管理が必要なため
```
$ export GO111MODULE=on
$ go mod init
```
## deploy
```bash
#初回
gcloud beta functions deploy saveArticleOnSlack --runtime go111 --entry-point SaveArticleOnSlack --trigger-http
#deployするだけなら
gcloud beta functions deploy saveArticleOnSlack
```
- 環境変数設定
```bash
--set-env-vars 変数名=値
--update-env-vars 変数名=値
```

## DataStore index
- indexを作成しないとFilterメソッドを使ってSQLのように条件検索ができない。
- index.yamlに定義
```bash
# index作成
  gcloud datastore indexes create index.yaml
# index削除
  gcloud datastore indexes cleanup index.yaml
```