FROM alpine
LABEL Name="autoheal" 
COPY ./autoheal /usr/local/bin/
CMD [ "autoheal" ]
EXPOSE 9999
