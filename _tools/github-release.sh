#!/bin/sh

ACCEPT_HEADER="Accept: application/vnd.github.jean-grey-preview+json"
TOKEN_HEADER="Authorization: token ${GITHUB_TOKEN}"
RELEASE_ENDPOINT="https://api.github.com/repos/ysugimoto/ginger/releases"

RELEASE_ID=$(curl -XPOST -H "${ACCEPT_HEADER}" -H "${TOKEN_HEADER}" -d "{\"tag_name\": \"${CIRCLE_TAG}\", \"name\": \"${CIRCLE_TAG}\"}" | jq .id)
RELEASE_URL="https://uploads.github.com/repos/ysugimoto/release-example/releases/${RELEASE_ID}/assets"

for FILE in `ls ./dist`; do
  curl -v -XPOST -H "${ACCEPT_HEADER}" -H "${TOKEN_HEADER}" -H "Content-Type: application/octet-stream" -d "@dist/${FILE}" "${RELEASE_URL}?${FILE}"
done
