#!/bin/bash


# ##Step Variables

COMMIT=62a05fc10707785b3897204afe4945e57778cf2bbf9ce70e4399570b14e9e0c7
SCORE=53ea1058003205b374f86a60785b34d571b76bf64883f8d4d966614a335b17fb
SCAN=353ed220f2edae7f0194c39818a795f8bb1555795fccbc691b79d7dd384656ff
CONTAINER_BUILD=cbff518f9c9fb1d215fa12122f1be87fbd418bb9c6cb67cb450c912b17dfa333
BINARY_BUILD=4f2816f7dc0d8e21b4025cf0ff64b64e075c1edce0d85be3dc5a7c64ba38db35


mkdir -p .witness

SIGSTORE_ROOT=.witness/root.pem
SIGSTORE_INTERMEDIATE=.witness/intermediate.pem

# Get Certs from Fulcio
./get_certs.sh  && mv root.pem ${SIGSTORE_ROOT} && mv intermediate.pem ${SIGSTORE_INTERMEDIATE}

# Create policy
policy-tool create -x=$COMMIT -r $SIGSTORE_ROOT -i $SIGSTORE_INTERMEDIATE --constraint-emails colek42@gmail.com -x $SCORE -r $SIGSTORE_ROOT -i $SIGSTORE_INTERMEDIATE -x $SCAN -r $SIGSTORE_ROOT -i $SIGSTORE_INTERMEDIATE -y .witness/sticky.yaml -x $CONTAINER_BUILD -r $SIGSTORE_ROOT -i $SIGSTORE_INTERMEDIATE -y .witness/sticky.yaml -t https://freetsa.org/files/cacert.pem > .witness/policy.json

# Create RSA public-private key pair for policy signing
openssl genrsa -out .witness/policy.key 2048
openssl rsa -in .witness/policy.key -pubout -out .witness/policy.pub

# Sign policy
witness sign -f .witness/policy.json -k .witness/policy.key -o .witness/policy-signed.json


# Verify commit
echo "Verifying by the commit"
witness verify -s 1e64684f8230fe662c384d0b1108ed6ec5ac36ee -p .witness/policy-signed.json -k .witness/policy.pub --enable-archivista

# Verify the container build by image ID
echo "Verifying by the container imageID"
witness verify -s 9efbee1c55fd477d97e8be2f625cdbf66ba5618c6797e7effdcfb56e56ef2adc -p .witness/policy-signed.json -k .witness/policy.pub --enable-archivista

# # Create policy for binary build
policy-tool create -x $COMMIT -r $SIGSTORE_ROOT -i $SIGSTORE_INTERMEDIATE --constraint-emails colek42@gmail.com -x $SCORE -r $SIGSTORE_ROOT -i $SIGSTORE_INTERMEDIATE -x $SCAN -r $SIGSTORE_ROOT -i $SIGSTORE_INTERMEDIATE -x $BINARY_BUILD -r $SIGSTORE_ROOT -i $SIGSTORE_INTERMEDIATE -t https://freetsa.org/files/cacert.pem > .witness/policy-bin.json

# # Sign policy for binary build
witness sign -f .witness/policy-bin.json -k .witness/policy.key -o .witness/policy-bin-signed.json

if [[ ! -d "dist" ]]; then
  echo "dist folder does not exist"
  echo "Please download the binaries from release step and unzip them to dist folder"
  echo "https://github.com/testifysec/galadriel/actions"
  exit 1
fi 

# # Recurse through dist folder and verify each binary without an extension
find ./dist -type f | while read FILE
do
  # Exclude config.yaml since it is common
  if [[ $FILE == *"config.yaml"* ]]; then
    continue
  fi

  # Run witness verify on the file
  echo "Verifying $FILE"
  witness verify -f $FILE -p .witness/policy-bin-signed.json -k .witness/policy.pub --enable-archivista
done