FROM gcr.io/distroless/static:nonroot

# OpenContainers Image Spec Labels
LABEL org.opencontainers.image.title="provider-plausible"
LABEL org.opencontainers.image.description="Crossplane provider for Plausible Analytics sites and goals management"
LABEL org.opencontainers.image.vendor="rossigee"
LABEL org.opencontainers.image.licenses="Apache-2.0"
LABEL org.opencontainers.image.source="https://github.com/rossigee/provider-plausible"
LABEL org.opencontainers.image.url="https://github.com/rossigee/provider-plausible"
LABEL org.opencontainers.image.documentation="https://github.com/rossigee/provider-plausible/blob/master/README.md"

ARG TARGETOS
ARG TARGETARCH

ADD bin/${TARGETOS}_${TARGETARCH}/provider /usr/local/bin/crossplane-plausible-provider

USER 65532
ENTRYPOINT ["crossplane-plausible-provider"]