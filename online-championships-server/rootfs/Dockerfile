FROM registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.22.0 AS builder

ENV GO111MODULE on
ENV CGO_ENABLED 1

WORKDIR /backend
COPY . .

RUN go build --trimpath --buildmode=plugin -o ./backend.so

FROM registry.heroiclabs.com/heroiclabs/nakama:3.22.0

COPY --from=builder /backend/backend.so /nakama/data/modules
COPY --from=builder /backend/nakama.yaml /nakama/data/nakama.yaml
