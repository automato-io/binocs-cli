# binocs-cli

## Release

### 1. Set version in `root.go`

`const BinocsVersion = "v0.4.0"`

### 2. Generate docs and update docs on the website

in `scripts`

`./generate-docs.sh 0.4.0`

### 3. Execute release via GitHub Actions

in `scripts`

`./release.sh 0.4.0`

### 4. Execute post-GitHub Actions script

```shell
$ cd ~/Code/automato/binocs-download/
$ ./update 0.4.0
```

This will 
- fetch assets from the GitHub Release
- zip files for Homebrew
- add files to github.com/automato-io/binocs-downloads
- sync files with download.binocs.sh

### 5. Create a Homebrew PR

- create a branch in `~/Code/automato/homebrew-cask/` manually? - test next time
then:
```shell
brew bump-cask-pr --version 0.4.0 binocs
```

### 6. Build Docker image

- update `BINOCS_VERSION` in `Dockerfile`

```shell
git commit -m 'bump version to v0.4.0'
git tag -a v0.4.0 -m 'release v0.4.0'
git push origin main
git push origin v0.4.0
```

### 7. Release website with updated downloads

- update `routes/web.php` and `config/binocs.php` to include the new default version v0.4.0

```shell
$ cd ~/Code/automato/binocs-website/
$ git push origin master
```

## Testing the continuous integration pipeline

```shell
git push --delete origin v69.1.0 && git tag -d v69.1.0
git tag -a v69.1.0 -m "release v69.1.0" && git push origin v69.1.0
```

## Develop completions

```shell
go run main.go completion bash > $(brew --prefix)/etc/bash_completion.d/binocs
source $(brew --prefix)/etc/bash_completion.d/binocs
```

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
