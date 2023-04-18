#!/bin/bash


# ##Step Variables

COMMIT=7b2dbf215fdf0d36e99eeb0e60f13f1bf2f2484622feb912df8a9e410ad30ce2
SCORE=77429816a31defbb3b47d6ece96707b7c81f182b87c69b75f15224430a82bf15
SCAN=cf912f0774d6ec3792ea8d0bf495a847602cec2876d8edee1112c1dfa36117a1
CONTAINER_BUILD=52097132a1c74e32ceac7c2f0b889d7489a121272d4ffb45e3bf8a5943b74c0b
BINARY_BUILD=13a5275ba0dc2eef450ed73cea4e0f0478ff356febbcfbaeb34106f651d9e2a2


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
witness verify -s 74858372912956e65554bde585846522485d7de7 -p .witness/policy-signed.json -k .witness/policy.pub --enable-archivista

# Verify the container build by image ID
echo "Verifying by the container imageID"
witness verify -s 18fe3c392c293a63200fc700a3e7a62d07ae180aac2040a1132cb1827cf8f720 -p .witness/policy-signed.json -k .witness/policy.pub --enable-archivista

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