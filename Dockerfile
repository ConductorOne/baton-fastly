FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-fastly"]
COPY baton-fastly /