# python:3.11.0b4-alpine3.15 (linux/amd64)
FROM python@sha256:70825f4d36b31382c5b06e5c0d37de0c0bde49a42f26e77e6b3479d37662eb95

EXPOSE 8000

WORKDIR /app

ADD doc/api/index.html /app/

ENTRYPOINT [ "python" ]

CMD ["-m", "http.server", "8000"]
