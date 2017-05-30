#!/bin/bash

# Copyright 2017 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Exit on error.
set -e

readonly ARGV0=${0}

# Go to this script's directory.
cd "$(/usr/bin/dirname $(/bin/readlink -e ${ARGV0}))"

# If PROTOC isn't set:
if [[ -z "${PROTOC}" ]]; then
  readonly PROTOC='/usr/local/bin/protoc'
else
  readonly PROTOC=${PROTOC}
fi

# If PROTOC is not an executable:
if ! [[ -x "${PROTOC}" ]]; then
  echo >&2 "
protoc is not available. Either install protoc, or point this script to protoc using:
  PROTOC=/path/to/protoc ${ARGV0}
"
  exit 1
fi

readonly PROTOS='
analysis
anomaly
api_utils
export
flows
jobs
knowledge_base
output_plugin
semantic
sysinfo
api/client
api/cron
api/flow
api/hunt
api/output_plugin
api/reflection
api/stats
api/user
api/vfs
'

function generate_protoc_m_opts {
  for p in ${PROTOS}; do
    echo -n "\
Mgrr/proto/${p}.proto\
=\
github.com/google/grr_go_api_client/grrproto/grr/proto/${p}_proto,"
  done
}

# These map proto import paths to go import paths.
readonly PROTOC_M_OPTS=$(generate_protoc_m_opts)

readonly GOOGLE_PATH='submodules/github.com/google/'
readonly GRR_PATH="${GOOGLE_PATH}/grr/"

readonly OUTPUT_PATH='grrproto/'

/bin/echo
/bin/echo '- Transpiling required protos to Go.'
for p in ${PROTOS}; do
  # Careful, p might contain slashes.

  source_file="${GRR_PATH}/grr/proto/${p}.proto"
  output_filename="$(/usr/bin/basename "${p}")"
  output_filepath="${OUTPUT_PATH}/grr/proto/${p}.pb.go"
  package_dir="${OUTPUT_PATH}/grr/proto/${p}_proto/"
  package_name="${output_filename}_proto"
  package_path="${package_dir}/${package_name}"

  /bin/echo
  /bin/echo "-- Creating package directory ${package_dir}"
  /bin/mkdir -p "${package_dir}"

  /bin/echo "-- Transpiling ${source_file} => ${output_filepath}"
  # Pass the updated PATH to the PROTOC executable.
  PATH="${GOPATH}/bin/:${PATH}" "${PROTOC}" \
    --go_out="${PROTOC_M_OPTS}:${OUTPUT_PATH}" \
    --proto_path="${GRR_PATH}" \
    "${source_file}"

  symlink_name="${package_dir}/${package_name}.go"
  symlink_target="../${output_filename}.pb.go"

  /bin/echo "-- Symlinking ${symlink_name} => ${symlink_target}"
  /bin/ln -fsT "${symlink_target}" "${symlink_name}"
done
