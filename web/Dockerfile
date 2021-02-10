FROM alpine

COPY ./web_linux /home/app/

WORKDIR /home/app

EXPOSE 8888

ENTRYPOINT ./web_linux $0 $@
