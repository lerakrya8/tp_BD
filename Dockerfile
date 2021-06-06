FROM golang:1.15 AS build

ADD . /app
WORKDIR /app
RUN go build ./cmd/main/main.go

FROM ubuntu:20.04

RUN apt-get -y update && apt-get install -y tzdata
ENV TZ=Russia/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

ENV PGVER 12
RUN apt-get -y update && apt-get install -y postgresql-$PGVER
USER postgres


RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER lera_bd WITH SUPERUSER PASSWORD 'password';" &&\
    createdb -O lera_bd db &&\
    /etc/init.d/postgresql stop


EXPOSE 5432
VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]
USER root

WORKDIR /usr/src/app
COPY . .
COPY --from=build /app/main/ .

EXPOSE 5000
ENV PGPASSWORD password
CMD service postgresql start && psql -h localhost -d db -U lera_bd -p 5432 -a -q -f ./sql/init.sql && ./main