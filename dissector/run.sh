#/bin/sh
# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

set -e

GONT_PATH=${GONT_PATH:-$(pwd)}
GONT_REMOTE_PATH=${GONT_REMOTE_PATH:-${GONT_PATH}}

WIRESHARK_OPTS="-Xlua_script:${GONT_PATH}/dissector/dissector.lua -i TCP@[::1]:42125"

echo "#"
echo "# Start capturing in Wireshark once the test is running and waiting"
echo "# for an incoming connection."
echo "#"
echo "# Tip: You can start the capture in Wireshark via Strg+E"
echo "#"
echo "########################################################################"

if [[ -n ${REMOTE} ]]; then
    RUN="ssh -L 42125:[::1]:42125 ${REMOTE} cd ${GONT_REMOTE_PATH};"
fi

wireshark ${WIRESHARK_OPTS} &

${RUN} sudo -E go test -count=1 -v -run TestTracer ./pkg --trace-socket="tcp:[::1]:42125"
