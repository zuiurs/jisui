# jisui

自炊した画像を処理するためのツールです．

## Prerequisites

- Docker
  - Windows, Mac 環境では必須
  - Linux では [Go Imagick v2]() が動く環境であればビルド可能です
    - CentOS とかではデフォルトで v1 用のライブラリが入るため面倒です
  - Windows
    - https://docs.docker.com/docker-for-windows/install
    - Memo: Hyper-V を裏で使用
  - Mac
    - `brew install docker`
    - Memo: xhyve を裏で使用
  - Linux
    - https://docs.docker.com/engine/installation

## Installation

Clone repository (not `go get`).

```
git clone https://github.com/zuiurs/jisui.git
```

Building container.

```
cd jisui
make container
```

## Usage

Docker コンテナで動かすことを想定しているため若干使い方が特殊です．

クローンしたリポジトリに適当なデータ置き場を作成します．

```
mkdir work
# copy your comic data to work
```

コンテナを起動してその中で作業します．

```
make run
```

起動時に行われる作業は下記になります．起動直後からすぐ最新のコマンドが使えるようになっています．

- 現在のディレクトリのマウント
- ソースコードのコンパイル & インストール
- マウントポイントでの bash 起動

あとは普通に使ってください．

```
jisui
```

### Other usage

`example.jpg` をカレントディレクトリの `example.png` として変換、保存します。

```
jisui comic example.jpg
```

`example.jpg` をカレントディレクトリの `example_resized.png` として Height が 1200 になるようにアス比を保ちつつ変換、保存します。

```
jisui comic -o example_resized.png -h 1200 example.jpg
```

`example` ディレクトリの中を一枚ずつ処理します。各ファイルの命名規則は ファイルの時の処理と同じです。

```
jisui comic example
```

`example` ディレクトリの中を一枚ずつ処理して `example_resized` ディレクトリに格納します。

```
jisui comic -o example_resized -h 1200 example
```

`example` ディレクトリの中を一枚ずつ処理した結果を `example.pdf` として PDF 化します。

```
jisui comic -pack -h 1200 example
```

`example` ディレクトリの中を一枚ずつ処理した結果を `example_resized.pdf` として PDF 化します。

```
jisui comic -pack -h 1200 -o example_resized.pdf example
```

`example` ディレクトリの中を一枚ずつ iPad Air2 の縦に合わせてリサイズしつつ、指定ページ番号はモノクロ化せずに `example_resized.pdf` として PDF 化します。

```
jisui comic -h 1536 -pack -skip 1,2,20-24 -o example_resized.pdf example
```

## Routine

自炊の流れを書いておきます．

1. 漫画のカバーと帯を取る

2. カバーと帯はシリーズまとめて ScanSnap iX500 のカバー用コンフィグでスキャン

3. Curl DC-200N で表紙折込，表紙 & 背表紙，裏表紙，裏表紙折込の 4 つに分割

4. 漫画本体を Durodex 200DX で裁断
    - 5mm くらいで攻めても大丈夫だが本体表紙の糊が付きやすい
    - 中央部分まで台詞がある漫画は多いため，攻めないと負ける
    - コツ: 本を固定したら裁断時に上から押さえつけながら素早く断つ．そうしないと切断部分が下に行くに従って多くなる．
    - コツ: 裁断後はパラパラすること．ページの付着を取り除かないとぐちゃぐちゃになる．

5. ScanSnap iX500 の漫画用コンフィグでスキャン
    - 表紙 & 背表紙 → 表紙折込 → 漫画本体 → 裏表折込 → 裏表紙
    - コツ: ずらして入れると 100 枚くらいスキャンできます

6. スキャン結果のディレクトリに移り，カバー系と漫画本体の表紙・裏表紙の裏面 (白紙) を手作業削除 (計 6 枚)
    - 漫画本体の表紙・裏表紙の裏面は本来カバーの折込にあたるため

7. 2. でスキャンした全体カバーと帯を適当に大きなナンバリングにして入れる

8. `jisui prepare <directory>` でゴミ削除，ナンバリング調整をする

この時点で Row なデータの取り込みは完了になります．コミカライズ用の手順は下記になります．

1. `jisui comic <directory>` で漫画用のデータに変換します
    - TODO: ここらへんの調整とかしやすくする

### Hardware

使用する機種は次の 3 つです．

- Durodex 200DX
  - https://www.amazon.co.jp/dp/B00A378TNU
- Curl DC-200N
  - https://www.amazon.co.jp/dp/B005GICA5Y
- ScanSnap iX500
  - https://www.amazon.co.jp/dp/B00T2B5L52

使用している機種のコンフィグです．

### ScanSnap iX500 Configuration

- 共通設定
  - アプリ選択
    - \[アプリケーションの選択]: [起動しません(ファイル保存のみ)\]
  - ファイル形式
    - \[ファイル形式の選択]: [JPEG (\*.jpg)\]
  - 原稿
    - \[原稿サイズの選択]: [サイズ自動検出\]
    - \[マルチフィード検出]: [重なりで検出 (超音波)\]
  - ファイルサイズ
    - \[圧縮]: [1\] 

- カバー・帯用設定
  - 保存先
    - \[イメージの保存先\]: `C:\path\to\cover\result`
    - ファイル名の設定
      - \[\*] [自分で名前をつけます\]
      - \[先頭文字列\]: `title_`
      - \[連番]: [1桁\]
      - 例: shingeki_1.jpg
  - 読み取りモード
    - \[画質の選択]: [スーパーファイン(カラー/グレー: 300dpi，白黒: 600dpi相当)\]
    - \[カラーモードの選択]: [カラー\]
    - \[読み取り面の選択]: [片面読み取り\]
    - \[向きの選択]: [回転しない\]
    - \[\_] [白紙ページを自動的に削除します\]
    - \[\*] [継続読み取りを有効にします\]
    - \[オプション\] は全てチェックを外す

- 漫画本体用設定
  - 保存先
    - \[イメージの保存先\]: `C:\path\to\main\result`
    - ファイル名の設定
      - \[\*] [自分で名前をつけます\]
      - \[先頭文字列\]: `title_number_`
      - \[連番]: [3桁\]
      - 例: shingeki_1_001.jpg
  - 読み取りモード
    - \[画質の選択]: [エクセレント(カラー/グレー: 600dpi，白黒: 1200dpi相当)\]
    - \[カラーモードの選択]: [カラー\]
    - \[読み取り面の選択]: [両面読み取り\]
    - \[向きの選択]: [回転しない\]
    - \[\_] [白紙ページを自動的に削除します\]
    - \[\*] [継続読み取りを有効にします\]
    - \[オプション\] は全てチェックを外す

## Image Processing

画像処理でやっていることを説明します。

### jisui comic

- 漫画の黄ばみを取り除くために Red Channel を抽出してグレースケール化します
  - 濃いシミの点があったりしたときに効果的です (薄い時はレベル補正でだいたい消えます)
- 色レベルを黒 `40%`、白 `85%` で補正します
  - 色々変えてみて最終的に行き着いた値です
  - 黒レベルを低くすると画像のシャープ感が得られません。高くするとトーンが必要以上に黒くなります
  - 白レベルを低くするとルビが細くなりすぎて消える可能性があります。高くすると裏写りが消えません
- 画像を Mitchell Filter でリサイズします
  - Box と Mitchell で検証しましたが、前者の方が縮小した時のトーンが点々に見えて汚いです
    - Mitchell はブラーの値を 1 以上にしないと結果が真っ黒になりますが、Box のブラー 0 はかなり汚いので 1 にしたときで比較しました
  - Mitchell にした理由は上と、ここの AutoDesk のドキュメントで大体 Mitchell が良くなると書いてあったからです
    - https://knowledge.autodesk.com/ja/support/3ds-max/learn-explore/caas/CloudHelp/cloudhelp/2017/JPN/3DSMax/files/GUID-DBFFF24F-5419-492B-8889-24E546029279-htm.html

## TODO

- ImageFormat の設定の見直し・検証
  - `png`、`pdf` に変更するタイミングは適当？
    - モノクロ化するときは未変更で、リサイズする時に `png` に変換している
    - PDF 化する時に `pdf` に変換している
  - やる必要はある？
- ディレクトリ処理時に毎回 ImageWand を生成する必要はある？
  - PDF みたいに `WriteImages()` の `adjoin` を `false` にしたら一枚一枚書き込んでくれる？ (未検証)

## About /scripts

昔使っていた自炊用のシェルスクリプトです．Go のツールがあるため今はほとんど使いません．

ImageMagick を実行できる環境が必要です．

基本的に原本はカラーの最高画質，圧縮なしで取り込み，あとからこのスクリプト群で調整します．

### pack.sh

下記のフォーマットの連番で管理されているファイルの番号を切り詰めます．

フォーマット: 最長マッチでアンダースコアを検索．3 桁の 0 パディングされた連番のあとに拡張子がついているファイル．

```
.*_XXX\.jpg
```

両面で取り込んで表紙とかの裏の白紙を削除した時に，これで切り詰めます．

### gray.sh

グレースケールに変換します．

具体的には下記のことをやっています．

1. Red のチャネルを削除 (黄ばみの除去) し，グレースケール変換
1. ヒストグラムの 90% を白とする (紙のノイズを除去)
1. ヒストグラムの 35% を黒とする (ベタを黒塗りする)

## License

This software is released under the MIT License, see LICENSE.txt.

