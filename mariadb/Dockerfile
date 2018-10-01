FROM mariadb:10.3

LABEL maintainer="kwf2030 <kwf2030@163.com>" \
      version=10.3

COPY my.cnf /etc/mysql/

COPY mariadb.cnf /etc/mysql/

COPY hiprice.sql /docker-entrypoint-initdb.d/