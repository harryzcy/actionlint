# typed: false
# frozen_string_literal: true

# This file was generated by GoReleaser. DO NOT EDIT.
class Actionlint < Formula
  desc "Static checker for GitHub Actions workflow files"
  homepage "https://github.com/rhysd/actionlint#readme"
  version "1.6.20"
  license "MIT"

  on_macos do
    if Hardware::CPU.intel?
      url "https://github.com/rhysd/actionlint/releases/download/v1.6.20/actionlint_1.6.20_darwin_amd64.tar.gz"
      sha256 "76bfe7a946d055b28007d72dd1cb3541013d0bb1d5a85051eedced36da161531"

      def install
        bin.install "actionlint"
        man1.install "man/actionlint.1"
      end
    end
    if Hardware::CPU.arm?
      url "https://github.com/rhysd/actionlint/releases/download/v1.6.20/actionlint_1.6.20_darwin_arm64.tar.gz"
      sha256 "48af85cd69bb379d82c82e972ea452dde8b40fc57675d37a176d139f146ae78c"

      def install
        bin.install "actionlint"
        man1.install "man/actionlint.1"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && !Hardware::CPU.is_64_bit?
      url "https://github.com/rhysd/actionlint/releases/download/v1.6.20/actionlint_1.6.20_linux_armv6.tar.gz"
      sha256 "eef680479b4fc8219a362ceaa31e0123ebde937371c591736553b084c74360a6"

      def install
        bin.install "actionlint"
        man1.install "man/actionlint.1"
      end
    end
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/rhysd/actionlint/releases/download/v1.6.20/actionlint_1.6.20_linux_arm64.tar.gz"
      sha256 "a73056e51ec83238790bb6fafb3e20a14d14856469f79d4c4e261d6215c1cd8f"

      def install
        bin.install "actionlint"
        man1.install "man/actionlint.1"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/rhysd/actionlint/releases/download/v1.6.20/actionlint_1.6.20_linux_amd64.tar.gz"
      sha256 "bc311cb7bf006072009ab55ef81f7b6545bf8fc35d63ae60e002c1e37e4afe04"

      def install
        bin.install "actionlint"
        man1.install "man/actionlint.1"
      end
    end
  end

  test do
    system "#{bin}/actionlint -version"
  end
end
