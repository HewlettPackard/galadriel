set -euo pipefail

hashes=$(echo $ARTIFACTS | jq --raw-output '.[] | {name, "digest": (.extra.Digest // .extra.Checksum)} | select(.digest) | {digest} + {name} | join("  ") | sub("^sha256:";"")' | base64 -w0)
if test "$hashes" = ""; then # goreleaser < v1.13.0
    checksum_file=$(echo "$ARTIFACTS" | jq -r '.[] | select (.type=="Checksum") | .path')
    hashes=$(cat $checksum_file | base64 -w0)
fi

echo "hashes=$hashes" >> $GITHUB_OUTPUT
