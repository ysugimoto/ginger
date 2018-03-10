#!/bin/sh

## Configuration for your project

### If you run this script as inline code like Jenkins, define as following:
#GITHUB_TOKEN="your github access token"

### We strongly recommend that specify access token at external environement variable.

### Your project repository
REPOSITORY=":owner/:repo"

### Directory contains build artifacts if you have.
### If you don't have any artifacts, please keep it empty.
ASSETS_DIR=""

### Determine release tag name
TAG=

### If you are using via some CI servcice, you can use following server specific variable:
#TAG="${CIRCLE_TAG}" Circle CI
#TAG="${TRAVIS_TAG}" Travis CI


####### You don't need to modify following area #######

ACCEPT_HEADER="Accept: application/vnd.github.jean-grey-preview+json"
TOKEN_HEADER="Authorization: token ${GITHUB_TOKEN}"
ENDPOINT="https://api.github.com/repos/${REPOSITORY}/releases"

echo "Creatting new release as version ${TAG}..."
RELEASE_ID=$(curl -H "${ACCEPT_HEADER}" -H "${TOKEN_HEADER}" -d "{\"tag_name\": \"${TAG}\", \"name\": \"${TAG}\"}" "${ENDPOINT}" | jq .id)
RELEASE_URL="https://uploads.github.com/repos/${REPOSITORY}/releases/${RELEASE_ID}/assets"

echo "Github release created as ID: ${RELEASE_ID}. https://github.com/${REPOSITORY}/release"

if [ "${ASSETS_DIR}" = "" ]; then
  echo "No upload assets, finished."
  exit
fi

for FILE in ${ASSETS_DIR}; do
  MIME=$(file -b --mime-type "${ASSETS_DIR}/${FILE}")
  echo "Uploading assets ${ASSETS_DIR}/${FILE} as ${MIME}..."
  curl -v \
    -H "${ACCEPT_HEADER}" \
    -H "${TOKEN_HEADER}" \
    -H "Content-Type: ${MIME}" \
    --data-binary "@${ASSETS_DIR}/${FILE}" \
    "${RELEASE_URL}?name=${FILE}"
done

echo "Finished."
