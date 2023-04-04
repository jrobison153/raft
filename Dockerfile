FROM golang:alpine3.17

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

copy main.go ./
COPY api ./api/
COPY bootstrap ./bootstrap/
COPY crypto ./crypto/
COPY environment ./environment/
COPY journal ./journal/
COPY policy ./policy/
COPY replication ./replication/
COPY server ./server/
COPY state ./state/
COPY telemetry ./telemetry/

run go build -o raft

CMD [ "./raft" ]