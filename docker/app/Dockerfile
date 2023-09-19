FROM golang:1.21

RUN go install github.com/cespare/reflex@latest
COPY reflex.conf /

COPY start.sh /

ENTRYPOINT ["reflex", "-c", "/reflex.conf"]
