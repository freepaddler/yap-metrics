#!/bin/bash

binPath="test/bin/metricstest-darwin-arm64"
#export LOG_LEVEL=debug

rnd() {
  echo $((RANDOM % 30000 + 20000))
}

(cd cmd/server || exit 1; go build -buildvcs=false -o server)
(cd cmd/agent || exit 1; go build -buildvcs=false  -o agent)

# inc1
"$binPath" -test.v -test.run=^TestIteration1$ \
-binary-path=cmd/server/server
[ $? -eq 0 ] || exit 1
sleep 2

# inc2
"$binPath" -test.v -test.run=^TestIteration2[AB]*$ \
-source-path=. \
-agent-binary-path=cmd/agent/agent
[ $? -eq 0 ] || exit 1
sleep 2

# inc3
"$binPath" -test.v -test.run=^TestIteration3[AB]*$ \
-source-path=. \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server
# shellcheck disable=SC2181
[ $? -eq 0 ] || exit 1
sleep 2

# inc4
SERVER_PORT=$(rnd)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE="/tmp/metrictest$(rnd)"
"$binPath" -test.v -test.run=^TestIteration4$ \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server \
-server-port="$SERVER_PORT" \
-source-path=.
[ $? -eq 0 ] || exit 1
sleep 2

# inc5
SERVER_PORT=$(rnd)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE="/tmp/metrictest$(rnd)"
"$binPath" -test.v -test.run=^TestIteration5$ \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server \
-server-port="$SERVER_PORT" \
-source-path=.
[ $? -eq 0 ] || exit 1

# inc6
SERVER_PORT=$(rnd)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE="/tmp/metrictest$(rnd)"
"$binPath" -test.v -test.run=^TestIteration6$ \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server \
-server-port="$SERVER_PORT" \
-source-path=.
[ $? -eq 0 ] || exit 1

# inc7
SERVER_PORT=$(rnd)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE="/tmp/metrictest$(rnd)"
"$binPath" -test.v -test.run=^TestIteration7$ \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server \
-server-port="$SERVER_PORT" \
-source-path=.
[ $? -eq 0 ] || exit 1

# inc8
SERVER_PORT=$(rnd)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE="/tmp/metrictest$(rnd)"
"$binPath" -test.v -test.run=^TestIteration8$ \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server \
-server-port="$SERVER_PORT" \
-source-path=.
[ $? -eq 0 ] || exit 1

# inc9
SERVER_PORT=$(rnd)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE="/tmp/metrictest$(rnd)"
"$binPath" -test.v -test.run=^TestIteration9$ \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server \
-file-storage-path="$TEMP_FILE" \
-server-port="$SERVER_PORT" \
-source-path=.
[ $? -eq 0 ] || exit 1

#export LOG_LEVEL=debug
# inc10
SERVER_PORT=$(rnd)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE="/tmp/metrictest$(rnd)"
"$binPath" -test.v -test.run=^TestIteration10[AB]$ \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server \
-database-dsn='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
-server-port="$SERVER_PORT" \
-source-path=.
[ $? -eq 0 ] || exit 1

# inc11
SERVER_PORT=$(rnd)
ADDRESS="localhost:${SERVER_PORT}"
TEMP_FILE="/tmp/metrictest$(rnd)"
"$binPath" -test.v -test.run=^TestIteration11$ \
-agent-binary-path=cmd/agent/agent \
-binary-path=cmd/server/server \
-database-dsn='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
-server-port="$SERVER_PORT" \
-source-path=.
[ $? -eq 0 ] || exit 1
#
## inc12
#SERVER_PORT=$(rnd)
#ADDRESS="localhost:${SERVER_PORT}"
#TEMP_FILE="/tmp/metrictest$(rnd)"
#"$binPath" -test.v -test.run=^TestIteration12$ \
#-agent-binary-path=cmd/agent/agent \
#-binary-path=cmd/server/server \
#-database-dsn='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
#-server-port="$SERVER_PORT" \
#-source-path=.
#[ $? -eq 0 ] || exit 1
#
## inc13
#SERVER_PORT=$(rnd)
#ADDRESS="localhost:${SERVER_PORT}"
#TEMP_FILE="/tmp/metrictest$(rnd)"
#"$binPath" -test.v -test.run=^TestIteration13$ \
#-agent-binary-path=cmd/agent/agent \
#-binary-path=cmd/server/server \
#-database-dsn='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
#-server-port="$SERVER_PORT" \
#-source-path=.
#[ $? -eq 0 ] || exit 1
#
## inc14
#SERVER_PORT=$(rnd)
#ADDRESS="localhost:${SERVER_PORT}"
#TEMP_FILE="/tmp/metrictest$(rnd)"
#"$binPath" -test.v -test.run=^TestIteration14$ \
#-agent-binary-path=cmd/agent/agent \
#-binary-path=cmd/server/server \
#-database-dsn='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
#-key="${TEMP_FILE}" \
#-server-port="$SERVER_PORT" \
#-source-path=.
#[ $? -eq 0 ] || exit 1
#
## inc14 race
#go test -v -race ./...
