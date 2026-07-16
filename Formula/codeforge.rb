# Homebrew formula (tap: NanoMindExplorer/codeforge or copy into a personal tap)
#   brew install --build-from-source Formula/codeforge.rb
#   brew install NanoMindExplorer/tap/codeforge   # when published
class Codeforge < Formula
  desc "Terminal AI coding companion (multi-provider agent + GitHub)"
  homepage "https://github.com/NanoMindExplorer/codeforge"
  version "1.9.3"
  license "Apache-2.0"

  on_macos do
    on_arm do
      url "https://github.com/NanoMindExplorer/codeforge/releases/download/v#{version}/codeforge_#{version}_darwin_arm64.tar.gz"
      sha256 "837dc9d7d8a506faaf0a446956d6afbe76b58f91c7d2c84ce567d197f3e09c34"
    end
    on_intel do
      url "https://github.com/NanoMindExplorer/codeforge/releases/download/v#{version}/codeforge_#{version}_darwin_amd64.tar.gz"
      sha256 "843e7d62047078911423315999b6f4ddde1184b2b12e35797410422c7f4e469b"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/NanoMindExplorer/codeforge/releases/download/v#{version}/codeforge_#{version}_linux_arm64.tar.gz"
      sha256 "d6970301e38fa23f83e1487ee786298c3c6274a04dbf23788cdd3484dbf469e9"
    end
    on_intel do
      url "https://github.com/NanoMindExplorer/codeforge/releases/download/v#{version}/codeforge_#{version}_linux_amd64.tar.gz"
      sha256 "24d6abde95ac747fc66d22c669386ea82dd33cbcca3be89226487a5a6d6c142d"
    end
  end

  def install
    bin.install "codeforge"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/codeforge version")
  end
end
