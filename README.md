# binocs-cli

## Release

### 1. Set version in `root.go`

`const BinocsVersion = "v0.4.0"`

### 2. Generate docs and update docs on the website

```shell
$ go run main.go docgen
$ mkdir ~/Code/automato/binocs-website/resources/docs/v0.4.0/
$ cp -a docs/* ~/Code/automato/binocs-website/resources/docs/v0.4.0/
```

Update `routes/web.php` and `config/binocs.php` in web project to include the new version and make it default.

### 3. Execute release via GitHub Actions

```shell
$ git commit -m 'bump version to 0.4.0'
$ git tag -a v0.4.0 -m 'release v0.4.0'
$ git push origin master
$ git push origin v0.4.0
```

Download binaries to `~/Code/automato/binocs-download/public/`

```shell
$ cd ~/Code/automato/binocs-download/
$ ./update 0.4.0
$ ./sync
```

Release `binocs-website`

## Test cases for valid UpCode regexp pattern

```
404
2xx
30x
200-301
200-301,404
404,200-301
200-202,300-302
200-202,300-302,404
200-202,404,300-302
200-202,300-302,403-404
---
200-301-404
200-2xx,404,300-302
099
4044
5xxx
20
2x
2
x
---
200-101
```
