# python:3.10.5-alpine3.16 (linux/amd64)
FROM python@sha256:52ce18e9d7a2556a3632d093f8f77700307735b7e7049dce3339c9bf9221ae7f

EXPOSE 8000

WORKDIR /app

ADD pkg/server/api/v1/ dev/api/ /app/

ENTRYPOINT [ "python" ]

CMD ["-m", "http.server", "8000"]
