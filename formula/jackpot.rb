class Jackpot < Formula
  desc ""
  homepage ""
  url "https://github.com/prabhjot98/jackpot"
  version "0.1"
  sha256 "b3d25aace78677caa9da91db69076b9ba01de39a5840024f1910c7c9ca490ce0"
  license ""

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w")
  end

  test do
    # `test do` will create, run in and delete a temporary directory.
    #
    # This test will fail and we won't accept that! For Homebrew/homebrew-core
    # this will need to be a test that verifies the functionality of the
    # software. Run the test with `brew test jackpot`. Options passed
    # to `brew install` such as `--HEAD` also need to be provided to `brew test`.
    #
    # The installed folder is not in the path, so use the entire path to any
    # executables being tested: `system bin/"program", "do", "something"`.
    system "false"
  end
end
