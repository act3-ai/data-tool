#!/bin/sh

# Example: ./homebrew.sh v1.2.3
version=$1
shift

# TODO: potentially make these arguments to support other repositories.
# If we really want to generalize this we could add an "archive" option for tar.gz
# or the like release assets.
repo_name="data-tool"
cli_name="ace-dt"
description="A CLI tool for packaging, uploading, and downloading data from OCI registries."

resolveURL() {
    os=$1
    arch=$2

    echo -n "https://github.com/act3-ai/${repo_name}/releases/download/${version}/${cli_name}-${version}-${os}-${arch}"
}

# resolveSHA returns the sha256 hash of an asset.
# It first attempts to find local copies of assets, falling back
# to downloading from the remote release assets if not available.
resolveSHA(){
    os=$1
    arch=$2

    bin_file="./release/assets/${cli_name}-${version}-${os}-${arch}"
    if [ -f $bin_file ]; then
        sha=$(cat $bin_file | shasum -a 256 | cut -f1 -d\ "")
    else
        echo "local release asset $bin_file not found, fetching from remote" >&2
        sha=$(curl -sLS "$(resolveURL $os $arch)" | shasum -a 256 | cut -f1 -d\ "")
    fi

    echo -n "${sha}"
}

cat > ${cli_name}.rb <<FORMULA
class Hops < Formula
  desc "${description}"
  homepage "https://github.com/act3-ai/${repo_name}"
  version "${version}"
  license "MIT"

  on_macos do
    on_intel do
      url "$(resolveURL darwin amd64)"
      sha256 "$(resolveSHA darwin amd64)"

      def install
        bin.install "bin/${cli_name}"
        generate_completions_from_executable(bin/${cli_name}, "completion")

        # Generate manpages
        mkdir "man" do
          system bin/${cli_name}, "gendocs", "man", "."
          man1.install Dir["*.1"]
        end
      end
    end
    on_arm do
      url "$(resolveURL darwin arm64)"
      sha256 "$(resolveSHA darwin arm64)"

      def install
        bin.install "bin/${cli_name}"
        generate_completions_from_executable(bin/${cli_name}, "completion")

        # Generate manpages
        mkdir "man" do
          system bin/${cli_name}, "gendocs", "man", "."
          man1.install Dir["*.1"]
        end
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "$(resolveURL linux amd64)"
        sha256 "$(resolveSHA linux amd64)"

        def install
          bin.install "bin/${cli_name}"
          generate_completions_from_executable(bin/${cli_name}, "completion")

          # Generate manpages
          mkdir "man" do
          system bin/${cli_name}, "gendocs", "man", "."
          man1.install Dir["*.1"]
          end
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "$(resolveURL linux arm64)"
      sha256 "$(resolveSHA linux arm64)"

        def install
          bin.install "bin/${cli_name}"
          generate_completions_from_executable(bin/${cli_name}, "completion")

          # Generate manpages
          mkdir "man" do
          system bin/${cli_name}, "gendocs", "man", "."
          man1.install Dir["*.1"]
          end
        end
      end
    end
  end

  test do
    system "#{bin}/${cli_name} --version"
  end
end
FORMULA

echo "Successfully updated! Check the formula before pushing."