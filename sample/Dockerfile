FROM reg.gitlab.com/act3-ai/asce/data/tool/ace-dt as acedt

FROM docker.io/library/ubuntu as base

COPY --from=acedt /ko-app/ace-dt /usr/local/bin/ace-dt

RUN ace-dt version