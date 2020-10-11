FROM golang
LABEL Name="autoheal" 
RUN go build main.go
COPY ./autoheal /usr/local/bin/
CMD [ "autoheal" ]
EXPOSE 9999
