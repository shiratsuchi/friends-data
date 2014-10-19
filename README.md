## 使い方

下記の「準備」を終えてから。

`go run main.go shirayuca`

として、ユーザ名を指定して実行する。

## 結果

同じディレクトリに、ユーザ名から始まる２つの CSV ファイルが作られる。(例: `shirayuca.csv`, `shirayuca_following_list.csv`)

`shirayuca.csv` には、フォローしているユーザの情報が記録される。
`shirayuca_following_list.csv` には、shirayuca がフォローしているユーザの中での、フォロー関係が記録される。

## 注意

API 制限に引っかかる場合が多く、15分くらい応答がないことがある。
待っていればちゃんと取得できる。

## 準備

同じディレクトリに `APIKEY.txt` という4行のテキストファイルを置く。
４行は、上から、

- Consumer key
- Consumer secret
- Access token
- Access token secret

(下記はファイルの例。使えません)

```
AAmpFxPtpaL73K5IyrJiLDgRx
amn1YlGhbv5jT1vZ20GhjrPZFearnHpGo9NV97IDU7HVs6PrwA
0814821-XbKjRwh0gqIwV9NlB2L1bMrX4bhs3b7xFP3K6FXA5u
AJENHnmg7cG3ZdP9rLKYxZxMtVwAolLa3qYAsWKOEsa2e
```

## 備考

コンパイル方法

```
$ GOOS=windows GOARCH=386 go build -o twitter-friends-data.exe main.go
```
