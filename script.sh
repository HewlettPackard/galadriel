#!/bin/bash


# ##Step Variables

COMMIT=105125cad7632159e8ce181a3abe39dc68b09020b827ce2a80bed91528329460
SCORE=c4942b9f6fe2b1cbbbefb066c1b2ebe65d2afaed2ffd162ddb98eabb50eb5fec
SCAN=cf912f0774d6ec3792ea8d0bf495a847602cec2876d8edee1112c1dfa36117a1
CONTAINER_BUILD=fd31d8885ec158d551312d9d1b41b2cb67dd28696f7682d8d8a8012a019ea94f
BINARY_BUILD=5b80eaf5161f1a896bb585d922e3d0c91ff1785cb09be9d77e9d6c8b7abbd843


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