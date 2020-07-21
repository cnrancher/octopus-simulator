FROM --platform=$TARGETPLATFORM scratch

# NB(thxCode): automatic platform ARGs, ref to:
# - https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /
# -- for modbus
EXPOSE 5020
# -- for opcua
EXPOSE 4840
# -- for mqtt
EXPOSE 1883
COPY bin/simulator_${TARGETOS}_${TARGETARCH} /octopus
ENTRYPOINT ["/octopus"]
