#!/bin/bash

CONFIG_FILES_LIST=$1
NEXT_ENV=$2
CONFIGS_UPDATED=false

for config in ${CONFIG_FILES_LIST}
do
  config_file_name=$(echo "${config}" | awk -F/ '{print $(NF)}')
  old_version=$(cat < "${NEXT_ENV}/${config_file_name}" | grep artifact_version | awk '{print $2}')
  new_version=$(cat < "${config}" | grep artifact_version | awk '{print $2}')
  if test -f "${NEXT_ENV}/${config_file_name}" && [[ ${new_version} != "${old_version}" ]]; then
    echo "[${NEXT_ENV}/${config_file_name}]: Updating version to ${new_version}..."
    sed -i "s/artifact_version: .*/artifact_version: ${new_version}/g" "${NEXT_ENV}"/"${config_file_name}"
    CONFIGS_UPDATED=true
  else
    echo "[${NEXT_ENV}/${config_file_name}]: Skipping update..."
    echo "[${NEXT_ENV}/${config_file_name}]: Versions are the same in both environments or there is no config file in ${NEXT_ENV} env"
  fi
done
echo "::set-output name=configs_updated::${CONFIGS_UPDATED}"
