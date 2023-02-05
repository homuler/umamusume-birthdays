# メンテナンス用データファイル

## characters.yaml

全ウマ娘の誕生日データ。

新規ウマ娘が登場したらこちらをメンテナンスする。

```yaml
- name: 名前
  birthday: MM/DD
  url: https://umamusume.jp/character/detail/?name=
  playable: true
  costumes:
    school: 制服のURL
    racing: 勝負服のURL
    original: 原案のURL
    sf: STARTING FUTUREのURL
  variations:
    - url: 実装されているキャラクターのURL
```

