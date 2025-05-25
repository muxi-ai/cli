# MUXI CLI Distribution Guide

This guide explains how to compile the MUXI CLI for multiple platforms and how to distribute it via Homebrew, APT, and other package managers.

---

## 🛠 Building MUXI for Different Platforms

### 🔧 Standard Build (Local)

```bash
go build -o muxi
```

This builds the CLI for your current OS and architecture.

---

### 🌍 Cross-Compiling

You can cross-compile MUXI using Go’s built-in environment variables.

#### macOS (arm64 and amd64)

```bash
GOOS=darwin GOARCH=arm64 go build -o muxi-darwin-arm64
GOOS=darwin GOARCH=amd64 go build -o muxi-darwin-amd64
```

#### Linux

```bash
GOOS=linux GOARCH=amd64 go build -o muxi-linux
GOOS=linux GOARCH=arm GOARM=7 go build -o muxi-linux-armv7
```

#### Windows

```bash
GOOS=windows GOARCH=amd64 go build -o muxi.exe
```

---


# MUXI CLI Distribution Guide

This guide explains how to compile the MUXI CLI for multiple platforms and how to distribute it via GitHub Releases. All package managers and install methods are configured to fetch the binary directly from GitHub — no need to publish to official OS repositories.

---

## 🛠 Building MUXI for Different Platforms

### 🔧 Standard Build (Local)

```bash
go build -o muxi
```

This builds the CLI for your current OS and architecture.

---

### 🌍 Cross-Compiling

Use Go’s environment variables to compile for any platform.

#### macOS (arm64 and amd64)

```bash
GOOS=darwin GOARCH=arm64 go build -o muxi-darwin-arm64
GOOS=darwin GOARCH=amd64 go build -o muxi-darwin-amd64
```

#### Linux

```bash
GOOS=linux GOARCH=amd64 go build -o muxi-linux
GOOS=linux GOARCH=arm GOARM=7 go build -o muxi-linux-armv7
```

#### Windows

```bash
GOOS=windows GOARCH=amd64 go build -o muxi.exe
```

Upload all resulting binaries to a GitHub release.

---

## 📦 macOS (Homebrew via Tap)

1. Create a tap repo: `github.com/yourname/homebrew-muxi`
2. Add a `muxi.rb` formula pointing to your GitHub release:

```ruby
class Muxi < Formula
  desc "MUXI CLI - AI agent registry and server interface"
  homepage "https://muxi.dev"
  url "https://github.com/yourname/muxi/releases/download/v1.0.0/muxi-darwin-amd64.tar.gz"
  sha256 "<computed-sha256>"
  version "1.0.0"

  def install
    bin.install "muxi"
  end
end
```

3. Users install with:

```bash
brew tap yourname/muxi
brew install muxi
```

---

## 🧰 Debian/Ubuntu (.deb from GitHub)

1. Package with fpm:

```bash
fpm -s dir -t deb -n muxi -v 1.0.0 muxi=/usr/local/bin/muxi
```

2. Upload `.deb` to GitHub release.
3. Users install with:

```bash
wget https://github.com/yourname/muxi/releases/download/v1.0.0/muxi_1.0.0_amd64.deb
sudo dpkg -i muxi_1.0.0_amd64.deb
```

---

## 📀 RedHat/Fedora (.rpm from GitHub)

1. Create `.rpm`:

```bash
fpm -s dir -t rpm -n muxi -v 1.0.0 muxi=/usr/local/bin/muxi
```

2. Upload to GitHub.
3. Users install with:

```bash
wget https://github.com/yourname/muxi/releases/download/v1.0.0/muxi-1.0.0.x86_64.rpm
sudo rpm -i muxi-1.0.0.x86_64.rpm
```

---

## 🏹 Arch Linux (AUR pointing to GitHub)

1. Create a `PKGBUILD`:

```bash
pkgname=muxi
pkgver=1.0.0
pkgrel=1
arch=('x86_64')
source=("https://github.com/yourname/muxi/releases/download/v1.0.0/muxi-linux")
sha256sums=("<sha256>")

package() {
  install -Dm755 muxi-linux "$pkgdir/usr/bin/muxi"
}
```

2. Publish to AUR (requires AUR account).

---

## 🧊 Alpine Linux (Direct binary)

1. Build statically:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o muxi-alpine
```

2. Upload to GitHub.
3. Users install:

```bash
wget https://github.com/yourname/muxi/releases/download/v1.0.0/muxi-alpine
chmod +x muxi-alpine
sudo mv muxi-alpine /usr/local/bin/muxi
```

---

## ❄️ NixOS (with fetchurl from GitHub)

1. Create `default.nix`:

```nix
{ stdenv, fetchurl }:
stdenv.mkDerivation {
  name = "muxi-1.0.0";
  src = fetchurl {
    url = "https://github.com/yourname/muxi/releases/download/v1.0.0/muxi-linux";
    sha256 = "<sha256>";
  };
  installPhase = ''
    mkdir -p $out/bin
    cp muxi-linux $out/bin/muxi
    chmod +x $out/bin/muxi
  '';
}
```

2. Build:

```bash
nix-build
```

---

## 🔚 Summary

All distributions point to **GitHub Releases** — you do **not** need to push packages to official OS repositories.

| Platform      | Install Method                                |
| ------------- | --------------------------------------------- |
| macOS         | Homebrew tap with formula pointing to GitHub  |
| Debian/Ubuntu | `.deb` via GitHub + `dpkg -i`                 |
| RedHat/Fedora | `.rpm` via GitHub + `rpm -i`                  |
| Arch          | AUR entry pointing to GitHub binary           |
| Alpine        | Direct download of statically compiled binary |
| NixOS         | `.nix` file using `fetchurl` from GitHub      |

