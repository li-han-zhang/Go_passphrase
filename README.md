# Introduction
Some Crypto Wallet can only create 12-word passphrase, but now the tool can create a 24-word passphrase and you are able to edit the binary-based file that can devires passphrase.
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
## 3.Initializes a new Go module.
```
go mod init passphrase_bitcoin
```
## 4.Replaces the github.com/skip2/go-qrcode dependency.
```
go mod edit -replace github.com/skip2/go-qrcode=./go-qrcode
```
## 5.Cleans up the go.mod and go.sum files.
```
go mod tidy
```
## 6.Builds the Go project into an executable file.
```
go build -ldflags "-s -w" -o passphrase_bitcoin main.go
```
# Usage
```
./passphrase_bitcoin -b -p
```
