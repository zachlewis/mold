FROM alpine

RUN echo '--------------------'

ADD ./mold /usr/bin/

CMD [ "/usr/bin/mold.exe" ]
