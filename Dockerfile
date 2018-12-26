FROM golang AS build

WORKDIR /go/src/github.com/leancloud/lean-cli

COPY . .

RUN make binaries

FROM debian

LABEL version="0.20.0"
LABEL repository="http://github.com/leancloud/lean-cli"
LABEL homepage="http://github.com/leancloud/lean-cli"
LABEL maintainer="LeanCloud <support@leancloud.rocks>"

LABEL com.github.actions.name="GitHub Action for lean-cli"
LABEL com.github.actions.description="Use the lean-cli to deploy to LeanEngine."
LABEL com.github.actions.icon="cpu"
LABEL com.github.actions.color="blue"

COPY --from=build /go/src/github.com/leancloud/lean-cli/_build/lean-linux-x64 /usr/bin/lean

ENTRYPOINT ["lean"]
CMD ["deploy", "--github-action"]
