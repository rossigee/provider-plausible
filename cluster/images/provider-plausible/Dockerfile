FROM gcr.io/distroless/static:nonroot

ARG TARGETOS
ARG TARGETARCH

ADD bin/${TARGETOS}_${TARGETARCH}/provider /usr/local/bin/crossplane-plausible-provider

USER 65532
ENTRYPOINT ["crossplane-plausible-provider"]