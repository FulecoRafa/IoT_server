FROM golang

WORKDIR /app

COPY ./mqtt-redirect .

COPY ./redirect_config.json .

CMD ["./mqtt-redirect", "-config", "redirect_config.json"]
