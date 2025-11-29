# Introduction
    Some crypto wallets can only create a 12‑word passphrase. However, this tool can generate a 24‑word passphrase, and it also allows you to edit the binary file that the passphrase is derived from.
This tool generates passphrases using BIP‑39, which contains 2048 words. Each word represents 11 bits of binary data, and you can modify the binary file as randomly as you like.
# Installation
## 1.Clone this repository
```
git clone https://github.com/li-han-zhang/Go_passphrase.git
```
## 2.Install Go
On ubuntu
```
sudo apt install go
```
On arch linux
```
sudo pacman -S go
```
## 3.Initializes a new Go module
```
go mod init passphrase_bitcoin
```
## 4.Replaces the github.com/skip2/go-qrcode dependency
```
go mod edit -replace github.com/skip2/go-qrcode=./go-qrcode
```
## 5.Cleans up the go.mod and go.sum files
```
go mod tidy
```
## 6.Builds the Go project into an executable file
```
go build -ldflags "-s -w" -o passphrase_bitcoin main.go
```
# Usage
```
./passphrase_bitcoin -b -p
```
### You can just download the executable file, passphrase_bitcoin, and use it.
