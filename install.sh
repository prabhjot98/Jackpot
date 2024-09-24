RELEASE_VERSION=$(curl -s https://api.github.com/repos/prabhjot98/jackpot/releases/latest | jq .name -r)
DOWNLOAD_URL="https://github.com/prabhjot98/jackpot/releases/download/${RELEASE_VERSION}/jackpot_Darwin_arm64.tar.gz"
echo "Downloading binary for Darwin arm64..."
curl -s -L "$DOWNLOAD_URL" -o jackpot.tar.gz
tar -xf jackpot.tar.gz
chmod +x jackpot
sudo cp jackpot /usr/local/bin
rm jackpot.tar.gz jackpot
echo "Done! Have fun gambling :)"
