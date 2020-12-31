#!/usr/bin/env bash

# This script is executed during release process and generates kyma artifacts. Artifacts are stored in $(ARTIFACTS) location
# that is automatically uploaded by Prow to GCS bucket in the following location:
# <plank gcs bucket>/pr-logs/pull/<org_repository>/<pull_request_number>/kyma-artifacts/<build_id>/artifacts
# Information about latest build id is stored in:
# <plank gcs bucket>/pr-logs/pull/<org_repository>/<pull_request_number>/kyma-artifacts/latest-build.txt

set -e

readonly SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck disable=SC1090
source "${SCRIPT_DIR}/library.sh"
# shellcheck source=prow/scripts/lib/gcloud.sh
source "${SCRIPT_DIR}/lib/gcloud.sh"
# shellcheck source=prow/scripts/lib/docker.sh
source "${SCRIPT_DIR}/lib/docker.sh"

# copy_artifacts copies artifacts to the destined bucket path.
# it accepts one argument BUCKET_PATH which should be formatted as:
# gs://bucket-name/bucket-folder
function copy_artifacts {
  readonly BUCKET_PATH=$1
  log::info "Copying artifacts to $BUCKET_PATH..."

  gsutil cp  "installation/scripts/is-installed.sh" "$BUCKET_PATH/is-installed.sh"
  gsutil cp "${ARTIFACTS}/kyma-installer-cluster.yaml" "$BUCKET_PATH/kyma-installer-cluster.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-cluster-runtime.yaml" "$BUCKET_PATH/kyma-installer-cluster-runtime.yaml"

  gsutil cp "${ARTIFACTS}/kyma-config-local.yaml" "$BUCKET_PATH/kyma-config-local.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-local.yaml" "$BUCKET_PATH/kyma-installer-local.yaml"

  gsutil cp "${ARTIFACTS}/kyma-installer.yaml" "$BUCKET_PATH/kyma-installer.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-cr-cluster.yaml" "$BUCKET_PATH/kyma-installer-cr-cluster.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-cr-local.yaml" "$BUCKET_PATH/kyma-installer-cr-local.yaml"
  gsutil cp "${ARTIFACTS}/kyma-installer-cr-cluster-runtime.yaml" "$BUCKET_PATH/kyma-installer-cr-cluster-runtime.yaml"
}

gcloud::authenticate "${GOOGLE_APPLICATION_CREDENTIALS}"
docker::start

log::info "Building kyma-installer"
# Building kyma-installer image using build.sh script.
# Handles basically everything related to building process including determining version, exporting DOCKER_TAG etc.
"${SCRIPT_DIR}"/build-generic.sh "tools/kyma-installer"

log::info "Create Kyma artifacts"
env KYMA_INSTALLER_VERSION="${DOCKER_TAG}" ARTIFACTS_DIR="${ARTIFACTS}" "${KYMA_PATH}/installation/scripts/release-generate-kyma-installer-artifacts.sh"

log::info "Content of the local artifacts directory"
ls -la "${ARTIFACTS}"

if [ -n "$PULL_NUMBER" ]; then
  copy_artifacts "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/$DOCKER_TAG"
elif [[ "$PULL_BASE_REF" =~ ^release-.* ]]; then
  copy_artifacts "${KYMA_ARTIFACTS_BUCKET}/${DOCKER_TAG}"
else
  copy_artifacts "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/$DOCKER_TAG"
  copy_artifacts "${KYMA_DEVELOPMENT_ARTIFACTS_BUCKET}/master"
fi

"${SCRIPT_DIR}"/changelog-generator.sh
