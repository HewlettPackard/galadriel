# #!/bin/bash


# ##Step Variables

COMMIT=6390cfef8b7f69c019481615b28c33209d8129d8416c9fa3200c29f7d86f8ba8
SCORE=a6a3923b64d3fd43b5b93add84ecdb65f95e05f2ba207942cfb066dc03c0f293
SCAN=5d69f1b35a26f40cbc23787e37fd356d4ad5cbed48bcd1a9ad3aed06fc8de52a
CONTAINER_BUILD=96425b681d29fd050773f62452631049640efe972494950a01258a9690360e39


# # ##Get Certs from Fulcio
# ./get_certs.sh
#

policy-tool create -x $COMMIT -r root.pem -i intermediate.pem -x $SCORE -r root.pem -i intermediate.pem -x $SCAN -r root.pem -i intermediate.pem -x $CONTAINER_BUILD -r root.pem -i intermediate.pem -t https://freetsa.org/files/cacert.pem > policy.json

##create RSA public private key pair for policy signing
openssl genrsa -out policy.key 2048
openssl rsa -in policy.key -pubout -out policy.pub

##sign policy
witness sign -f policy.json -k policy.key -o policy-signed.json

# ##download binaries zip from https://github.com/testifysec/galadriel/suites/12250032590/artifacts/648297307

# wget https://github.com/testifysec/galadriel/suites/12250032590/artifacts/648297307 -O galadriel.zip

# ##unzip binaries to dist folder
# unzip galadriel.zip -d dist

# ## verify each binary, recurse through dist folder and verify each binary without an extension




# for file in dist/*; do



#   echo "Verifying $file"
#   witness verify -f $file -p policy-signed.json -k policy.pub
# done
witness verify -f harvester_cli -p policy-signed.json -k policy.pub --enable-archivista