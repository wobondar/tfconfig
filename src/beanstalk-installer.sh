#!/usr/bin/env bash

mkdir -p /tmp/tfconfig
cd /tmp/tfconfig
curl -L -o tfconfig.zip https://github.com/wobondar/tfconfig/releases/download/v${TF_CONFIG_VERSION}/tfconfig_v${TF_CONFIG_VERSION}_linux_amd64.zip
echo "${TF_CONFIG_SHA256} tfconfig.zip" | sha256sum -c --quiet
unzip tfconfig.zip
mv tfconfig /usr/bin
chmod +x /usr/bin/tfconfig
rm -rf /tmp/tfconfig